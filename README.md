# YouTube Transcript API Go

A Go library and CLI tool to get transcripts/subtitles from YouTube videos. This library supports multiple languages, different output formats, and various formatting options.

## Features

- Fetch transcripts from YouTube videos
- Support for multiple languages
- JSON and Text output formats
- Concurrent processing of transcripts
- Preserve or strip formatting
- Include/exclude timestamps
- Support for both auto-generated and manually created subtitles

## Installation

### As a CLI Tool

```bash
# Install the CLI tool
go install github.com/horiagug/youtube-transcript-api-go@latest
```

### As a Library

```bash
# Add to your Go project
go get github.com/horiagug/youtube-transcript-api-go
```

## CLI Usage

```bash
# Basic usage
youtube-transcript-api-go [flags] VIDEO_ID

# Flags:
  -languages string
        Comma-separated list of language codes (default "en")
  -formatter string
        Formatter to use (json, text) (default "json")
  -preserve_formatting
        Preserve formatting (default true)
  -with_timestamps
        Include timestamps (default true)
  -exclude_manually_created
        Exclude manually created subtitles
  -exclude_auto_generated
        Exclude auto-generated subtitles
```

### Examples

```bash
# Get English transcripts in JSON format
youtube-transcript-api-go dQw4w9WgXcQ

# Get Spanish transcripts in text format
youtube-transcript-api-go -languages es -formatter text dQw4w9WgXcQ

# Get transcripts without timestamps
youtube-transcript-api-go -with_timestamps=false dQw4w9WgXcQ
```

## Library Usage

```go
package main

import (
    "fmt"
    "github.com/horiagug/youtube-transcript-api-go/pkg/client"
    "github.com/horiagug/youtube-transcript-api-go/pkg/formatters"
)

func main() {
    // Create a new client with JSON formatter
    client := client.NewClient(
        client.WithFormatter(formatters.NewJSONFormatter()),
    )

    // Get formatted transcripts
    videoID := "dQw4w9WgXcQ"
    languages := []string{"en"}
    transcript, err := client.GetFormattedTranscripts(videoID, languages, true)
    if err != nil {
        panic(err)
    }

    fmt.Println(transcript)

    // Or get raw transcript data
    transcripts, err := client.GetTranscripts(videoID, languages)
    if err != nil {
        panic(err)
    }

    // Process transcripts as needed
    for _, t := range transcripts {
        fmt.Printf("Language: %s\n", t.Language)
        for _, line := range t.Lines {
            fmt.Printf("%s: %s\n", line.Start, line.Text)
        }
    }
}
```

## Custom Formatting

The library supports both JSON and Text formatters with configurable options:

```go
// JSON formatter with custom options
jsonFormatter := formatters.NewJSONFormatter(
    formatters.WithPrettyPrint(true),
    formatters.WithTimestamps(true),
)

// Text formatter with custom options
textFormatter := formatters.NewTextFormatter(
    formatters.WithTimestamps(true),
)

// Use formatter with client
client := client.NewClient(
    client.WithFormatter(jsonFormatter),
)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
