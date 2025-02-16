package client

import (
	"context"
	"fmt"
	"time"

	"github.com/horiagug/youtube-transcript-api-go/pkg/formatters"
	"github.com/horiagug/youtube-transcript-api-go/pkg/repository"
	"github.com/horiagug/youtube-transcript-api-go/pkg/service"
)

type Client struct {
	transcriptService *service.TranscriptService
	timeout           int
	formatter         formatters.Formatter
}

var preserve_formatting_default = false

func New(options ...Option) *Client {
	// Set default values
	formatter := formatters.NewJSONFormatter()
	formatter.Configure(formatters.WithPrettyPrint(true))
	client := &Client{
		timeout:   30,
		formatter: formatter,
	}

	for _, opt := range options {
		opt(client)
	}

	if client.transcriptService == nil {
		fetcher := repository.NewHTMLFetcher()
		client.transcriptService = service.NewTranscriptService(fetcher)
	}

	return client
}

func (c *Client) GetTranscript(videoID string, languages []string, preserve_formatting bool) (string, error) {
	_, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeout)*time.Second)
	defer cancel()

	transcripts, err := c.transcriptService.GetTranscripts(videoID, languages, preserve_formatting)
	if err != nil {
		return "", err
	}

	if len(transcripts) == 0 {
		return "", fmt.Errorf("No transcripts found")
	}

	return c.formatter.Format(transcripts)

}
