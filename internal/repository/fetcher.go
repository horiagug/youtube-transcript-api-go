package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

var video_base_url = "https://www.youtube.com/watch?v=%s"

const INNERTUBE_API_URL = "https://www.youtube.com/youtubei/v1/player?key=%s"

var INNERTUBE_CONTEXT = map[string]interface{}{
	"client": map[string]interface{}{
		"clientName":    "ANDROID",
		"clientVersion": "20.10.38",
	},
}

type HTMLFetcherType interface {
	Fetch(url string, cookie *http.Cookie) ([]byte, error)
	FetchVideo(videoID string) ([]byte, error)
	FetchInnertubeData(videoID string, apiKey string) (map[string]interface{}, error)
}

type HTMLFetcher struct{}

func NewHTMLFetcher() *HTMLFetcher {
	return &HTMLFetcher{}
}

func (f *HTMLFetcher) Fetch(url string, cookie *http.Cookie) ([]byte, error) {
	var body []byte
	var err error

	for i := range 3 {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Accept-Language", "en-US")
		if cookie != nil {
			req.AddCookie(cookie)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("Retry %d: failed to fetch: %v\n", i+1, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Retry %d: received non-OK status code: %d\n", i+1, resp.StatusCode)
			time.Sleep(2 * time.Second)
			continue
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Retry %d: failed to read response body: %v\n", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}

		if len(body) > 0 {
			return body, nil // Success
		}

		fmt.Printf("Retry %d: empty response body\n", i+1)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to fetch after retries: %w", err)
}

func (f *HTMLFetcher) FetchVideo(videoID string) ([]byte, error) {
	video_url := fmt.Sprintf(video_base_url, videoID)

	body, err := f.Fetch(video_url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video page: %w", err)
	}

	if consentRequired(body) {
		fmt.Println("Consent required, attempting to set cookie and retry")
		cookie, err := f.createConsentCookie(video_url)
		if err != nil {
			return nil, fmt.Errorf("failed to create consent cookie: %w", err)
		}

		body, err = f.Fetch(video_url, cookie) // Retry fetch with cookie
		if err != nil {
			return nil, fmt.Errorf("failed to fetch video page after setting consent: %w", err)
		}
	}

	return body, nil
}

func (f *HTMLFetcher) createConsentCookie(videoID string) (*http.Cookie, error) {
	html, err := f.Fetch(videoID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch HTML to extract consent value: %w", err)
	}

	re := regexp.MustCompile(`name="v" value="(.*?)"`)
	match := re.FindSubmatch(html)
	if len(match) < 2 {
		return nil, fmt.Errorf("failed to find consent value in HTML")
	}
	consentValue := string(match[1])

	cookieValue := "YES+" + consentValue
	cookie := &http.Cookie{
		Name:   "CONSENT",
		Value:  cookieValue,
		Domain: ".youtube.com",
	}
	return cookie, nil
}

func consentRequired(body []byte) bool {
	consentRegex := regexp.MustCompile(`action="https://consent\.youtube\.com/s`)
	return consentRegex.Match(body)
}

func (f *HTMLFetcher) FetchInnertubeData(videoID string, apiKey string) (map[string]interface{}, error) {

	url := fmt.Sprintf(INNERTUBE_API_URL, apiKey)

	payload := map[string]interface{}{
		"context": INNERTUBE_CONTEXT,
		"videoId": videoID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
	}

	var responseData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return responseData, nil
}
