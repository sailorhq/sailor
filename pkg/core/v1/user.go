package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type CreateUserResponse struct {
	Note string `json:"note"`
	Pass string `json:"pass"`
}

func (c *CoreAPIClient) CreateUser(user, token string) (*CreateUserResponse, error) {
	// Construct the URL
	url := fmt.Sprintf("%s/api/v1/auth/user", c.BaseURL)

	b, err := json.Marshal(map[string]string{"email": user})
	if err != nil {
		return nil, err
	}

	// Create the PUT request
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(b))
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

	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, serverMessageToErr(b)
	}

	var cur CreateUserResponse
	if err := json.Unmarshal(b, &cur); err != nil {
		return nil, err
	}

	return &cur, nil
}
