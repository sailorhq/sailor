package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type TokenResponse struct {
	Token string `json:"token"`
}

func (c *CoreAPIClient) LoginBasic(user, pass string) (*TokenResponse, error) {
	// Construct the URL
	url := fmt.Sprintf("%s/api/v1/auth/basic", c.BaseURL)

	// Create the PUT request
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("x-user", user)
	req.Header.Set("x-pass", pass)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokResp TokenResponse
	if err := json.Unmarshal(b, &tokResp); err != nil {
		return nil, err
	}

	return &tokResp, nil
}
