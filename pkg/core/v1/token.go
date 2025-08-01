package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func (c *CoreAPIClient) GetToken(user string) (*TokenResponse, error) {
	// Construct the URL
	url := fmt.Sprintf("%s/api/v1/auth/token", c.BaseURL)

	// Create the PUT request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("x-user", user)
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

	if resp.StatusCode != 200 {
		var errMsg map[string]any
		if err := json.Unmarshal(b, &errMsg); err != nil {
			return nil, err
		}

		return nil, errors.New(errMsg["message"].(string))
	}

	var tokResp TokenResponse
	if err := json.Unmarshal(b, &tokResp); err != nil {
		return nil, err
	}

	return &tokResp, nil
}
