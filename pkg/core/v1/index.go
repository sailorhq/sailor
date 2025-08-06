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
	"net/http"
	"time"
)

// HTTPClient interface to make the client testable
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type CoreAPIClient struct {
	BaseURL string
	Client  HTTPClient
}

// CoreV1 creates a new API client instance
func CoreV1(baseURL string) *CoreAPIClient {
	return &CoreAPIClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}
