package yt_transcript_formatters

import (
	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript_models"
)

type Formatter interface {
	Format(transcripts []yt_transcript_models.Transcript) (string, error)
}

type BaseFormatter struct {
	IncludeTimestamps bool
}

type FormatterOption func(f *BaseFormatter)

func WithTimestamps(include bool) FormatterOption {
	return func(f *BaseFormatter) {
		f.IncludeTimestamps = include
	}
}
