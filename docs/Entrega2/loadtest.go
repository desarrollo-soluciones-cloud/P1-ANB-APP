package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Result struct {
	Start      time.Time
	Latency    time.Duration
	StatusCode int
	Err        string
	Bytes      int64
}

type Summary struct {
	Target             string        `json:"target"`
	Method             string        `json:"method"`
	Duration           time.Duration `json:"duration"`
	PlannedRateRPS     float64       `json:"planned_rate_rps"`
	PlannedConcurrency int           `json:"planned_concurrency"`
	TotalRequests      int64         `json:"total_requests"`
	SuccessRequests    int64         `json:"success_requests"`
	ErrorRequests      int64         `json:"error_requests"`
	SuccessRate        float64       `json:"success_rate"`
	ThroughputRPS      float64       `json:"throughput_rps"`
	LatencyMeanMs      float64       `json:"latency_mean_ms"`
	P50Ms              float64       `json:"p50_ms"`
	P90Ms              float64       `json:"p90_ms"`
	P95Ms              float64       `json:"p95_ms"`
	P99Ms              float64       `json:"p99_ms"`
	LatencyMaxMs       float64       `json:"latency_max_ms"`
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 100 {
		return sorted[len(sorted)-1]
	}
	pos := (p / 100.0) * (float64(len(sorted)) - 1)
	l := int(math.Floor(pos))
	u := int(math.Ceil(pos))
	if l == u {
		return sorted[l]
	}
	return sorted[l] + (sorted[u]-sorted[l])*(pos-float64(l))
}

func buildClient(timeout time.Duration, insecure bool) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        10000,
		MaxIdleConnsPerHost: 10000,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		Proxy:               http.ProxyFromEnvironment,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: insecure}, //nolint:gosec
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

func loadHeaders(path string) (http.Header, error) {
	h := http.Header{}
	if path == "" {
		return h, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("header inválido: %s", line)
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		h.Add(key, val)
	}
	return h, sc.Err()
}

func main() {
	// Flags
	url := flag.String("url", "", "Endpoint objetivo (ej: https://api.tuapp.com/upload)")
	method := flag.String("method", "GET", "Método HTTP (GET, POST, etc.)")
	bodyFile := flag.String("body", "", "Ruta a archivo con cuerpo de la petición (opcional)")
	headersFile := flag.String("headers", "", "Ruta a archivo de headers 'Clave: Valor' por línea (opcional)")
	concurrency := flag.Int("concurrency", 10, "Usuarios concurrentes (workers)")
	rate := flag.Float64("rate", 50, "RPS objetivo (peticiones por segundo)")
	duration := flag.Duration("duration", 30*time.Second, "Duración de la prueba (ej: 30s, 2m)")
	reqTimeout := flag.Duration("timeout", 15*time.Second, "Timeout por solicitud")
	insecureTLS := flag.Bool("insecure", false, "Permitir TLS inseguro (self-signed)")
	outCSV := flag.String("out_csv", "", "Ruta para CSV de resultados por petición (opcional)")
	outJSON := flag.String("out_json", "", "Ruta para JSON con resumen (opcional)")
	flag.Parse()

	if *url == "" {
		fmt.Println("Uso: go run loadtest.go -url https://host/endpoint -method POST -body data.json -headers headers.txt -concurrency 20 -rate 50 -duration 60s")
		os.Exit(1)
	}

	// Cargar body y headers
	var bodyBytes []byte
	var err error
	if *bodyFile != "" {
		bodyBytes, err = ioutil.ReadFile(*bodyFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error leyendo body: %v\n", err)
			os.Exit(1)
		}
	}
	headers, err := loadHeaders(*headersFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error leyendo headers: %v\n", err)
		os.Exit(1)
	}

	// Contexto de prueba
	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	client := buildClient(*reqTimeout, *insecureTLS)

	// Rate limiter por ticker
	interval := time.Duration(float64(time.Second) / *rate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Canales y sincronización
	type job struct{}
	jobs := make(chan job, 1024)
	results := make(chan Result, 100000)
	var wg sync.WaitGroup

	// Workers
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				start := time.Now()
				reqBodyReader := io.Reader(nil)
				if len(bodyBytes) > 0 {
					reqBodyReader = ioutil.NopCloser(strings.NewReader(string(bodyBytes)))
				}
				req, err := http.NewRequestWithContext(ctx, *method, *url, reqBodyReader)
				if err != nil {
					results <- Result{Start: start, Latency: time.Since(start), StatusCode: 0, Err: err.Error()}
					continue
				}
				// Si hay body y no se clonó, recrear reader:
				if len(bodyBytes) > 0 {
					req.Body = ioutil.NopCloser(strings.NewReader(string(bodyBytes)))
				}
				for k, vs := range headers {
					for _, v := range vs {
						req.Header.Add(k, v)
					}
				}
				resp, err := client.Do(req)
				if err != nil {
					results <- Result{Start: start, Latency: time.Since(start), StatusCode: 0, Err: err.Error()}
					continue
				}
				b, _ := io.Copy(ioutil.Discard, resp.Body)
				resp.Body.Close()
				results <- Result{
					Start:      start,
					Latency:    time.Since(start),
					StatusCode: resp.StatusCode,
					Err:        "",
					Bytes:      b,
				}
			}
		}()
	}

	// Productor: emite "jobs" a ritmo fijo hasta que termine el tiempo
	var totalEmitted int64
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(jobs)
				return
			case <-ticker.C:
				atomic.AddInt64(&totalEmitted, 1)
				select {
				case jobs <- job{}:
				default:
					// Si el buffer de jobs está lleno, descartamos para respetar ritmo
				}
			}
		}
	}()

	// Recolector de resultados
	var collectWg sync.WaitGroup
	collectWg.Add(1)
	collected := make([]Result, 0, 100000)
	go func() {
		defer collectWg.Done()
		for {
			select {
			case r, ok := <-results:
				if !ok {
					return
				}
				collected = append(collected, r)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Esperar fin de workers
	wg.Wait()
	close(results)
	collectWg.Wait()

	// Métricas
	var success, errors int64
	latencies := make([]float64, 0, len(collected))
	var maxLat float64
	for _, r := range collected {
		if r.Err != "" || r.StatusCode >= 500 || r.StatusCode == 0 {
			errors++
		} else {
			success++
		}
		ms := float64(r.Latency.Microseconds()) / 1000.0
		latencies = append(latencies, ms)
		if ms > maxLat {
			maxLat = ms
		}
	}
	total := int64(len(collected))
	elapsed := *duration
	if total == 0 {
		fmt.Println("No se recolectaron resultados. Revisa conectividad/URL/headers.")
		os.Exit(2)
	}
	// ordenar latencias
	sortFloat64s(latencies)

	mean := meanFloat64(latencies)
	p50 := percentile(latencies, 50)
	p90 := percentile(latencies, 90)
	p95 := percentile(latencies, 95)
	p99 := percentile(latencies, 99)
	throughput := float64(total) / elapsed.Seconds()
	successRate := 100.0 * float64(success) / float64(total)

	sum := Summary{
		Target:             *url,
		Method:             *method,
		Duration:           *duration,
		PlannedRateRPS:     *rate,
		PlannedConcurrency: *concurrency,
		TotalRequests:      total,
		SuccessRequests:    success,
		ErrorRequests:      errors,
		SuccessRate:        successRate,
		ThroughputRPS:      throughput,
		LatencyMeanMs:      mean,
		P50Ms:              p50,
		P90Ms:              p90,
		P95Ms:              p95,
		P99Ms:              p99,
		LatencyMaxMs:       maxLat,
	}

	// Print resumen amigable
	fmt.Printf("\n=== RESUMEN ===\n")
	fmt.Printf("Target: %s %s\n", *method, *url)
	fmt.Printf("Duración: %v | Concurrency: %d | Rate planificado: %.1f rps\n", *duration, *concurrency, *rate)
	fmt.Printf("Total: %d | Éxitos: %d | Errores: %d | Éxito: %.2f%%\n", total, success, errors, successRate)
	fmt.Printf("Throughput real: %.2f rps\n", throughput)
	fmt.Printf("Latencia (ms) -> mean: %.2f | p50: %.2f | p90: %.2f | p95: %.2f | p99: %.2f | max: %.2f\n",
		mean, p50, p90, p95, p99, maxLat)

	// CSV opcional
	if *outCSV != "" {
		f, err := os.Create(*outCSV)
		if err == nil {
			w := csv.NewWriter(f)
			_ = w.Write([]string{"start_iso", "latency_ms", "status_code", "error", "bytes"})
			for _, r := range collected {
				_ = w.Write([]string{
					r.Start.Format(time.RFC3339Nano),
					fmt.Sprintf("%.3f", float64(r.Latency.Microseconds())/1000.0),
					fmt.Sprintf("%d", r.StatusCode),
					r.Err,
					fmt.Sprintf("%d", r.Bytes),
				})
			}
			w.Flush()
			f.Close()
			fmt.Printf("CSV guardado en: %s\n", *outCSV)
		} else {
			fmt.Fprintf(os.Stderr, "No se pudo escribir CSV: %v\n", err)
		}
	}

	// JSON opcional
	if *outJSON != "" {
		b, _ := json.MarshalIndent(sum, "", "  ")
		_ = os.WriteFile(*outJSON, b, 0644)
		fmt.Printf("JSON guardado en: %s\n", *outJSON)
	}
}

// utilidades simples sin dependencias externas
func sortFloat64s(a []float64) {
	quickSort(a, 0, len(a)-1)
}
func quickSort(a []float64, lo, hi int) {
	if lo >= hi {
		return
	}
	p := partition(a, lo, hi)
	quickSort(a, lo, p-1)
	quickSort(a, p+1, hi)
}
func partition(a []float64, lo, hi int) int {
	pivot := a[hi]
	i := lo
	for j := lo; j < hi; j++ {
		if a[j] < pivot {
			a[i], a[j] = a[j], a[i]
			i++
		}
	}
	a[i], a[hi] = a[hi], a[i]
	return i
}
func meanFloat64(a []float64) float64 {
	if len(a) == 0 {
		return 0
	}
	var s float64
	for _, v := range a {
		s += v
	}
	return s / float64(len(a))
}
