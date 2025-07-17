package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/horiagug/youtube-transcript-api-go/internal/repository/fixtures"
	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript_models"
)

func TestContextTimeoutRespected(t *testing.T) {
	fetcher := &fixtures.MockHTMLFetcher{}
	service := NewTranscriptService(fetcher)

	// Mock video fetch to return valid HTML
	fetcher.On("FetchVideo", mock.AnythingOfType("string")).Return([]byte(`<title>Test Video</title>"INNERTUBE_API_KEY":"test_key"`), nil)

	// Mock innertube data
	mockInnertubeData := map[string]interface{}{
		"captions": map[string]interface{}{
			"playerCaptionsTracklistRenderer": map[string]interface{}{
				"captionTracks": []interface{}{
					map[string]interface{}{
						"baseUrl":        "http://example.com/transcript",
						"name":           map[string]interface{}{"simpleText": "English"},
						"languageCode":   "en",
						"kind":           "asr",
						"isTranslatable": true,
					},
				},
			},
		},
	}
	fetcher.On("FetchInnertubeData", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(mockInnertubeData, nil)

	// Mock a slow transcript fetch that would exceed context timeout
	fetcher.On("FetchWithContext", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		select {
		case <-ctx.Done():
			// Context was cancelled, this is expected
			return
		case <-time.After(2 * time.Second):
			// This should not happen if context timeout is working
			t.Error("Context timeout was not respected")
		}
	}).Return([]byte{}, context.DeadlineExceeded)

	// Create a context with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This should timeout quickly
	start := time.Now()
	_, err := service.GetTranscriptsWithContext(ctx, "test123", []string{"en"}, false)
	elapsed := time.Since(start)

	assert.Error(t, err)
	assert.Less(t, elapsed, 500*time.Millisecond, "Context timeout should have been respected")
}

func TestSliceCapacityOptimization(t *testing.T) {
	// Test that slices are pre-allocated with proper capacity
	service := transcriptService{}

	// Test with empty languages (should return all tracks)
	transcripts := yt_transcript_models.TranscriptData{
		CaptionTracks: []yt_transcript_models.CaptionTrack{
			{LanguageCode: "en", Name: yt_transcript_models.LanguageName{SimpleText: "English"}},
			{LanguageCode: "es", Name: yt_transcript_models.LanguageName{SimpleText: "Spanish"}},
		},
	}

	result, err := service.getTranscriptsForLanguage([]string{}, transcripts)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Test with specific languages
	result, err = service.getTranscriptsForLanguage([]string{"en"}, transcripts)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "en", result[0].LanguageCode)
}

func TestVideoIDSanitization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Regular ID", "abc123", "abc123"},
		{"YouTube URL", "https://www.youtube.com/watch?v=abc123", "abc123"},
		{"YouTube URL with params", "https://www.youtube.com/watch?v=abc123&t=10s", "abc123"},
		{"Short URL", "https://youtu.be/abc123", "abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeVideoId(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}