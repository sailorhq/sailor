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
)

// EnvConfig has name
type EnvConfig struct {
	Name string `json:"name"`
	Host string `json:"host"`
}

// SailorManifest is the manifest with which Sailor is configured
// Right now we only have environment mentioned here
type SailorManifest struct {
	Envs []EnvConfig `json:"envs"`
}

func serverMessageToErr(b []byte) error {
	var errMsg map[string]string
	if err := json.Unmarshal(b, &errMsg); err != nil {
		return err
	}
	return errors.New(errMsg["message"])
}
