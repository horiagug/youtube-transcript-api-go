package formatters

import (
	"github.com/horiagug/youtube-transcript-api-go/pkg/models"
)

// Formatter defines the interface for transcript formatters
type Formatter interface {
	// Format converts transcript lines into a specific format
	Format(transcripts []models.Transcript) (string, error)
}

// BaseFormatter contains common formatting utilities
type BaseFormatter struct {
	// Common configuration options could go here
	IncludeTimestamps bool
}

// FormatterOption is a function type for formatter configuration
type FormatterOption func(f *BaseFormatter)

// WithTimestamps configures whether timestamps should be included
func WithTimestamps(include bool) FormatterOption {
	return func(f *BaseFormatter) {
		f.IncludeTimestamps = include
	}
}
