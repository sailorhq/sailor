package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
