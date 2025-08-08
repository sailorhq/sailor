package v1

import (
	"bytes"
	"encoding/json"
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

type Webhook struct {
	OnOIDCSuccess string `json:"on_oidc_success"`
}
type SailorSetting struct {
	OIDC     *OIDCSetting   `json:"oidc"`
	TokenKey string         `json:"-"`
	Webhook  Webhook        `json:"webhook"`
	Manifest SailorManifest `json:"manifest"`
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
