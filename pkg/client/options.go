package client

import (
	"github.com/horiagug/youtube-transcript-api-go/pkg/formatters"
	"github.com/horiagug/youtube-transcript-api-go/pkg/service"
)

type Option func(*Client)

func WithCustomFetcher(fetcher service.HTMLFetcherType) Option {
	return func(c *Client) {
		c.transcriptService = service.NewTranscriptService(fetcher)
	}
}

func WithTimeout(seconds int) Option {
	return func(c *Client) {
		c.timeout = seconds
	}
}
func WithFormatter(formatter formatters.Formatter) Option {
	return func(c *Client) {
		c.formatter = formatter
	}
}
