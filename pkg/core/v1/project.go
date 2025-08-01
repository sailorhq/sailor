package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ProjectResponse struct {
	Key string `json:"key"`
}

func (c *CoreAPIClient) CreateProject(namespace, app, token string) (*ProjectResponse, error) {
	// Construct the URL
	url := fmt.Sprintf("%s/api/v1/project/%s/%s", c.BaseURL, namespace, app)

	// Create the PUT request
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("x-token", token)
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
		return nil, serverMessageToErr(b)
	}

	var projectResp ProjectResponse
	if err := json.Unmarshal(b, &projectResp); err != nil {
		return nil, err
	}

	return &projectResp, nil
}
