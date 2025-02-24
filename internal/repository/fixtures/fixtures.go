package fixtures

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

// MockHTMLFetcher implements HTMLFetcherType for testing
type MockHTMLFetcher struct {
	mock.Mock
}

func (m *MockHTMLFetcher) Fetch(url string, cookie *http.Cookie) ([]byte, error) {
	args := m.Called(url, cookie)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockHTMLFetcher) FetchVideo(videoID string) ([]byte, error) {
	args := m.Called(videoID)
	return args.Get(0).([]byte), args.Error(1)
}
