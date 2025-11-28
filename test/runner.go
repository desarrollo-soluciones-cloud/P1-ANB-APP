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

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/videos/upload", apiBase),
		&buf,
	)
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

	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err == nil {
		if v, ok := obj["id"]; ok {
			return fmt.Sprint(v), body, nil
		}
		if v, ok := obj["video_id"]; ok {
			return fmt.Sprint(v), body, nil
		}
	}

	return "", body, fmt.Errorf("no id/video_id en respuesta upload")
}

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

						if v, ok := it["id"].(float64); ok {
							resp.Body.Close()
							return fmt.Sprintf("%.0f", v), nil
						}
						if v, ok := it["video_id"].(float64); ok {
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
					}
				}
			}
			resp.Body.Close()
		}

		select {
		case <-ctx.Done():
			return "", fmt.Errorf("cancelado esperando título=%s", upTitle)
		case <-time.After(2 * time.Second):
		}
	}

	return "", fmt.Errorf("timeout esperando título=%s", upTitle)
}

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
		return "", fmt.Errorf("GET /videos → HTTP %d: %s", resp.StatusCode, string(b))
	}

	var items []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return "", err
	}

	for _, it := range items {
		if v, ok := it["id"]; ok {
			return fmt.Sprint(v), nil
		}
		if v, ok := it["video_id"]; ok {
			return fmt.Sprint(v), nil
		}
	}

	return "", fmt.Errorf("lista /videos vacía")
}

func main() {

	apiBase := flag.String("api_base",
		"http://anb-backend-alb-713812655.us-east-1.elb.amazonaws.com/api/v1",
		"Base API")

	email := flag.String("email", "carlos@anb.com", "Email")
	password := flag.String("password", "password", "Password")

	videoPath := flag.String("video_path", "", "Ruta .mp4 real")
	loadtestPath := flag.String("loadtest_path", "./loadtest.go", "Ruta loadtest.go")

	runEsc2 := flag.Bool("run_esc2", false, "Ejecutar escenario 2")
	insecure := flag.Bool("insecure", false, "Pasar -insecure")
	outDirFlag := flag.String("out_dir", "", "Carpeta de salida")

	flag.Parse()

	if _, err := os.Stat(*loadtestPath); err != nil {
		fmt.Println("ERROR: no existe loadtest.go")
		os.Exit(1)
	}

	if *videoPath == "" {
		fmt.Println("-video_path obligatorio")
		os.Exit(1)
	}

	st, errStat := os.Stat(*videoPath)
	if errStat != nil || st.IsDir() {
		fmt.Println("video_path no válido")
		os.Exit(1)
	}

	outDir := *outDirFlag
	if outDir == "" {
		outDir = "resultados-" + nowStamp()
	}
	must(os.MkdirAll(outDir, 0755))

	fmt.Println("Resultados en:", outDir)

	ctx := context.Background()
	var lr loginResp

	err := httpJSON(ctx, "POST", *apiBase+"/auth/login",
		map[string]string{"email": *email, "password": *password},
		nil,
		&lr)

	if err != nil || lr.AccessToken == "" {
		fmt.Println("Login falló:", err)
		os.Exit(1)
	}

	fmt.Println("Login OK")

	headersAuth := filepath.Join(outDir, "headers_auth.txt")
	writeFile(headersAuth, []byte("Authorization: Bearer "+lr.AccessToken))

	headersJSON := filepath.Join(outDir, "headers_json.txt")
	writeFile(headersJSON, []byte("Content-Type: application/json"))

	loginBody := filepath.Join(outDir, "body_login.json")
	bodyStr := fmt.Sprintf(`{"email":"%s","password":"%s"}`, *email, *password)
	writeFile(loginBody, []byte(bodyStr))

	withCommon := func(a ...string) []string {
		if *insecure {
			a = append(a, "-insecure")
		}
		return a
	}

	// -------------------------------
	//  ESCENARIO 1 (AUMENTADO)
	// -------------------------------

	// PUBLIC
	err = goRunLoadtest(*loadtestPath, withCommon(
		"-url", *apiBase+"/public/videos",
		"-method", "GET",
		"-concurrency", "100", // AUMENTO
		"-rate", "150", // AUMENTO
		"-duration", "7m", // AUMENTO
		"-out_json", filepath.Join(outDir, "esc1_public.json"),
		"-out_csv", filepath.Join(outDir, "esc1_public.csv"),
	), "ESC1 PUBLIC")

	if err != nil {
		fmt.Println("WARN:", err)
	}

	// PRIVATE
	err = goRunLoadtest(*loadtestPath, withCommon(
		"-url", *apiBase+"/videos",
		"-method", "GET",
		"-headers", headersAuth,
		"-concurrency", "140", // AUMENTO
		"-rate", "200", // AUMENTO
		"-duration", "7m", // AUMENTO
		"-out_json", filepath.Join(outDir, "esc1_private.json"),
		"-out_csv", filepath.Join(outDir, "esc1_private.csv"),
	), "ESC1 PRIVATE")

	if err != nil {
		fmt.Println("WARN:", err)
	}

	// -------------------------------
	//  ESCENARIO 2 (AUMENTADO)
	// -------------------------------

	if *runEsc2 {

		err = goRunLoadtest(*loadtestPath, withCommon(
			"-url", *apiBase+"/public/videos",
			"-method", "GET",
			"-concurrency", "220", // AUMENTO
			"-rate", "420", // AUMENTO
			"-duration", "8m", // AUMENTO
			"-out_json", filepath.Join(outDir, "esc2_public.json"),
			"-out_csv", filepath.Join(outDir, "esc2_public.csv"),
		), "ESC2 PUBLIC")
		if err != nil {
			fmt.Println("WARN:", err)
		}

		err = goRunLoadtest(*loadtestPath, withCommon(
			"-url", *apiBase+"/videos",
			"-method", "GET",
			"-headers", headersAuth,
			"-concurrency", "200", // AUMENTO
			"-rate", "350", // AUMENTO
			"-duration", "8m", // AUMENTO
			"-out_json", filepath.Join(outDir, "esc2_private.json"),
			"-out_csv", filepath.Join(outDir, "esc2_private.csv"),
		), "ESC2 PRIVATE")
		if err != nil {
			fmt.Println("WARN:", err)
		}

		// BURST FINAL
		err = goRunLoadtest(*loadtestPath, withCommon(
			"-url", *apiBase+"/public/videos",
			"-method", "GET",
			"-concurrency", "260", // aumento fuerte
			"-rate", "520",
			"-duration", "90s",
			"-out_json", filepath.Join(outDir, "esc2_public_burst.json"),
			"-out_csv", filepath.Join(outDir, "esc2_public_burst.csv"),
		), "ESC2 PUBLIC BURST")

		err = goRunLoadtest(*loadtestPath, withCommon(
			"-url", *apiBase+"/videos",
			"-method", "GET",
			"-headers", headersAuth,
			"-concurrency", "240",
			"-rate", "480",
			"-duration", "90s",
			"-out_json", filepath.Join(outDir, "esc2_private_burst.json"),
			"-out_csv", filepath.Join(outDir, "esc2_private_burst.csv"),
		), "ESC2 PRIVATE BURST")
	}

	// -------------------------------
	//   SUBIDA DE VIDEO REAL
	// -------------------------------

	title := "LoadTest-" + nowStamp()
	videoID, uploadRaw, err := uploadVideo(ctx, *apiBase, lr.AccessToken, *videoPath, title)
	writeFile(filepath.Join(outDir, "upload_resp.json"), uploadRaw)

	if err != nil || videoID == "" {

		fmt.Println("Upload sin id, intentando polling...")

		ctxPoll, cancel := context.WithTimeout(ctx, 90*time.Second)
		defer cancel()

		if id2, err2 := waitVideoIDByTitle(ctxPoll, *apiBase, lr.AccessToken, title, 90*time.Second); err2 == nil {
			videoID = id2
		} else if id3, err3 := getAnyVideoID(ctx, *apiBase, lr.AccessToken); err3 == nil {
			videoID = id3
		} else {
			fmt.Println("No video_id. Termina pruebas sin por-id.")
			return
		}
	} else {
		fmt.Println("Upload OK, video_id =", videoID)
	}

	// -------------------------------
	//   PRUEBAS POR ID (NO aumentadas)
	// -------------------------------

	goRunLoadtest(*loadtestPath, withCommon(
		"-url", *apiBase+"/videos/"+videoID,
		"-method", "GET",
		"-headers", headersAuth,
		"-concurrency", "10",
		"-rate", "20",
		"-duration", "1m",
	), "GET /videos/:id")

	goRunLoadtest(*loadtestPath, withCommon(
		"-url", *apiBase+"/videos/"+videoID+"/download",
		"-method", "GET",
		"-headers", headersAuth,
		"-concurrency", "10",
		"-rate", "20",
		"-duration", "1m",
	), "GET /videos/:id/download")

	goRunLoadtest(*loadtestPath, withCommon(
		"-url", *apiBase+"/videos/"+videoID+"/mark-processed",
		"-method", "POST",
		"-headers", headersAuth,
		"-concurrency", "5",
		"-rate", "5",
		"-duration", "30s",
	), "POST mark-processed")

	goRunLoadtest(*loadtestPath, withCommon(
		"-url", *apiBase+"/public/videos/"+videoID+"/vote",
		"-method", "POST",
		"-headers", headersAuth,
		"-concurrency", "5",
		"-rate", "5",
		"-duration", "30s",
	), "POST vote")

	goRunLoadtest(*loadtestPath, withCommon(
		"-url", *apiBase+"/public/videos/"+videoID+"/vote",
		"-method", "DELETE",
		"-headers", headersAuth,
		"-concurrency", "5",
		"-rate", "5",
		"-duration", "30s",
	), "DELETE vote")

	fmt.Println("\nPruebas completadas. Carpeta:", outDir)
}
