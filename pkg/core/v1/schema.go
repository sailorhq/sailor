package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func (c *CoreAPIClient) GetSchema(ns, app, kind, name string) (*map[string]any, error) {
	// Construct the URL
	var url string

	switch kind {
	case "config", "secret":
		url = fmt.Sprintf("%s/api/v1/resource/%s/%s/%s/schema", c.BaseURL, ns, app, kind)
	case "misc":
		if name == "" {
			return nil, errors.New("misc resource must have a name")
		}
		url = fmt.Sprintf("%s/api/v1/resource/%s/%s/%s/%s/schema", c.BaseURL, ns, app, kind, name)
	}

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

	var schema map[string]any
	if err := json.Unmarshal(b, &schema); err != nil {
		return nil, err
	}

	return &schema, nil
}

func (c *CoreAPIClient) UpdateSchema(schema map[string]any, ns, app, kind, name string) error {
	// Construct the URL
	var url string

	switch kind {
	case "config", "secret":
		url = fmt.Sprintf("%s/api/v1/resource/%s/%s/%s/schema", c.BaseURL, ns, app, kind)
	case "misc":
		if name == "" {
			return errors.New("misc resource must have a name")
		}
		url = fmt.Sprintf("%s/api/v1/resource/%s/%s/%s/%s/schema", c.BaseURL, ns, app, kind, name)
	}

	b, err := json.Marshal(schema)
	if err != nil {
		return err
	}

	// Create the PUT request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("x-token", c.Token)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return serverMessageToErr(b)
	}

	return nil
}
