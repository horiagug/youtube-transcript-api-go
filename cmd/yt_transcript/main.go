package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript"
	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript_formatters"
)

func main() {
	var (
		languages                = flag.String("languages", "en", "Comma-separated list of language codes")
		formatter                = flag.String("formatter", "json", "Formatter to use (json, text)")
		preserve_formatting      = flag.Bool("preserve_formatting", true, "Preserve formatting")
		with_timestamps          = flag.Bool("with_timestamps", true, "Include timestamps")
		with_language_code       = flag.Bool("with_language_code", true, "Include language code")
		exclude_manually_created = flag.Bool("exclude_manually_created", false, "Exclude manually created subtitles") // not in use yet
		exclude_auto_generated   = flag.Bool("exclude_auto_generated", false, "Exclude auto-generated subtitles")     // not in use yet
	)
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Please provide at least one video ID")
		os.Exit(1)
	}

	if *exclude_manually_created && *exclude_auto_generated {
		fmt.Println("Cannot exclude both manually created and auto-generated subtitles")
		os.Exit(1)
	}

	var outputFormatter yt_transcript_formatters.Formatter

	if *formatter == "text" {
		outputFormatter = yt_transcript_formatters.NewTextFormatter(
			yt_transcript_formatters.WithTimestamps(*with_timestamps),
			yt_transcript_formatters.WithLanguageCode(*with_language_code),
		)
	} else {
		outputFormatter = yt_transcript_formatters.NewJSONFormatter(
			yt_transcript_formatters.WithTimestamps(*with_timestamps),
			yt_transcript_formatters.WithLanguageCode(*with_language_code),
		)
	}

	client := yt_transcript.NewClient(
		yt_transcript.WithFormatter(outputFormatter),
	)

	videoID := flag.Arg(0)
	t, err := client.GetFormattedTranscripts(videoID, []string{*languages}, *preserve_formatting)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s", t)
	os.Exit(0)
}
