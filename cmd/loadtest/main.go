package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

func main() {
	targetRPS := flag.Int("rps", 1000, "целевое число запросов в секунду")
	duration := flag.Duration("duration", 10*time.Second, "как долго гнать нагрузку")
	baseURL := flag.String("url", "http://localhost:8082", "адрес сервера")
	flag.Parse()

	walletID := uuid.New().String()
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        2000,
			MaxIdleConnsPerHost: 2000,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	var (
		total       int64
		success     int64
		clientErr   int64 
		serverErr   int64 
		mu          sync.Mutex
		latencies   []time.Duration
		wg          sync.WaitGroup
	)

	interval := time.Second / time.Duration(*targetRPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	deadline := time.Now().Add(*duration)

	for time.Now().Before(deadline) {
		<-ticker.C
		wg.Add(1)
		go func() {
			defer wg.Done()

			body, _ := json.Marshal(map[string]any{
				"valletId":      walletID,
				"operationType": "DEPOSIT",
				"amount":        1,
			})

			start := time.Now()
			resp, err := client.Post(*baseURL+"/api/v1/wallet", "application/json", bytes.NewReader(body))
			elapsed := time.Since(start)

			atomic.AddInt64(&total, 1)

			if err != nil {
				atomic.AddInt64(&serverErr, 1)
				return
			}
			defer resp.Body.Close()

			mu.Lock()
			latencies = append(latencies, elapsed)
			mu.Unlock()

			switch {
			case resp.StatusCode >= 500:
				n := atomic.AddInt64(&serverErr, 1)
				if n <= 3 { 
					respBody, _ := io.ReadAll(resp.Body)
					fmt.Printf("[5xx sample] status=%d elapsed=%s body=%q\n", resp.StatusCode, elapsed, string(respBody))
				}
			case resp.StatusCode >= 400:
				atomic.AddInt64(&clientErr, 1)
			default:
				atomic.AddInt64(&success, 1)
			}
		}()
	}

	wg.Wait()

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	p50 := percentile(latencies, 50)
	p99 := percentile(latencies, 99)

	fmt.Println("=== Load test result ===")
	fmt.Printf("target RPS: %d, duration: %s\n", *targetRPS, *duration)
	fmt.Printf("total requests: %d\n", total)
	fmt.Printf("success (2xx):  %d\n", success)
	fmt.Printf("client err (4xx): %d\n", clientErr)
	fmt.Printf("SERVER ERR (5xx): %d  <-- должно быть 0\n", serverErr)
	fmt.Printf("p50 latency: %s\n", p50)
	fmt.Printf("p99 latency: %s\n", p99)
	fmt.Printf("actual throughput: %.0f req/sec\n", float64(total)/duration.Seconds())
}

func percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := len(sorted) * p / 100
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}