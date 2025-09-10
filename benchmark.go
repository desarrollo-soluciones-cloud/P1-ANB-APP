package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// Configuración del benchmark
type Config struct {
	BaseURL     string
	Concurrency int
	Requests    int
	Endpoints   []Endpoint
}

type Endpoint struct {
	Name    string
	Method  string
	Path    string
	Headers map[string]string
	Body    interface{}
}

// Métricas por endpoint
type EndpointMetrics struct {
	Name          string
	TotalRequests int
	SuccessCount  int
	ErrorCount    int
	ResponseTimes []time.Duration
	AvgResponse   time.Duration
	MinResponse   time.Duration
	MaxResponse   time.Duration
	P95Response   time.Duration
	P99Response   time.Duration
	Throughput    float64
	ErrorRate     float64
}

// Resultado de una request individual
type RequestResult struct {
	Endpoint     string
	Duration     time.Duration
	StatusCode   int
	Success      bool
	Error        error
	ResponseSize int64
}

func main() {
	config := Config{
		BaseURL:     getEnv("API_URL", "http://localhost:9090"),
		Concurrency: 10,
		Requests:    100,
		Endpoints: []Endpoint{
			{
				Name:   "Health Check",
				Method: "GET",
				Path:   "/health",
			},
			{
				Name:   "Public Videos",
				Method: "GET",
				Path:   "/api/v1/public/videos",
			},
			{
				Name:   "Public Rankings",
				Method: "GET",
				Path:   "/api/v1/public/rankings",
			},
			{
				Name:   "User Registration",
				Method: "POST",
				Path:   "/api/v1/auth/signup",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: map[string]interface{}{
					"first_name": "Test",
					"last_name":  "User",
					"email":      fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
					"password":   "password123",
					"password2":  "password123",
					"city":       "Bogotá",
					"country":    "Colombia",
				},
			},
		},
	}

	fmt.Printf("🚀 ANB API Performance Benchmark\n")
	fmt.Printf("📊 Base URL: %s\n", config.BaseURL)
	fmt.Printf("⚡ Concurrency: %d\n", config.Concurrency)
	fmt.Printf("📈 Requests per endpoint: %d\n", config.Requests)
	fmt.Printf("🎯 Endpoints: %d\n\n", len(config.Endpoints))

	// Ejecutar benchmark para cada endpoint
	var allMetrics []EndpointMetrics
	for _, endpoint := range config.Endpoints {
		fmt.Printf("Testing %s...\n", endpoint.Name)
		metrics := benchmarkEndpoint(config, endpoint)
		allMetrics = append(allMetrics, metrics)
		printEndpointMetrics(metrics)
		fmt.Println()
	}

	// Resumen general
	printSummary(allMetrics)
}

func benchmarkEndpoint(config Config, endpoint Endpoint) EndpointMetrics {
	results := make(chan RequestResult, config.Requests)
	var wg sync.WaitGroup

	startTime := time.Now()

	// Crear workers concurrentes
	requestsPerWorker := config.Requests / config.Concurrency
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				result := makeRequest(config.BaseURL, endpoint)
				results <- result
			}
		}()
	}

	// Manejar requests restantes
	remainder := config.Requests % config.Concurrency
	for i := 0; i < remainder; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := makeRequest(config.BaseURL, endpoint)
			results <- result
		}()
	}

	wg.Wait()
	close(results)

	totalDuration := time.Since(startTime)

	// Procesar resultados
	return processResults(endpoint.Name, results, totalDuration)
}

func makeRequest(baseURL string, endpoint Endpoint) RequestResult {
	var body io.Reader

	if endpoint.Body != nil {
		jsonBody, _ := json.Marshal(endpoint.Body)
		body = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(endpoint.Method, baseURL+endpoint.Path, body)
	if err != nil {
		return RequestResult{
			Endpoint: endpoint.Name,
			Success:  false,
			Error:    err,
		}
	}

	// Agregar headers
	for key, value := range endpoint.Headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return RequestResult{
			Endpoint: endpoint.Name,
			Duration: duration,
			Success:  false,
			Error:    err,
		}
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)

	success := resp.StatusCode >= 200 && resp.StatusCode < 400

	return RequestResult{
		Endpoint:     endpoint.Name,
		Duration:     duration,
		StatusCode:   resp.StatusCode,
		Success:      success,
		ResponseSize: int64(len(responseBody)),
	}
}

func processResults(endpointName string, results chan RequestResult, totalDuration time.Duration) EndpointMetrics {
	var responseTimes []time.Duration
	var successCount, errorCount int
	var totalResponseSize int64

	for result := range results {
		responseTimes = append(responseTimes, result.Duration)
		totalResponseSize += result.ResponseSize

		if result.Success {
			successCount++
		} else {
			errorCount++
		}
	}

	totalRequests := len(responseTimes)

	// Ordenar tiempos de respuesta para percentiles
	sort.Slice(responseTimes, func(i, j int) bool {
		return responseTimes[i] < responseTimes[j]
	})

	// Calcular métricas
	var avgResponse time.Duration
	if totalRequests > 0 {
		var total time.Duration
		for _, rt := range responseTimes {
			total += rt
		}
		avgResponse = total / time.Duration(totalRequests)
	}

	return EndpointMetrics{
		Name:          endpointName,
		TotalRequests: totalRequests,
		SuccessCount:  successCount,
		ErrorCount:    errorCount,
		ResponseTimes: responseTimes,
		AvgResponse:   avgResponse,
		MinResponse:   getMin(responseTimes),
		MaxResponse:   getMax(responseTimes),
		P95Response:   getPercentile(responseTimes, 95),
		P99Response:   getPercentile(responseTimes, 99),
		Throughput:    float64(totalRequests) / totalDuration.Seconds(),
		ErrorRate:     float64(errorCount) / float64(totalRequests) * 100,
	}
}

func printEndpointMetrics(metrics EndpointMetrics) {
	fmt.Printf("📋 %s Results:\n", metrics.Name)
	fmt.Printf("   Total Requests: %d\n", metrics.TotalRequests)
	fmt.Printf("   Success: %d | Errors: %d\n", metrics.SuccessCount, metrics.ErrorCount)
	fmt.Printf("   ⏱️  Response Times:\n")
	fmt.Printf("      Average: %v\n", metrics.AvgResponse)
	fmt.Printf("      Min: %v | Max: %v\n", metrics.MinResponse, metrics.MaxResponse)
	fmt.Printf("      P95: %v | P99: %v\n", metrics.P95Response, metrics.P99Response)
	fmt.Printf("   🚀 Throughput: %.2f req/sec\n", metrics.Throughput)
	fmt.Printf("   ❌ Error Rate: %.2f%%\n", metrics.ErrorRate)

	// Evaluación contra objetivos
	evaluateMetrics(metrics)
}

func evaluateMetrics(metrics EndpointMetrics) {
	fmt.Printf("   📊 Performance Evaluation:\n")

	// Tiempo de respuesta promedio
	if metrics.AvgResponse < 500*time.Millisecond {
		fmt.Printf("      ✅ Response Time: EXCELLENT (< 500ms objetivo interno)\n")
	} else if metrics.AvgResponse < 1000*time.Millisecond {
		fmt.Printf("      ✅ Response Time: GOOD (< 1000ms objetivo cliente)\n")
	} else {
		fmt.Printf("      ❌ Response Time: POOR (> 1000ms objetivo cliente)\n")
	}

	// Tasa de errores
	if metrics.ErrorRate < 1.0 {
		fmt.Printf("      ✅ Error Rate: EXCELLENT (< 1%% objetivo interno)\n")
	} else if metrics.ErrorRate < 5.0 {
		fmt.Printf("      ✅ Error Rate: GOOD (< 5%% objetivo cliente)\n")
	} else {
		fmt.Printf("      ❌ Error Rate: POOR (> 5%% objetivo cliente)\n")
	}

	// Throughput
	if metrics.Throughput > 100 {
		fmt.Printf("      ✅ Throughput: EXCELLENT (> 100 req/sec objetivo cliente)\n")
	} else if metrics.Throughput > 50 {
		fmt.Printf("      ✅ Throughput: GOOD (> 50 req/sec objetivo interno)\n")
	} else {
		fmt.Printf("      ❌ Throughput: POOR (< 50 req/sec objetivo interno)\n")
	}
}

func printSummary(allMetrics []EndpointMetrics) {
	fmt.Printf("🎯 PERFORMANCE SUMMARY\n")
	fmt.Printf("=" + strings.Repeat("=", 50) + "\n")

	var totalRequests, totalSuccess, totalErrors int
	var totalAvgResponseTime time.Duration
	var totalThroughput float64

	for _, metrics := range allMetrics {
		totalRequests += metrics.TotalRequests
		totalSuccess += metrics.SuccessCount
		totalErrors += metrics.ErrorCount
		totalAvgResponseTime += metrics.AvgResponse
		totalThroughput += metrics.Throughput
	}

	avgResponseTime := totalAvgResponseTime / time.Duration(len(allMetrics))
	avgThroughput := totalThroughput / float64(len(allMetrics))
	overallErrorRate := float64(totalErrors) / float64(totalRequests) * 100

	fmt.Printf("📊 Overall Metrics:\n")
	fmt.Printf("   Total Requests: %d\n", totalRequests)
	fmt.Printf("   Success Rate: %.2f%%\n", float64(totalSuccess)/float64(totalRequests)*100)
	fmt.Printf("   Average Response Time: %v\n", avgResponseTime)
	fmt.Printf("   Average Throughput: %.2f req/sec\n", avgThroughput)
	fmt.Printf("   Overall Error Rate: %.2f%%\n", overallErrorRate)

	fmt.Printf("\n🎯 Objectives Compliance:\n")

	// Evaluar objetivos generales
	if avgResponseTime < 500*time.Millisecond {
		fmt.Printf("   ✅ Response Time: Meets internal objective (< 500ms)\n")
	} else if avgResponseTime < 1000*time.Millisecond {
		fmt.Printf("   ✅ Response Time: Meets client objective (< 1000ms)\n")
	} else {
		fmt.Printf("   ❌ Response Time: Does not meet objectives\n")
	}

	if avgThroughput > 100 {
		fmt.Printf("   ✅ Throughput: Meets client objective (> 100 req/sec)\n")
	} else if avgThroughput > 50 {
		fmt.Printf("   ✅ Throughput: Meets internal objective (> 50 req/sec)\n")
	} else {
		fmt.Printf("   ❌ Throughput: Does not meet objectives\n")
	}

	if overallErrorRate < 1.0 {
		fmt.Printf("   ✅ Error Rate: Meets internal objective (< 1%%)\n")
	} else if overallErrorRate < 5.0 {
		fmt.Printf("   ✅ Error Rate: Meets client objective (< 5%%)\n")
	} else {
		fmt.Printf("   ❌ Error Rate: Does not meet objectives\n")
	}
}

// Funciones utilitarias
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getMin(times []time.Duration) time.Duration {
	if len(times) == 0 {
		return 0
	}
	return times[0]
}

func getMax(times []time.Duration) time.Duration {
	if len(times) == 0 {
		return 0
	}
	return times[len(times)-1]
}

func getPercentile(times []time.Duration, percentile int) time.Duration {
	if len(times) == 0 {
		return 0
	}
	index := int(float64(len(times)) * float64(percentile) / 100.0)
	if index >= len(times) {
		index = len(times) - 1
	}
	return times[index]
}
