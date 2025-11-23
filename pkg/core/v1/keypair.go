package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *CoreAPIClient) GetKeyPair(ns, app string) (*KeyPair, error) {
	// Construct the URL
	url := fmt.Sprintf("%s/api/v1/keypair/%s/%s", c.BaseURL, ns, app)

	// Create the PUT request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("x-token", c.Token)
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

	var keyPair KeyPair
	if err := json.Unmarshal(b, &keyPair); err != nil {
		return nil, err
	}

	return &keyPair, nil
}
