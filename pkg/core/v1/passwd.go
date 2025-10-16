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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func (c *CoreAPIClient) ChangePassword(user, pass, token string) error {
	// Construct the URL
	url := fmt.Sprintf("%s/api/v1/setting/passwd", c.BaseURL)

	// Create the PUT request
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("x-token", token)
	req.Header.Set("x-user", user)
	req.Header.Set("x-pass", pass)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return nil
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var loginResp map[string]string
	if err := json.Unmarshal(b, &loginResp); err != nil {
		return err
	}

	return errors.New(loginResp["message"])
}
