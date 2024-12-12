package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
	//"strings"
)

func main() {
	const urlCount uint64 = 5
	const repeatCount = 5

	urls, _ := GetDlUrls(urlCount)
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

	for i := 0; i < 5; i++ {
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
	}

	wg.Wait()
	totalDuration := time.Since(startTime).Seconds()

	if totalDuration == 0 {
		return 0
	}

	// Convert bytes to megabits and calculate speed
	megabits := float64(totalBytes * 8) / 1_000_000.00
	averageSpeed := megabits / totalDuration
	return averageSpeed
}

func downloadFile(url string) int64 {
	//url = strings.Replace(url, "speedtest", "speedtest/range/0-", 1)
	//fmt.Printf("URL: %s\n", url)
	buffer := make([]byte, 1024 * 1024)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request for %s: %v\n", url, err)
		return 0
	}

	// Set the User-Agent header
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error downloading from %s: %v\n", url, err)
		return 0
	}
	defer resp.Body.Close()

	totalBytes := int64(0)
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
			break
		}
	}
	return totalBytes
}