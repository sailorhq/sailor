package v1

import (
	"net/http"
	"time"
)

// HTTPClient interface to make the client testable
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type CoreAPIClient struct {
	BaseURL string
	Client  HTTPClient
}

// CoreV1 creates a new API client instance
func CoreV1(baseURL string) *CoreAPIClient {
	return &CoreAPIClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}
