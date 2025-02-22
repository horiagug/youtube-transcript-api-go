package errors

type TranscriptError string

func (e TranscriptError) Error() string {
	return string(e)
}

const (
	ErrNoTranscript    = TranscriptError("no transcript found")
	ErrInvalidVideoID  = TranscriptError("invalid video ID")
	ErrTooManyRequests = TranscriptError("too many requests")
)
