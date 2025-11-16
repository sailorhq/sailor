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
	"strconv"
)

type DeploySetting struct {
	K8s bool `json:"k8s"`
}

type SchemaSetting struct {
	Strict bool `json:"strict"`
}

type ResourceSetting struct {
	Deploy DeploySetting `json:"deploy"`
	Schema SchemaSetting `json:"schema"`
}

type SailorResource struct {
	Schema  map[string]any   `json:"schema"`
	Setting *ResourceSetting `json:"setting"`
}

type ResourceData struct {
	Version int    `json:"version"`
	Data    []byte `json:"data"`
}

// CreateResource creates a resource inside sailor instance
//
// TODO :: kind to be of type ResourceKind and token should be the last parameter
func (c *CoreAPIClient) CreateResource(ns, app, token, name string, kind string, setting ResourceSetting) error {
	// Construct the URL
	var url string

	switch kind {
	case "config", "secret":
		url = fmt.Sprintf("%s/api/v1/resource/%s/%s/%s", c.BaseURL, ns, app, kind)
	case "misc":
		if name == "" {
			return errors.New("misc resource must have a name")
		}
		url = fmt.Sprintf("%s/api/v1/resource/%s/%s/%s/%s", c.BaseURL, ns, app, kind, name)
	}

	b, err := json.Marshal(&setting)
	if err != nil {
		return err
	}

	// Create the PUT request
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(b))
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

// GetResource gets a resource inside sailor instance
//
// TODO :: kind to be of type ResourceKind and token should be the last parameter
func (c *CoreAPIClient) GetResource(ns, app, kind, name, token string, keyPair KeyPair) (*ResourceData, error) {
	// Construct the URL
	var url string

	switch kind {
	case "config", "secret":
		url = fmt.Sprintf("%s/api/v1/resource/%s/%s/%s", c.BaseURL, ns, app, kind)
	case "misc":
		if name == "" {
			return nil, errors.New("misc resource must have a name")
		}
		url = fmt.Sprintf("%s/api/v1/resource/%s/%s/%s/%s", c.BaseURL, ns, app, kind, name)
	}

	// Create the PUT request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("x-access-key", keyPair.AccessKey)
	req.Header.Set("x-secret-key", keyPair.SecretKey)
	req.Header.Set("x-token", token)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK && resp.Header.Get("x-resource-version") == "" {
		return nil, errors.New("resource version not found in response header")
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, serverMessageToErr(b)
	}

	version, err := strconv.Atoi(resp.Header.Get("x-resource-version"))
	if err != nil {
		return nil, err
	}

	return &ResourceData{
		Version: version,
		Data:    b,
	}, nil
}
