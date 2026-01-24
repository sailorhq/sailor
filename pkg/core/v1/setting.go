package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type OIDCSetting struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	IssuerURL    string   `json:"issuer_url"`
	Scopes       []string `json:"scopes"`
	RedirectURL  string   `json:"redirect_url"`
}

type SailorSetting struct {
	OIDC      *OIDCSetting   `json:"oidc"`
	TokenKey  string         `json:"-"`
	AccessKey string         `json:"accessKey"`
	SecretKey string         `json:"secretKey"`
	Manifest  SailorManifest `json:"manifest"`
	S3        *S3Setting     `json:"s3"`
	HostURL   string         `json:"hostURL"`
	Rxs       []RxSetting    `json:"rxs"`
}

type S3Setting struct {
	Bucket     string `json:"bucket"`
	Region     string `json:"region"`
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
	FolderPath string `json:"folderPath"`
}

type SignalFailurePolicy struct {
	OnDeploy string `json:"onDeploy"`
}
type PlugFailurePolicy struct {
	Signal SignalFailurePolicy `json:"signal"`
}
type RxSetting struct {
	Name          string            `json:"name"`
	Port          string            `json:"port"`
	Enabled       bool              `json:"enabled"`
	FailurePolicy PlugFailurePolicy `json:"failurePolicy"`
	BootConfig    map[string]any    `json:"bootConfig"`
}

func (c *CoreAPIClient) GetSailorSetting(token string) (*SailorSetting, error) {
	// Construct the URL
	url := fmt.Sprintf("%s/api/v1/setting", c.BaseURL)

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

	var ss SailorSetting
	if err := json.Unmarshal(b, &ss); err != nil {
		return nil, err
	}

	return &ss, nil
}

func (c *CoreAPIClient) UpdateSailorSetting(ss SailorSetting, token string) error {
	// Construct the URL
	url := fmt.Sprintf("%s/api/v1/setting", c.BaseURL)

	b, err := json.Marshal(&ss)
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

func (c *CoreAPIClient) GetResourceSetting(ns, app, kind, name, token string) (*ResourceSetting, error) {
	// Construct the URL
	var url string

	switch kind {
	case "config", "secret":
		url = fmt.Sprintf("%s/api/v1/resource/%s/%s/%s/setting", c.BaseURL, ns, app, kind)
	case "misc":
		if name == "" {
			return nil, errors.New("misc resource must have a name")
		}
		url = fmt.Sprintf("%s/api/v1/resource/%s/%s/%s/%s/setting", c.BaseURL, ns, app, kind, name)
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

	var rs ResourceSetting
	if err := json.Unmarshal(b, &rs); err != nil {
		return nil, err
	}

	return &rs, nil
}

func (c *CoreAPIClient) UpdateResourceSetting(rs ResourceSetting, ns, app, kind, name, token string) error {
	// Construct the URL
	var url string

	switch kind {
	case "config", "secret":
		url = fmt.Sprintf("%s/api/v1/resource/%s/%s/%s/setting", c.BaseURL, ns, app, kind)
	case "misc":
		if name == "" {
			return errors.New("misc resource must have a name")
		}
		url = fmt.Sprintf("%s/api/v1/resource/%s/%s/%s/%s/setting", c.BaseURL, ns, app, kind, name)
	}

	b, err := json.Marshal(&rs)
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
