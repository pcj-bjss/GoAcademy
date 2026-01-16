package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	// Configuration
	// We will hit the /get endpoint as it is safe to call repeatedly without filling up memory.
	targetURL := "http://localhost:8080/get"
	concurrency := 50      // Number of parallel users (goroutines)
	totalRequests := 10000 // Total requests to send across all users

	fmt.Printf("Bombarding %s with %d requests using %d concurrent workers...\n", targetURL, totalRequests, concurrency)

	// Used to ensure the main program doesn't exit until all 50 workers have finished their tasks.
	var wg sync.WaitGroup
	// Starts a stopwatch to measure the total duration.
	start := time.Now()

	// Channels to distribute work (jobs) and collect results
	jobs := make(chan struct{}, totalRequests)

	// Workers drop an error (or nil for success) in the results channel after every request so we can count them later.
	results := make(chan error, totalRequests)

	// 1. Start Workers
	// We spin up 'concurrency' number of goroutines. They all listen to the 'jobs' channel.
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Create a client with a timeout to prevent hanging
			client := &http.Client{
				Timeout: 5 * time.Second,
			}

			for range jobs {
				// Perform the HTTP Request
				resp, err := client.Get(targetURL)
				if err != nil {
					results <- err
					continue
				}
				// Crucial: We must close the body to ensure the TCP connection can be reused (Keep-Alive).
				// If we don't do this, we will run out of file descriptors/ports very quickly.
				resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					results <- fmt.Errorf("status code: %d", resp.StatusCode)
				} else {
					results <- nil // Success
				}
			}
		}()
	}

	// 2. Send Jobs
	// Fill the channel with empty structs. Each one represents a request to be made.
	for i := 0; i < totalRequests; i++ {
		jobs <- struct{}{} //struct{} uses zero memory, and is ideal for signaling without data. The second {} creates an instance of the empty struct value.
	}
	close(jobs) // Close the channel to signal workers that no more jobs are coming.

	// 3. Wait for completion
	wg.Wait()
	close(results)

	duration := time.Since(start)

	// 4. Calculate Statistics
	var successCount, failCount int
	for err := range results {
		if err != nil {
			failCount++
		} else {
			successCount++
		}
	}

	fmt.Println("\n--- Results ---")
	fmt.Printf("Time taken:   %v\n", duration)
	fmt.Printf("Total reqs:   %d\n", totalRequests)
	fmt.Printf("Success:      %d\n", successCount)
	fmt.Printf("Failed:       %d\n", failCount)
	fmt.Printf("Throughput:   %.2f req/sec\n", float64(totalRequests)/duration.Seconds())
}
