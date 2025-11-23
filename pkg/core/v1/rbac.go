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
	"fmt"
	"io"
	"net/http"
)

type RBACConstraints struct {
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
	AllowedApps []string `json:"allowed_apps"`
}

type RBACRequest struct {
	Addition RBACConstraints `json:"add"`
	Deletion RBACConstraints `json:"del"`
}

func (c *CoreAPIClient) UpdateRBAC(rbacReq RBACRequest, user string) error {
	// Construct the URL
	url := fmt.Sprintf("%s/api/v1/auth/rbac?user=%s", c.BaseURL, user)

	b, err := json.Marshal(&rbacReq)
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

	var projectResp ProjectResponse
	if err := json.Unmarshal(b, &projectResp); err != nil {
		return err
	}

	return nil
}
