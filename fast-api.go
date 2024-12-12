package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

// GetDlUrls returns a list of URLs to the fast API downloads
func GetDlUrls(urlCount uint64) ([]string, error) {
	token, err := getFastToken()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve API token for URL generation: %w", err)
	}

	url := fmt.Sprintf("https://api.fast.com/netflix/speedtest/v2?https=true&token=%s&urlCount=%d", token, urlCount)
	jsonData, err := getPage(url)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch URL list from API endpoint: %w", err)
	}

	re := regexp.MustCompile(`"url":"(.*?)"`)
	reUrls := re.FindAllStringSubmatch(jsonData, -1)

	var urls []string
	for _, arr := range reUrls {
		urls = append(urls, arr[1])
	}

	return urls, nil
}

func getFastToken() (string, error) {
	baseURL := "https://fast.com"
	fastBody, err := getPage(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch HTML content from %s: %w", baseURL, err)
	}

	re := regexp.MustCompile(`app-.*\.js`)
	scriptNames := re.FindAllString(fastBody, 1)
	if len(scriptNames) == 0 {
		return "", fmt.Errorf("no JavaScript app file found in the HTML content of %s", baseURL)
	}

	scriptURL := fmt.Sprintf("%s/%s", baseURL, scriptNames[0])

	scriptBody, err := getPage(scriptURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch JavaScript app file from %s: %w", scriptURL, err)
	}

	re = regexp.MustCompile(`token:"([a-zA-Z]*)"`)
	tokens := re.FindAllString(scriptBody, 1)

	if len(tokens) > 0 {
		token := tokens[0][7 : len(tokens[0])-1]
		return token, nil
	}

	return "", fmt.Errorf("no API token found in the JavaScript app file at %s", scriptURL)
}

func getPage(url string) (string, error) {
	buffer := bytes.NewBuffer(nil)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to perform HTTP GET request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(buffer, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read HTTP response body from %s: %w", url, err)
	}

	return buffer.String(), nil
}
