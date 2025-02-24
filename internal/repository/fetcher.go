package repository

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
)

var video_base_url = "https://www.youtube.com/watch?v=%s"

type HTMLFetcherType interface {
	Fetch(url string, cookie *http.Cookie) ([]byte, error)
	FetchVideo(videoID string) ([]byte, error)
}

type HTMLFetcher struct{}

func NewHTMLFetcher() *HTMLFetcher {
	return &HTMLFetcher{}
}

func (f *HTMLFetcher) Fetch(url string, cookie *http.Cookie) ([]byte, error) {
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
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
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
