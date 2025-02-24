package repository

import (
	"encoding/xml"
	"html"
	"regexp"
	"strconv"
	"strings"

	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript_models"
)

// transcriptParser struct handles transcript parsing
type transcriptParser struct {
	htmlRegex *regexp.Regexp
}

// Formatting tags to preserve
var formattingTags = []string{
	"strong", "em", "b", "i", "mark", "small", "del", "ins", "sub", "sup",
}

var htmlRegex = regexp.MustCompile(`(?i)<[^>]*>`)

// NewTranscriptParser initializes the parser with or without preserving formatting tags
func NewTranscriptParser(preserveFormatting bool) *transcriptParser {
	htmlRegex := getHTMLRegex(preserveFormatting)
	return &transcriptParser{htmlRegex: htmlRegex}
}

// getHTMLRegex returns a regex pattern for removing unwanted HTML tags
func getHTMLRegex(preserveFormatting bool) *regexp.Regexp {
	if preserveFormatting {
		// Match tags that are NOT in the allowed list
		formatsRegex := `</?(?:` + strings.Join(formattingTags, "|") + `)\b[^>]*>`
		return regexp.MustCompile(`(?i)<[^>]*>(?:(?i)` + formatsRegex + `)?`)
	}
	// Remove all HTML tags
	return regexp.MustCompile(`(?i)<[^>]*>`)
}

func cleanHTML(text string, preserveFormatting bool) string {
	// Remove all HTML tags
	cleaned := htmlRegex.ReplaceAllString(text, "")

	if preserveFormatting {
		// Manually re-add allowed tags (approximation)
		for _, tag := range formattingTags {
			cleaned = strings.ReplaceAll(cleaned, "&lt;"+tag+"&gt;", "<"+tag+">")
			cleaned = strings.ReplaceAll(cleaned, "&lt;/"+tag+"&gt;", "</"+tag+">")
		}
	}

	return cleaned
}

// Parse extracts transcript text, start time, and duration from XML
func (p *transcriptParser) Parse(plainData string) ([]yt_transcript_models.TranscriptLine, error) {
	type XMLTranscript struct {
		XMLName xml.Name `xml:"transcript"`
		Texts   []struct {
			Text     string `xml:",chardata"`
			Start    string `xml:"start,attr"`
			Duration string `xml:"dur,attr"`
		} `xml:"text"`
	}

	var parsedXML XMLTranscript
	err := xml.Unmarshal([]byte(plainData), &parsedXML)
	if err != nil {
		return nil, err
	}

	var results []yt_transcript_models.TranscriptLine
	for _, entry := range parsedXML.Texts {
		// First clean HTML, then unescape HTML entities
		text := cleanHTML(entry.Text, false)
		text = html.UnescapeString(text)

		start, err := strconv.ParseFloat(entry.Start, 64)
		if err != nil {
			start = 0.0
		}

		duration, err := strconv.ParseFloat(entry.Duration, 64)
		if err != nil {
			duration = 0.0
		}

		results = append(results, yt_transcript_models.TranscriptLine{
			Text:     text,
			Start:    start,
			Duration: duration,
		})
	}
	return results, nil
}
