package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const defaultURL = "http://localhost:8080/api/v1"

type payload struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	OrderID   string `json:"order_id"`
}

func main() {
	// CLI Flags
	productID := flag.String("id", "", "Target Product UUID (required)")
	count := flag.Int("n", 50, "Total number of requests")
	concurrency := flag.Int("c", 50, "Max concurrent workers")
	flag.Parse()

	if *productID == "" {
		flag.Usage()
		os.Exit(1)
	}

	log.SetFlags(0) // Disable timestamp for cleaner output
	log.Printf("Starting stress test on %s...", *productID)
	log.Printf("Config: %d requests, %d concurrency", *count, *concurrency)

	var (
		success   atomic.Int64
		conflicts atomic.Int64 // 409 Optimistic Lock
		failures  atomic.Int64
	)

	start := time.Now()
	wg := &sync.WaitGroup{}
	sem := make(chan struct{}, *concurrency) // Semaphore to control concurrency

	for i := 0; i < *count; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			if err := sendRequest(*productID, idx, &success, &conflicts, &failures); err != nil {
				log.Printf("[Err] Worker %d: %v", idx, err)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	printReport(duration, *count, success.Load(), conflicts.Load(), failures.Load())
}

func sendRequest(pid string, idx int, success, conflicts, failures *atomic.Int64) error {
	data := payload{
		ProductID: pid,
		Quantity:  1,
		OrderID:   fmt.Sprintf("STRESS-%d", idx),
	}

	body, _ := json.Marshal(data)
	resp, err := http.Post(defaultURL+"/stock/reserve", "application/json", bytes.NewBuffer(body))
	if err != nil {
		failures.Add(1)
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) // Drain body

	switch resp.StatusCode {
	case http.StatusOK:
		success.Add(1)
	case http.StatusConflict:
		conflicts.Add(1)
	default:
		failures.Add(1)
	}
	return nil
}

func printReport(d time.Duration, total int, success, conflicts, failures int64) {
	rps := float64(total) / d.Seconds()

	fmt.Println("\n--- Stress Test Report ---")
	fmt.Printf("Time Taken      : %v\n", d)
	fmt.Printf("Throughput      : %.2f Req/sec\n", rps)
	fmt.Printf("Success (200)   : %d\n", success)
	fmt.Printf("Conflicts (409) : %d (Optimistic Lock Hit)\n", conflicts)
	fmt.Printf("Failures        : %d\n", failures)
	fmt.Println("--------------------------")
}
