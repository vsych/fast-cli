package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

func main() {
	const urlCount uint64 = 5 // Number of download URLs to test

	// Fetch the download URLs from the API
	urls, _  := GetDlUrls(urlCount)
	if len(urls) == 0 {
		fmt.Println("No URLs retrieved. Exiting.")
		return
	}

	// Measure download speeds
	var wg sync.WaitGroup
	var totalSpeed float64
	var mu sync.Mutex // To safely sum speeds across goroutines

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			speed := measureDownloadSpeed(url)
			if speed > 0 {
				fmt.Printf("Download speed for %s: %.2f Mbps\n", url, speed)
				mu.Lock()
				totalSpeed += speed // Add to total in a thread-safe way
				mu.Unlock()
			} else {
				fmt.Printf("Failed to measure speed for %s\n", url)
			}
		}(url)
	}

	wg.Wait()

	// Print aggregate speed
	fmt.Printf("Total aggregate download speed: %.2f Mbps\n", totalSpeed)
}

func measureDownloadSpeed(url string) float64 {
	fmt.Printf("Starting download from: %s\n", url)

	startTime := time.Now()

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error downloading from %s: %v\n", url, err)
		return 0
	}
	defer resp.Body.Close()

	totalBytes := int64(0)
	buffer := make([]byte, 16 * 1024 * 1024) // 16mb buffer
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			totalBytes += int64(n)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading response body: %v\n", err)
			return 0
		}
	}

	duration := time.Since(startTime).Seconds()
	if duration == 0 {
		return 0
	}

	// Convert bytes to megabits and calculate speed
	megabits := float64(totalBytes) * 8 / 1_000_000
	speed := megabits / duration
	return speed
}
