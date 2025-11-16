// sailor
// Copyright (C) 2025 SailorHQ and Ashish Shekar (codekidX)

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type DeploymentResponse struct {
	Deployments []Deployment `json:"deployments"`
}

type Deployment struct {
	Version     string `json:"version"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

func (c *CoreAPIClient) GetDeployment(ns, app, kind, name, token, version string) (*DeploymentResponse, error) {
	// Construct the URL
	var url string

	switch kind {
	case "config", "secret":
		url = fmt.Sprintf("%s/api/v1/resource/%s/%s/%s/deployment", c.BaseURL, ns, app, kind)
	case "misc":
		if name == "" {
			return nil, errors.New("misc resource must have a name")
		}
		url = fmt.Sprintf("%s/api/v1/resource/%s/%s/%s/%s/deployment", c.BaseURL, ns, app, kind, name)
	}

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

	var depResp DeploymentResponse
	if err := json.Unmarshal(b, &depResp); err != nil {
		return nil, err
	}

	return &depResp, nil
}

func (c *CoreAPIClient) CreateDeployment(ns, app, kind, name, token, desc string, localVersion int, data any) (*any, error) {
	// Construct the URL
	var url string
	dataKey := fmt.Sprintf("%s_data", kind)
	switch kind {
	case "config", "secret":
		url = fmt.Sprintf("%s/api/v1/resource/%s/%s/%s/deployment", c.BaseURL, ns, app, kind)
	case "misc":
		if name == "" {
			return nil, errors.New("misc resource must have a name")
		}
		url = fmt.Sprintf("%s/api/v1/resource/%s/%s/%s/%s/deployment", c.BaseURL, ns, app, kind, name)
	}

	b, err := json.Marshal(&map[string]any{
		dataKey:         data,
		"desc":          desc,
		"local_version": localVersion,
	})
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

	var dataResp map[string]any
	if err := json.Unmarshal(b, &dataResp); err != nil {
		return nil, err
	}

	version := dataResp["version"]

	return &version, nil
}

func (c *CoreAPIClient) Deploy(ns, app, kind, name, token, version string) error {
	// Construct the URL
	var url string
	switch kind {
	case "config", "secret":
		url = fmt.Sprintf("%s/api/v1/resource/%s/%s/%s/deploy", c.BaseURL, ns, app, kind)
	case "misc":
		if name == "" {
			return errors.New("misc resource must have a name")
		}
		url = fmt.Sprintf("%s/api/v1/resource/%s/%s/%s/%s/deploy", c.BaseURL, ns, app, kind, name)
	}

	b, err := json.Marshal(&map[string]any{
		"version": version,
	})
	if err != nil {
		return err
	}

	// Create the PUT request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("x-token", token)
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
