package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AuditEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Username  string    `json:"username"`
	Namespace string    `json:"namespace,omitempty"`
	App       string    `json:"app,omitempty"`
	Action    string    `json:"action"`
	Details   any       `json:"details,omitempty"`
}

func (c *CoreAPIClient) GetAuditLogEvents(limit int, token string) (*[]AuditEvent, error) {
	// Construct the URL
	url := fmt.Sprintf("%s/api/v1/audit", c.BaseURL)

	// Create the PUT request
	req, err := http.NewRequest("GET", url, nil)
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

	var events []AuditEvent
	if err := json.Unmarshal(b, &events); err != nil {
		return nil, err
	}

	return &events, nil
}
