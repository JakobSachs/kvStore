package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	. "jakobsachs.blog/kvStore/shared"
)

const (
	keyLength      = 5
	valueLength    = 1000
	testDuration   = 20 * time.Second
	numWorkers     = 50 // Concurrency level
	serverURL      = "http://localhost:8080/"
	readWriteRatio = 0.25 // 0.5 means 50% writes, 50% reads (for writes)
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// sendRequestToServer sends a request and returns the response body, request size, and an error.
func sendRequestToServer(req Request) ([]byte, int, error) {
	jsonData, err := Serialize(req)
	if err != nil {
		return nil, 0, fmt.Errorf("error marshalling request to JSON: %w", err)
	}
	requestSize := len(jsonData)

	httpReq, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, requestSize, fmt.Errorf("error creating HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json") // Assuming server expects JSON

	client := &http.Client{Timeout: 5 * time.Second} // Add a timeout
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, requestSize, fmt.Errorf("error sending HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read body for more error info, but don't fail fatally
		bodyBytes, _ := io.ReadAll(resp.Body)
		return bodyBytes, requestSize, fmt.Errorf("server returned non-OK status: %s, body: %s", resp.Status, string(bodyBytes))
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, requestSize, fmt.Errorf("error reading response body: %w", err)
	}
	return responseBody, requestSize, nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var totalRequests, successfulRequests, failedRequests atomic.Uint64
	var totalBytesWritten, totalBytesRead atomic.Uint64
	var currentRequestID atomic.Uint64

	var wg sync.WaitGroup
	startTime := time.Now()
	done := make(chan struct{})

	log.Printf("Starting load test for %v with %d workers...", testDuration, numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					reqID := currentRequestID.Add(1)
					var req Request
					req.Id = reqID

					if rand.Float64() < readWriteRatio { // Write operation
						req.Type = Write
						req.Key = randomString(keyLength)
						req.Value = randomString(valueLength)
					} else { // Read operation
						req.Type = Read
						req.Key = randomString(keyLength) // Could read a non-existent key
					}

					totalRequests.Add(1)
					responseBody, reqSize, err := sendRequestToServer(req)
					totalBytesWritten.Add(uint64(reqSize))

					if err != nil {
						failedRequests.Add(1)
						// Optionally log errors, but be mindful of log volume
						// log.Printf("Request failed: %v, ReqID: %d", err, reqID)
					} else {
						successfulRequests.Add(1)
						totalBytesRead.Add(uint64(len(responseBody)))
					}
				}
			}
		}()
	}

	time.Sleep(testDuration)
	close(done)
	wg.Wait()

	elapsedTime := time.Since(startTime)
	totalReq := totalRequests.Load()
	successfulReq := successfulRequests.Load()
	failedReq := failedRequests.Load()
	bytesWritten := totalBytesWritten.Load()
	bytesRead := totalBytesRead.Load()

	fmt.Println("\n--- Load Test Results ---")
	fmt.Printf("Test Duration: %s\n", elapsedTime.Round(time.Millisecond))
	fmt.Printf("Total Requests: %d\n", totalReq)
	fmt.Printf("Successful Requests: %d\n", successfulReq)
	fmt.Printf("Failed Requests: %d\n", failedReq)

	if elapsedTime.Seconds() > 0 {
		rps := float64(totalReq) / elapsedTime.Seconds()
		successfulRps := float64(successfulReq) / elapsedTime.Seconds()
		fmt.Printf("Total RPS: %.2f\n", rps)
		fmt.Printf("Successful RPS: %.2f\n", successfulRps)

		writeThroughput := float64(bytesWritten) / elapsedTime.Seconds() / 1024 // KB/s
		readThroughput := float64(bytesRead) / elapsedTime.Seconds() / 1024     // KB/s
		fmt.Printf("Write Throughput: %.2f KB/s\n", writeThroughput)
		fmt.Printf("Read Throughput: %.2f KB/s\n", readThroughput)
	} else {
		fmt.Println("Test duration too short to calculate meaningful rates.")
	}
	fmt.Printf("Total Bytes Written: %d bytes\n", bytesWritten)
	fmt.Printf("Total Bytes Read: %d bytes\n", bytesRead)
}
