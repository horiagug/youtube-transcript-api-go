package client

import (
	"context"
	"fmt"
	"time"

	"github.com/horiagug/youtube-transcript-api-go/pkg/formatters"
	"github.com/horiagug/youtube-transcript-api-go/pkg/models"
	"github.com/horiagug/youtube-transcript-api-go/pkg/repository"
	"github.com/horiagug/youtube-transcript-api-go/pkg/service"
)

type Client struct {
	transcriptService *service.TranscriptService
	Timeout           int
	Formatter         formatters.Formatter
}

var preserve_formatting_default = false

func NewClient(options ...Option) *Client {

	// Set default values
	formatter := formatters.NewJSONFormatter()
	formatter.Configure(formatters.WithPrettyPrint(true))

	client := &Client{
		Timeout:   30,
		Formatter: formatter,
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

func (c *Client) GetFormattedTranscripts(videoID string, languages []string, preserve_formatting bool) (string, error) {
	_, cancel := context.WithTimeout(context.Background(), time.Duration(c.Timeout)*time.Second)
	defer cancel()

	transcripts, err := c.transcriptService.GetTranscripts(videoID, languages, preserve_formatting)
	if err != nil {
		return "", err
	}

	if len(transcripts) == 0 {
		return "", fmt.Errorf("No transcripts found")
	}

	return c.Formatter.Format(transcripts)
}

func (c *Client) GetTranscripts(videoID string, languages []string) ([]models.Transcript, error) {
	_, cancel := context.WithTimeout(context.Background(), time.Duration(c.Timeout)*time.Second)
	defer cancel()

	transcripts, err := c.transcriptService.GetTranscripts(videoID, languages, true)
	if err != nil {
		return []models.Transcript{}, err
	}

	return transcripts, nil
}
