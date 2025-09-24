package service

import (
	"testing"

	"github.com/horiagug/youtube-transcript-api-go/internal/repository"
	"github.com/horiagug/youtube-transcript-api-go/internal/repository/fixtures"
	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript_models"
	"github.com/stretchr/testify/mock"
)

func BenchmarkProcessCaptionTracks(b *testing.B) {
	fetcher := &fixtures.MockHTMLFetcher{}
	service := NewTranscriptService(fetcher)

	// Create multiple caption tracks to simulate real usage
	captionTracks := make([]yt_transcript_models.CaptionTrack, 5)
	for i := 0; i < 5; i++ {
		captionTracks[i] = yt_transcript_models.CaptionTrack{
			BaseUrl:        "http://example.com/transcript",
			Name:           yt_transcript_models.LanguageName{SimpleText: "English"},
			LanguageCode:   "en",
			IsTranslatable: true,
		}
	}

	mockXML := `<?xml version="1.0" encoding="utf-8" ?><transcript>
		<text start="0" dur="1">Hello world</text>
		<text start="1" dur="1">This is a test</text>
		<text start="2" dur="1">Performance benchmark</text>
		<text start="3" dur="1">Multiple lines</text>
		<text start="4" dur="1">Final line</text>
	</transcript>`

	// Mock all the fetch calls
	fetcher.On("FetchWithContext", mock.Anything, mock.Anything, mock.Anything).Return([]byte(mockXML), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.processCaptionTracks("test123", captionTracks, "Test Video", false)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRegexCompilation(b *testing.B) {
	xmlContent := `<?xml version="1.0" encoding="utf-8" ?><transcript>
		<text start="0" dur="1"><b>Bold text</b> and <i>italic text</i> with <strong>strong text</strong></text>
		<text start="1" dur="1">Regular text</text>
	</transcript>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test the optimized regex compilation via parsing
		parser := repository.NewTranscriptParser(true)
		_, err := parser.Parse(xmlContent)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExtractInnertubeVideoDetails(b *testing.B) {
	// Sample data structure similar to what YouTube returns
	mockData := map[string]interface{}{
		"captions": map[string]interface{}{
			"playerCaptionsTracklistRenderer": map[string]interface{}{
				"captionTracks": []interface{}{
					map[string]interface{}{
						"baseUrl":        "http://example.com/transcript1",
						"name":           map[string]interface{}{"simpleText": "English"},
						"languageCode":   "en",
						"kind":           "asr",
						"isTranslatable": true,
					},
					map[string]interface{}{
						"baseUrl":        "http://example.com/transcript2",
						"name":           map[string]interface{}{"simpleText": "Spanish"},
						"languageCode":   "es",
						"isTranslatable": true,
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := extractInnertubeVideoDetails(mockData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

