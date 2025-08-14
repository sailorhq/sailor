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
package types

import (
	"time"
)

type SailorMeta struct {
	Version string `json:"version"`
}

type SailorOpts struct {
	Logging        bool
	DisableRefresh bool
	RefreshTimeout time.Duration
	BackupURL      string
	AccessKey      string
	SecretKey      string
}

type Deployment struct {
	Description string `json:"description"`
	Version     string `json:"version"`
	Deployed    bool   `json:"deployed"`
	Diff        string `json:"diff"`
	CreatedAt   string `json:"created_at"`
	CreatedBy   string `json:"created_by"`
	Data        []byte `json:"data"`
}

type Secret struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type AdminBackupState struct {
	Bucket string `json:"bucket"`
}

type ListAppsResponse struct {
	Apps             []string          `json:"apps"`
	AdminBackupState *AdminBackupState `json:"admin_backup_state"`
}

type AdminSailorState struct {
	Meta        SailorMeta     `json:"meta"`
	Configs     map[string]any `json:"configs"`
	Secrets     []Secret       `json:"secrets"`
	AccessKey   string         `json:"access_key"`
	SecretKey   string         `json:"secret_key"`
	Rules       string         `json:"rules"`
	Policy      string         `json:"policy"`
	Deployments []Deployment   `json:"deployments"`
}

type SailorState struct {
	Version string            `json:"config_ver"`
	Config  []byte            `json:"config"`
	Secrets map[string][]byte `json:"secrets"`
}
