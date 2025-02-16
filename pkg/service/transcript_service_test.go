package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/horiagug/youtube-transcript-api-go/pkg/models"
	"github.com/horiagug/youtube-transcript-api-go/pkg/repository/fixtures"
)

func TestNewTranscriptService(t *testing.T) {
	fetcher := &fixtures.MockHTMLFetcher{}
	service := NewTranscriptService(fetcher)
	assert.NotNil(t, service, "Service should not be nil")
}

func TestGetTranscripts(t *testing.T) {
	tests := []struct {
		name               string
		videoID            string
		languages          []string
		preserveFormatting bool
		mockVideoHTML      string
		mockTranscriptXML  string
		expectedError      error
		expectedResult     []models.Transcript
	}{
		{
			name:               "Success case - Single transcript",
			videoID:            "abc123",
			languages:          []string{"en"},
			preserveFormatting: false,
			mockVideoHTML:      `{"captions":{"playerCaptionsTracklistRenderer":{"captionTracks":[{"baseUrl":"http://example.com/transcript","name":{"simpleText":"English"},"languageCode":"en","kind":"asr","isTranslatable":true}]}},"videoDetails":{"someKey":"some details"}}`,
			mockTranscriptXML: `<?xml version="1.0" encoding="utf-8" ?><transcript>
		              <text start="0" dur="1">Hello world</text>
		          </transcript>`,
			expectedError: nil,
			expectedResult: []models.Transcript{
				{
					VideoID:        "abc123",
					Language:       "English",
					LanguageCode:   "en",
					IsGenerated:    false,
					IsTranslatable: true,
					Lines: []models.TranscriptLine{
						{
							Text:     "Hello world",
							Start:    0,
							Duration: 1,
						},
					},
				},
			},
		},
		{
			name:          "Too many requests",
			videoID:       "abc123",
			languages:     []string{"en"},
			mockVideoHTML: `<div class="g-recaptcha"></div>`,
			expectedError: errors.New("failed to extract list of transcripts: TooManyRequests"),
		},
		{
			name:          "No Playability Status",
			videoID:       "abc123",
			languages:     []string{"en"},
			mockVideoHTML: `{"someOtherData": true}`,
			expectedError: errors.New("failed to extract list of transcripts: VideoUnavailable"),
		},
		{
			name:          "No Transcript Data",
			videoID:       "nonexistent",
			languages:     []string{"en"},
			mockVideoHTML: `{"playabilityStatus": {"status": "ERROR"}}`,
			expectedError: errors.New("failed to extract list of transcripts: NoTranscriptData"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := &fixtures.MockHTMLFetcher{}

			if tt.mockVideoHTML != "" {
				fetcher.On("FetchVideo", mock.AnythingOfType("string")).Return([]byte(tt.mockVideoHTML), nil)
			}

			if tt.mockTranscriptXML != "" {
				fetcher.On("Fetch", mock.AnythingOfType("string"), mock.Anything).Return([]byte(tt.mockTranscriptXML), nil)
			}

			service := NewTranscriptService(fetcher)
			result, err := service.GetTranscripts(tt.videoID, tt.languages, tt.preserveFormatting)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			fetcher.AssertExpectations(t)
		})
	}
}

func TestSanitizeVideoID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Regular video ID",
			input:    "dQw4w9WgXcQ",
			expected: "dQw4w9WgXcQ",
		},
		{
			name:     "YouTube URL",
			input:    "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			expected: "dQw4w9WgXcQ",
		},
		{
			name:     "YouTube URL with additional parameters",
			input:    "https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=42s",
			expected: "dQw4w9WgXcQ",
		},
		{
			name:     "Invalid URL",
			input:    "https://example.com/video",
			expected: "https://example.com/video",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeVideoId(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessCaptionTracks(t *testing.T) {
	t.Run("Process multiple tracks concurrently", func(t *testing.T) {
		fetcher := &fixtures.MockHTMLFetcher{}
		service := NewTranscriptService(fetcher)

		captionTracks := []models.CaptionTrack{
			{
				BaseUrl:        "http://example.com/en",
				Name:           models.LanguageName{SimpleText: "English"},
				LanguageCode:   "en",
				IsTranslatable: true,
			},
			{
				BaseUrl:        "http://example.com/es",
				Name:           models.LanguageName{SimpleText: "Spanish"},
				LanguageCode:   "es",
				IsTranslatable: true,
			},
		}

		mockXML := `<?xml version="1.0" encoding="utf-8" ?><transcript>
            <text start="0" dur="1">Test content</text>
        </transcript>`

		fetcher.On("Fetch", "http://example.com/en", mock.Anything).Return([]byte(mockXML), nil)
		fetcher.On("Fetch", "http://example.com/es", mock.Anything).Return([]byte(mockXML), nil)

		results := service.processCaptionTracks("test123", captionTracks, false)

		assert.Len(t, results, 2)
		assert.Equal(t, "English", results[0].Language)
		assert.Equal(t, "Spanish", results[1].Language)
		assert.Equal(t, "Test content", results[0].Lines[0].Text)
		assert.Equal(t, "Test content", results[1].Lines[0].Text)

		fetcher.AssertExpectations(t)
	})

	t.Run("Handle failed track processing", func(t *testing.T) {
		fetcher := &fixtures.MockHTMLFetcher{}
		service := NewTranscriptService(fetcher)

		captionTracks := []models.CaptionTrack{
			{
				BaseUrl:      "http://example.com/en",
				Name:         models.LanguageName{SimpleText: "English"},
				LanguageCode: "en",
			},
		}

		fetcher.On("Fetch", "http://example.com/en", mock.Anything).
			Return([]byte{}, errors.New("failed to fetch"))

		results := service.processCaptionTracks("test123", captionTracks, false)

		assert.Empty(t, results)
		fetcher.AssertExpectations(t)
	})
}
