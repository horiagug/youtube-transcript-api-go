package formatters

import (
	"encoding/json"

	"github.com/horiagug/youtube-transcript-api-go/pkg/models"
)

type JSONTranscripts struct {
	LanguageCode []models.TranscriptLine `json:"languageCode"`
}

// JSONFormatterOption is specifically for JSON formatter options
type JSONFormatterOption func(*JSONFormatter)

type JSONFormatter struct {
	BaseFormatter
	PrettyPrint bool
}

func NewJSONFormatter(baseOptions ...FormatterOption) *JSONFormatter {
	f := &JSONFormatter{
		BaseFormatter: BaseFormatter{
			IncludeTimestamps: true,
		},
		PrettyPrint: false,
	}

	// Apply base options
	for _, opt := range baseOptions {
		opt(&f.BaseFormatter)
	}
	return f
}

// WithPrettyPrint returns a function that sets the PrettyPrint option
func WithPrettyPrint(pretty bool) JSONFormatterOption {
	return func(f *JSONFormatter) {
		f.PrettyPrint = pretty
	}
}

// Configure allows applying JSON-specific options after creation
func (f *JSONFormatter) Configure(options ...JSONFormatterOption) {
	for _, opt := range options {
		opt(f)
	}
}

func (f *JSONFormatter) Format(transcripts []models.Transcript) (string, error) {
	jsonTranscripts := make([]JSONTranscripts, len(transcripts))

	for i, transcript := range transcripts {

		jsonTranscripts[i] = JSONTranscripts{
			LanguageCode: transcript.Lines,
		}
	}

	var (
		bytes []byte
		err   error
	)

	if f.PrettyPrint {
		bytes, err = json.MarshalIndent(jsonTranscripts, "", "  ")
	} else {
		bytes, err = json.Marshal(jsonTranscripts)
	}

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
