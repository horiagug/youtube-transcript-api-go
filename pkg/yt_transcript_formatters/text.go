package yt_transcript_formatters

import (
	"fmt"
	"strings"

	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript_models"
)

type TextFormatter struct {
	BaseFormatter
}

func NewTextFormatter(options ...FormatterOption) *TextFormatter {
	f := &TextFormatter{
		BaseFormatter: BaseFormatter{
			IncludeTimestamps: true,
		},
	}

	for _, opt := range options {
		opt(&f.BaseFormatter)
	}

	return f
}

func (t *TextFormatter) Format(transcripts []yt_transcript_models.Transcript) (string, error) {

	var (
		text strings.Builder
		err  error
	)

	for _, transcript := range transcripts {
		_, err = text.WriteString(fmt.Sprintf("Language: %s\n", transcript.Language))
		if err != nil {
			return "", err
		}

		for _, line := range transcript.Lines {
			if t.IncludeTimestamps {
				_, err = text.WriteString(fmt.Sprintf("%f: %s\n", line.Start, line.Text))
			} else {
				_, err = text.WriteString(line.Text + "\n")
			}
		}
	}

	if err != nil {
		return "", err
	}

	return text.String(), nil
}
