package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/horiagug/youtube-transcript-api-go/internal/repository"
	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript_models"
	"golang.org/x/net/html"
)

type TranscriptService interface {
	GetTranscripts(videoID string, langauges []string, preserve_formatting bool) ([]yt_transcript_models.Transcript, error)
}

type transcriptService struct {
	fetcher repository.HTMLFetcherType
}

type transcriptResult struct {
	transcript yt_transcript_models.Transcript
	err        error
}

func NewTranscriptService(fetcher repository.HTMLFetcherType) *transcriptService {
	return &transcriptService{
		fetcher: fetcher,
	}
}

func (t transcriptService) GetTranscripts(videoID string, languages []string, preserve_formatting bool) ([]yt_transcript_models.Transcript, error) {

	videoID = sanitizeVideoId(videoID)

	trascript_data, err := t.extractTranscriptList(videoID)
	if err != nil {
		return []yt_transcript_models.Transcript{}, fmt.Errorf("failed to extract list of transcripts: %w", err)
	}

	transcripts, err := t.getTranscriptsForLanguage(languages, *trascript_data.Transcripts)
	if err != nil {
		return []yt_transcript_models.Transcript{}, fmt.Errorf("failed to get transcript: %w", err)
	}

	return t.processCaptionTracks(videoID, transcripts, trascript_data.Title, preserve_formatting)
}

func (t *transcriptService) processCaptionTracks(video_id string, captionTracks []yt_transcript_models.CaptionTrack, title string, preserve_formatting bool) ([]yt_transcript_models.Transcript, error) {
	resultChan := make(chan transcriptResult, len(captionTracks))
	var wg sync.WaitGroup

	for _, transcript := range captionTracks {
		wg.Add(1)
		go func(tr yt_transcript_models.CaptionTrack) {
			defer wg.Done()

			is_generated := true
			if tr.Kind != nil && *tr.Kind == "asr" {
				is_generated = false
			}

			lines, err := t.getTranscriptFromTrack(tr, preserve_formatting)
			if err != nil {
				resultChan <- transcriptResult{err: fmt.Errorf("error getting transcript from track: %w", err)}
				return
			}

			result := yt_transcript_models.Transcript{
				VideoID:        video_id,
				VideoTitle:     title,
				Language:       tr.Name.SimpleText,
				LanguageCode:   tr.LanguageCode,
				IsGenerated:    is_generated,
				IsTranslatable: tr.IsTranslatable,
				Lines:          lines,
			}

			resultChan <- transcriptResult{transcript: result}
		}(transcript)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var results []yt_transcript_models.Transcript
	for result := range resultChan {
		if result.err != nil {
			fmt.Printf("Error processing transcript: %v\n", result.err)
			return results, result.err
		}
		results = append(results, result.transcript)
	}
	return results, nil
}

func extractInnerTubeApiKey(htmlContent string) string {
	// Define the regex pattern for INNERTUBE_API_KEY
	pattern := `"INNERTUBE_API_KEY":\s*"([a-zA-Z0-9_-]+)"`
	re := regexp.MustCompile(pattern)

	// Search for the pattern in the HTML content
	match := re.FindStringSubmatch(htmlContent)
	if len(match) == 2 {
		return match[1]
	}

	return ""
}

func extractTitle(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		fmt.Printf("Error fetching the title")
		return ""
	}

	var title string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			if n.FirstChild != nil {
				title = n.FirstChild.Data
				return
			}
		}
		for c := n.FirstChild; c != nil && title == ""; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)
	return title
}

func (t *transcriptService) extractTranscriptList(video_id string) (*yt_transcript_models.VideoTranscriptData, error) {
	html, err := t.fetcher.FetchVideo(video_id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video page: %w", err)
	}

	body := string(html)

	title := extractTitle(body)

	innertube_api_key := extractInnerTubeApiKey(body)

	innertube_data, err := t.fetcher.FetchInnertubeData(video_id, innertube_api_key)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch video page: %w", err)
	}

	innertube_data_json, err := json.Marshal(innertube_data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal innertube data: %w", err)
	}

	var videoDetails yt_transcript_models.InnertubeData
	err = json.Unmarshal(innertube_data_json, &videoDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if videoDetails.Captions.PlayerCaptionsTracklistRenderer == nil {
		return nil, fmt.Errorf("playerCaptionsTracklistRenderer not found")
	}

	transcripts := videoDetails.Captions.PlayerCaptionsTracklistRenderer

	return &yt_transcript_models.VideoTranscriptData{Transcripts: transcripts, Title: title}, nil
}

func (s transcriptService) getTranscriptsForLanguage(languages []string, transcripts yt_transcript_models.TranscriptData) ([]yt_transcript_models.CaptionTrack, error) {

	if len(languages) == 0 {
		return transcripts.CaptionTracks, nil
	}

	caption_tracks := make([]yt_transcript_models.CaptionTrack, 0)

	for _, lang := range languages {
		for _, track := range transcripts.CaptionTracks {
			if track.LanguageCode == lang {
				caption_tracks = append(caption_tracks, track)
			}
		}
	}

	if len(caption_tracks) == 0 {
		return []yt_transcript_models.CaptionTrack{}, fmt.Errorf("no transcript found for languages %s", languages)
	}

	return caption_tracks, nil
}

func (s transcriptService) getTranscriptFromTrack(track yt_transcript_models.CaptionTrack, preserve_formatting bool) ([]yt_transcript_models.TranscriptLine, error) {
	url := strings.Replace(track.BaseUrl, "&fmt=srv3", "", -1)
	body, err := s.fetcher.Fetch(url, nil)
	if err != nil {
		return []yt_transcript_models.TranscriptLine{}, fmt.Errorf("failed to fetch transcript: %w", err)
	}

	parser := repository.NewTranscriptParser(preserve_formatting)

	transcript, err := parser.Parse(string(body))
	if err != nil {
		return []yt_transcript_models.TranscriptLine{}, fmt.Errorf("failed to parse transcript: %w", err)
	}
	return transcript, nil
}

func sanitizeVideoId(videoID string) string {
	if strings.HasPrefix(videoID, "http://") || strings.HasPrefix(videoID, "https://") || strings.HasPrefix(videoID, "www.") {
		if strings.Contains(videoID, "youtube.com") {
			u, err := url.Parse(videoID)
			if err != nil {
				fmt.Println("Error parsing URL")
			}
			return u.Query().Get("v")
		}
		fmt.Println("Warning: this doesn't look like a youtube video, we'll still try to process it.")
	}
	return videoID
}
