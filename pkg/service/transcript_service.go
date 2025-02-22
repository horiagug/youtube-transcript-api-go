package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/horiagug/youtube-transcript-api-go/pkg/models"
	"github.com/horiagug/youtube-transcript-api-go/pkg/repository"
)

type TranscriptService struct {
	fetcher HTMLFetcherType
}

type HTMLFetcherType interface {
	Fetch(url string, cookie *http.Cookie) ([]byte, error)
	FetchVideo(videoID string) ([]byte, error)
}

func NewTranscriptService(fetcher HTMLFetcherType) *TranscriptService {
	return &TranscriptService{
		fetcher: fetcher,
	}
}

func (t *TranscriptService) GetTranscripts(videoID string, languages []string, preserve_formatting bool) ([]models.Transcript, error) {

	videoID = sanitizeVideoId(videoID)

	trascript_data, err := t.extractTranscriptList(videoID)
	if err != nil {
		return []models.Transcript{}, fmt.Errorf("failed to extract list of transcripts: %w", err)
	}

	transcripts, err := t.getTranscriptsForLanguage(languages, *trascript_data)
	if err != nil {
		return []models.Transcript{}, fmt.Errorf("failed to get transcript: %w", err)
	}

	return t.processCaptionTracks(videoID, transcripts, preserve_formatting), nil
}

type transcriptResult struct {
	transcript models.Transcript
	err        error
}

func (t *TranscriptService) processCaptionTracks(video_id string, captionTracks []models.CaptionTrack, preserve_formatting bool) []models.Transcript {
	resultChan := make(chan transcriptResult, len(captionTracks))
	var wg sync.WaitGroup

	// Launch goroutines for each caption track
	for _, transcript := range captionTracks {
		wg.Add(1)
		go func(tr models.CaptionTrack) {
			defer wg.Done()

			// Process single transcript
			is_generated := true
			if tr.Kind != nil && *tr.Kind == "asr" {
				is_generated = false
			}

			lines, err := t.getTranscriptFromTrack(tr, preserve_formatting)
			if err != nil {
				resultChan <- transcriptResult{err: fmt.Errorf("error getting transcript from track: %w", err)}
				return
			}

			result := models.Transcript{
				VideoID:        video_id,
				Language:       tr.Name.SimpleText,
				LanguageCode:   tr.LanguageCode,
				IsGenerated:    is_generated,
				IsTranslatable: tr.IsTranslatable,
				Lines:          lines,
			}

			resultChan <- transcriptResult{transcript: result}
		}(transcript)
	}

	// Close channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var results []models.Transcript
	for result := range resultChan {
		if result.err != nil {
			fmt.Printf("Error processing transcript: %v\n", result.err)
			continue
		}
		results = append(results, result.transcript)
	}
	return results
}

func (t *TranscriptService) extractTranscriptList(video_id string) (*models.TranscriptData, error) {
	// get the html
	html, err := t.fetcher.FetchVideo(video_id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video page: %w", err)
	}

	body := string(html)

	parts := strings.Split(body, `"captions":`)

	if len(parts) <= 1 {
		if strings.Contains(body, `class="g-recaptcha"`) {
			return nil, fmt.Errorf("TooManyRequests")
		}
		if !strings.Contains(body, `"playabilityStatus":`) {
			return nil, fmt.Errorf("VideoUnavailable")
		}
		return nil, fmt.Errorf("NoTranscriptData")
	}

	if !strings.Contains(body, `"captions":`) {
		return nil, fmt.Errorf("NoTranscriptData")
	}

	video_details := strings.Split(parts[1], `,"videoDetails`)[0]
	video_details_parsed := strings.ReplaceAll(video_details, "\n", "")

	var videoDetails models.VideoDetails
	err = json.Unmarshal([]byte(video_details_parsed), &videoDetails)
	if err != nil {
		fmt.Println("Error unmarshalling JSON")
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if videoDetails.PlayerCaptionsTracklistRenderer == nil {
		return nil, fmt.Errorf("playerCaptionsTracklistRenderer not found")
	}

	transcripts := videoDetails.PlayerCaptionsTracklistRenderer

	return transcripts, nil
}

func (s TranscriptService) getTranscriptsForLanguage(language []string, transcripts models.TranscriptData) ([]models.CaptionTrack, error) {

	if len(language) == 0 {
		return transcripts.CaptionTracks, nil
	}

	caption_tracks := make([]models.CaptionTrack, 0)

	for _, lang := range language {
		for _, track := range transcripts.CaptionTracks {
			if track.LanguageCode == lang {
				caption_tracks = append(caption_tracks, track)
			}
		}
	}

	if len(caption_tracks) == 0 {
		return []models.CaptionTrack{}, fmt.Errorf("no transcript found for languages %s", language)
	}

	return caption_tracks, nil
}

func (s TranscriptService) getTranscriptFromTrack(track models.CaptionTrack, preserve_formatting bool) ([]models.TranscriptLine, error) {
	body, err := s.fetcher.Fetch(track.BaseUrl, nil)
	if err != nil {
		return []models.TranscriptLine{}, fmt.Errorf("failed to fetch transcript: %w", err)
	}

	parser := repository.NewTranscriptParser(preserve_formatting)

	transcript, err := parser.Parse(string(body))
	if err != nil {
		return []models.TranscriptLine{}, fmt.Errorf("failed to parse transcript: %w", err)
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
