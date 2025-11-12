package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type loginResp struct {
	AccessToken string `json:"access_token"`
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func writeFile(path string, data []byte) {
	must(os.WriteFile(path, data, 0644))
}

func nowStamp() string {
	return time.Now().Format("20060102-150405")
}

func goRunLoadtest(loadtestPath string, args []string, label string) error {
	fmt.Printf(">> [%s] go run %s %s\n", label, loadtestPath, strings.Join(args, " "))
	cmd := exec.Command("go", append([]string{"run", loadtestPath}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func httpJSON(ctx context.Context, method, url string, body any, headers map[string]string, v any) error {
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, r)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v2 := range headers {
		req.Header.Set(k, v2)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}
	if v != nil {
		if err := json.Unmarshal(b, v); err != nil {
			return fmt.Errorf("unmarshal: %w (%s)", err, string(b))
		}
	}
	return nil
}

func uploadVideo(ctx context.Context, apiBase, token, videoPath, title string) (string, []byte, error) {
	f, err := os.Open(videoPath)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("video", filepath.Base(videoPath))
	if err != nil {
		return "", nil, err
	}
	if _, err := io.Copy(fw, f); err != nil {
		return "", nil, err
	}
	_ = w.WriteField("title", title)
	_ = w.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/videos/upload", apiBase), &buf)
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", body, fmt.Errorf("upload HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Intentamos id o video_id (algunos backends no lo devuelven aquí)
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err == nil {
		if v, ok := obj["id"]; ok {
			return fmt.Sprint(v), body, nil
		}
		if v, ok := obj["video_id"]; ok {
			return fmt.Sprint(v), body, nil
		}
	}
	return "", body, fmt.Errorf("no se encontró id/video_id en la respuesta")
}

// Polling por título hasta que aparezca en GET /videos
func waitVideoIDByTitle(ctx context.Context, apiBase, token, upTitle string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 10 * time.Second}

	for time.Now().Before(deadline) {
		req, _ := http.NewRequestWithContext(ctx, "GET", apiBase+"/videos", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			var items []map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&items); err == nil {
				for _, it := range items {
					title, _ := it["title"].(string)
					if title == upTitle {
						// intenta varias claves y tipos
						if v, ok := it["id"].(float64); ok {
							resp.Body.Close()
							return fmt.Sprintf("%.0f", v), nil
						}
						if v, ok := it["video_id"].(float64); ok {
							resp.Body.Close()
							return fmt.Sprintf("%.0f", v), nil
						}
						if v, ok := it["videoId"].(float64); ok {
							resp.Body.Close()
							return fmt.Sprintf("%.0f", v), nil
						}
						if v, ok := it["id"].(string); ok {
							resp.Body.Close()
							return v, nil
						}
						if v, ok := it["video_id"].(string); ok {
							resp.Body.Close()
							return v, nil
						}
						if v, ok := it["videoId"].(string); ok {
							resp.Body.Close()
							return v, nil
						}
					}
				}
			}
			resp.Body.Close()
		}
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("cancelado esperando video con title=%s", upTitle)
		case <-time.After(2 * time.Second): // reintento cada 2s
		}
	}
	return "", fmt.Errorf("no apareció video con title=%s en %s", upTitle, timeout)
}

// Fallback: tomar cualquier video propio si existe (para no bloquear por pruebas por-id)
func getAnyVideoID(ctx context.Context, apiBase, token string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", apiBase+"/videos", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GET /videos -> HTTP %d: %s", resp.StatusCode, string(b))
	}
	var items []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return "", err
	}
	for _, it := range items {
		// devuelve la primera clave válida
		if v, ok := it["id"]; ok {
			return fmt.Sprint(v), nil
		}
		if v, ok := it["video_id"]; ok {
			return fmt.Sprint(v), nil
		}
		if v, ok := it["videoId"]; ok {
			return fmt.Sprint(v), nil
		}
	}
	return "", fmt.Errorf("no hay videos disponibles en /videos")
}

func main() {
	// Flags
	apiBase := flag.String("api_base", "http://34.207.169.60:9090/api/v1", "Base de la API")
	email := flag.String("email", "carlos@anb.com", "Email para login")
	password := flag.String("password", "password", "Password para login")
	videoPath := flag.String("video_path", "", "Ruta a un archivo de video existente (requerido)")
	loadtestPath := flag.String("loadtest_path", "./loadtest.go", "Ruta a loadtest.go")
	runEsc2 := flag.Bool("run_esc2", false, "Ejecutar también escenario 2")
	insecure := flag.Bool("insecure", false, "Pasar -insecure a loadtest.go")
	outDirFlag := flag.String("out_dir", "", "Carpeta de salida (opcional)")
	flag.Parse()

	// Validaciones
	if _, err := os.Stat(*loadtestPath); err != nil {
		fmt.Fprintf(os.Stderr, "No encuentro loadtest.go en: %s\n", *loadtestPath)
		os.Exit(1)
	}
	if *videoPath == "" {
		fmt.Fprintln(os.Stderr, "-video_path es obligatorio (ruta a un .mp4 real).")
		os.Exit(1)
	}
	if st, err := os.Stat(*videoPath); err != nil || st.IsDir() {
		fmt.Fprintf(os.Stderr, "El -video_path no existe o es carpeta: %s\n", *videoPath)
		os.Exit(1)
	}

	// Carpeta resultados
	outDir := *outDirFlag
	if outDir == "" {
		outDir = "resultados-" + nowStamp()
	}
	must(os.MkdirAll(outDir, 0755))
	fmt.Println("Resultados en:", outDir)

	// === Login para token ===
	ctx := context.Background()
	var lr loginResp
	err := httpJSON(ctx, "POST", *apiBase+"/auth/login",
		map[string]string{"email": *email, "password": *password},
		nil, &lr)
	if err != nil || lr.AccessToken == "" {
		fmt.Fprintf(os.Stderr, "Login falló: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Login OK →", *email)

	// Headers auxiliares (loadtest.go consume headers desde archivo)
	headersAuthPath := filepath.Join(outDir, "headers_auth.txt")
	headersJSONPath := filepath.Join(outDir, "headers_json.txt")
	writeFile(headersAuthPath, []byte("Authorization: Bearer "+lr.AccessToken))
	writeFile(headersJSONPath, []byte("Content-Type: application/json"))

	// Guarda body del login para prueba ligera de /auth/login
	loginBodyPath := filepath.Join(outDir, "body_login.json")
	writeFile(loginBodyPath, []byte(fmt.Sprintf(`{"email":"%s","password":"%s"}`, *email, *password)))

	// Helper para armar args comunes
	withCommon := func(a ...string) []string {
		if *insecure {
			a = append(a, "-insecure")
		}
		return a
	}

	// === [Ligero] POST /auth/login ===
	// AUMENTO: concurrency 30 (antes 25), rate 30 (antes 25), duration 60s (se mantiene fuerte)
	err = goRunLoadtest(*loadtestPath, withCommon(
		"-url", *apiBase+"/auth/login",
		"-method", "POST",
		"-headers", headersJSONPath,
		"-body", loginBodyPath,
		"-concurrency", "30", // AUMENTO
		"-rate", "30", // AUMENTO
		"-duration", "60s",
		"-out_json", filepath.Join(outDir, "login_load.json"),
		"-out_csv", filepath.Join(outDir, "login_load.csv"),
	), "AUTH login (ligero)")
	if err != nil {
		fmt.Println("WARN:", err)
	}

	// === Escenario 1 ===
	// GET /public/videos (moderado+)
	// AUMENTO: concurrency 60 (antes 30), rate 90 (antes 40), duration 5m (antes 3m)
	err = goRunLoadtest(*loadtestPath, withCommon(
		"-url", *apiBase+"/public/videos",
		"-method", "GET",
		"-concurrency", "60", // AUMENTO
		"-rate", "90", // AUMENTO
		"-duration", "5m", // AUMENTO
		"-out_json", filepath.Join(outDir, "public_videos_esc1.json"),
		"-out_csv", filepath.Join(outDir, "public_videos_esc1.csv"),
	), "PUBLIC /public/videos esc1")
	if err != nil {
		fmt.Println("WARN:", err)
	}

	// GET /videos (privado) (moderado+)
	// AUMENTO: concurrency 80 (antes 40), rate 110 (antes 60), duration 5m (antes 3m)
	err = goRunLoadtest(*loadtestPath, withCommon(
		"-url", *apiBase+"/videos",
		"-method", "GET",
		"-headers", headersAuthPath,
		"-concurrency", "80", // AUMENTO
		"-rate", "110", // AUMENTO
		"-duration", "5m", // AUMENTO
		"-out_json", filepath.Join(outDir, "videos_esc1.json"),
		"-out_csv", filepath.Join(outDir, "videos_esc1.csv"),
	), "PRIVATE /videos esc1")
	if err != nil {
		fmt.Println("WARN:", err)
	}

	// === Escenario 2 ===
	if *runEsc2 {
		// Plateau fuerte: PUBLIC @260 rps, conc 140, 7m
		// AUMENTO sustancial vs entrega previa (120 rps, conc 60, 3m)
		err = goRunLoadtest(*loadtestPath, withCommon(
			"-url", *apiBase+"/public/videos",
			"-method", "GET",
			"-concurrency", "140", // AUMENTO
			"-rate", "260", // AUMENTO
			"-duration", "7m", // AUMENTO
			"-out_json", filepath.Join(outDir, "public_videos_esc2.json"),
			"-out_csv", filepath.Join(outDir, "public_videos_esc2.csv"),
		), "PUBLIC /public/videos esc2 (plateau)")
		if err != nil {
			fmt.Println("WARN:", err)
		}

		// Plateau fuerte: PRIVATE @220 rps, conc 120, 7m
		// AUMENTO sustancial vs entrega previa (100 rps, conc 60, 3m)
		err = goRunLoadtest(*loadtestPath, withCommon(
			"-url", *apiBase+"/videos",
			"-method", "GET",
			"-headers", headersAuthPath,
			"-concurrency", "120", // AUMENTO
			"-rate", "220", // AUMENTO
			"-duration", "7m", // AUMENTO
			"-out_json", filepath.Join(outDir, "videos_esc2.json"),
			"-out_csv", filepath.Join(outDir, "videos_esc2.csv"),
		), "PRIVATE /videos esc2 (plateau)")
		if err != nil {
			fmt.Println("WARN:", err)
		}

		// Burst final: PUBLIC @320 rps, conc 160, 60s (misma Escenario 2, no es escenario nuevo)
		// AUMENTO: añade una ráfaga para detectar saturación/elasticidad
		err = goRunLoadtest(*loadtestPath, withCommon(
			"-url", *apiBase+"/public/videos",
			"-method", "GET",
			"-concurrency", "160", // AUMENTO
			"-rate", "320", // AUMENTO
			"-duration", "60s", // AUMENTO (nuevo burst)
			"-out_json", filepath.Join(outDir, "public_videos_esc2_burst.json"),
			"-out_csv", filepath.Join(outDir, "public_videos_esc2_burst.csv"),
		), "PUBLIC /public/videos esc2 (burst)")
		if err != nil {
			fmt.Println("WARN:", err)
		}

		// Burst final: PRIVATE @280 rps, conc 150, 60s
		// AUMENTO: ráfaga equivalente en el endpoint privado
		err = goRunLoadtest(*loadtestPath, withCommon(
			"-url", *apiBase+"/videos",
			"-method", "GET",
			"-headers", headersAuthPath,
			"-concurrency", "150", // AUMENTO
			"-rate", "280", // AUMENTO
			"-duration", "60s", // AUMENTO (nuevo burst)
			"-out_json", filepath.Join(outDir, "videos_esc2_burst.json"),
			"-out_csv", filepath.Join(outDir, "videos_esc2_burst.csv"),
		), "PRIVATE /videos esc2 (burst)")
		if err != nil {
			fmt.Println("WARN:", err)
		}
	}

	// === Upload con video real (OBLIGATORIO para por-id) ===
	title := "LoadTest-" + nowStamp()
	videoID, uploadRaw, err := uploadVideo(ctx, *apiBase, lr.AccessToken, *videoPath, title)
	writeFile(filepath.Join(outDir, "upload_resp.json"), uploadRaw)

	if err != nil || videoID == "" {
		fmt.Fprintf(os.Stderr, "Upload sin id directo: %v\n", err)
		fmt.Println("Intentando localizar el video por título (polling)…")

		// Intento 1: polling por título
		ctxPoll, cancel := context.WithTimeout(ctx, 90*time.Second)
		defer cancel()
		if id2, err2 := waitVideoIDByTitle(ctxPoll, *apiBase, lr.AccessToken, title, 90*time.Second); err2 == nil && id2 != "" {
			videoID = id2
			fmt.Println("OK: encontrado por título → video_id =", videoID)
		} else {
			fmt.Println("Polling no encontró el video por título:", err2)
			// Intento 2 (fallback): tomar cualquier video propio
			if id3, err3 := getAnyVideoID(ctx, *apiBase, lr.AccessToken); err3 == nil && id3 != "" {
				videoID = id3
				fmt.Println("Fallback: usando video_id existente →", videoID)
			} else {
				fmt.Println("No hay video_id disponible. Se omiten pruebas por-id.")
				fmt.Println("\n Pruebas completadas (sin por-id). Carpeta:", outDir)
				return
			}
		}
	} else {
		fmt.Println("Upload OK → video_id =", videoID)
	}

	// === Pruebas por-ID (ligeras; se mantienen iguales) ===
	// GET /videos/:id
	err = goRunLoadtest(*loadtestPath, withCommon(
		"-url", *apiBase+"/videos/"+videoID,
		"-method", "GET",
		"-headers", headersAuthPath,
		"-concurrency", "10",
		"-rate", "20",
		"-duration", "1m",
		"-out_json", filepath.Join(outDir, fmt.Sprintf("video_%s_get.json", videoID)),
		"-out_csv", filepath.Join(outDir, fmt.Sprintf("video_%s_get.csv", videoID)),
	), "PRIVATE /videos/:id")
	if err != nil {
		fmt.Println("WARN:", err)
	}

	// GET /videos/:id/download
	err = goRunLoadtest(*loadtestPath, withCommon(
		"-url", *apiBase+"/videos/"+videoID+"/download",
		"-method", "GET",
		"-headers", headersAuthPath,
		"-concurrency", "10",
		"-rate", "20",
		"-duration", "1m",
		"-out_json", filepath.Join(outDir, fmt.Sprintf("video_%s_download.json", videoID)),
		"-out_csv", filepath.Join(outDir, fmt.Sprintf("video_%s_download.csv", videoID)),
	), "PRIVATE /videos/:id/download")
	if err != nil {
		fmt.Println("WARN:", err)
	}

	// POST /videos/:id/mark-processed
	err = goRunLoadtest(*loadtestPath, withCommon(
		"-url", *apiBase+"/videos/"+videoID+"/mark-processed",
		"-method", "POST",
		"-headers", headersAuthPath,
		"-concurrency", "5",
		"-rate", "5",
		"-duration", "30s",
		"-out_json", filepath.Join(outDir, fmt.Sprintf("video_%s_mark.json", videoID)),
		"-out_csv", filepath.Join(outDir, fmt.Sprintf("video_%s_mark.csv", videoID)),
	), "PRIVATE /videos/:id/mark-processed")
	if err != nil {
		fmt.Println("WARN:", err)
	}

	// POST /public/videos/:id/vote
	err = goRunLoadtest(*loadtestPath, withCommon(
		"-url", *apiBase+"/public/videos/"+videoID+"/vote",
		"-method", "POST",
		"-headers", headersAuthPath,
		"-concurrency", "5",
		"-rate", "5",
		"-duration", "30s",
		"-out_json", filepath.Join(outDir, fmt.Sprintf("vote_%s_post.json", videoID)),
		"-out_csv", filepath.Join(outDir, fmt.Sprintf("vote_%s_post.csv", videoID)),
	), "VOTE POST /public/videos/:id/vote")
	if err != nil {
		fmt.Println("WARN:", err)
	}

	// DELETE /public/videos/:id/vote
	err = goRunLoadtest(*loadtestPath, withCommon(
		"-url", *apiBase+"/public/videos/"+videoID+"/vote",
		"-method", "DELETE",
		"-headers", headersAuthPath,
		"-concurrency", "5",
		"-rate", "5",
		"-duration", "30s",
		"-out_json", filepath.Join(outDir, fmt.Sprintf("vote_%s_delete.json", videoID)),
		"-out_csv", filepath.Join(outDir, fmt.Sprintf("vote_%s_delete.csv", videoID)),
	), "VOTE DELETE /public/videos/:id/vote")
	if err != nil {
		fmt.Println("WARN:", err)
	}

	fmt.Println("\n Pruebas completadas. Carpeta:", outDir)
}
