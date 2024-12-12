package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

func main() {
	const urlCount uint64 = 5
	const repeatCount = 5

	urls, _ := GetDlUrls(urlCount)

	if len(urls) == 0 {
		fmt.Println("No URLs retrieved. Exiting.")
		return
	}

	var overallAverageSpeed float64

	for i := 0; i < repeatCount; i++ {
		averageSpeed := measureParallelDownloadSpeed(urls)
		overallAverageSpeed += averageSpeed
		fmt.Printf("Speed: %.2f Mbps\n", averageSpeed)
	}

	fmt.Printf("Speed: %.2f Mbps\n", overallAverageSpeed/float64(repeatCount))
}

func measureParallelDownloadSpeed(urls []string) float64 {
	var wg sync.WaitGroup
	var mu sync.Mutex

	totalBytes := int64(0)
	startTime := time.Now()

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			bytesDownloaded := downloadFile(url)
			mu.Lock()
			totalBytes += bytesDownloaded
			mu.Unlock()
		}(url)
	}

	wg.Wait()
	totalDuration := time.Since(startTime).Seconds()

	if totalDuration == 0 {
		return 0
	}

	// Convert bytes to megabits and calculate speed
	megabits := float64(totalBytes) * 8 / 1_000_000
	averageSpeed := megabits / totalDuration
	return averageSpeed
}

func downloadFile(url string) int64 {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error downloading from %s: %v\n", url, err)
		return 0
	}
	defer resp.Body.Close()

	totalBytes := int64(0)
	buffer := make([]byte, 16 * 1024 * 1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			totalBytes += int64(n)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading response body from %s: %v\n", url, err)
			return totalBytes
		}
	}
	return totalBytes
}
