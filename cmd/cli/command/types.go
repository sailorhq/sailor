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
package command

import (
	"github.com/sailorhq/sailor/internal/types"
	v1 "github.com/sailorhq/sailor/pkg/core/v1"
)

type CLIConfig struct {
	// Manifest is used to know details about different environment sailor is hosted in
	Manifest v1.SailorManifest `json:"manifest"`

	// SailorRoot is mostly ~/.sailor
	SailorRoot string `json:"-"`

	// Env is the current selected environment by the user
	Env string `json:"env"`

	// SailorHost is the host of the current selected environment, splatted for ease
	// of use
	SailorHost string `json:"-"`

	// SailorClient is the REST API client created with SailorHost
	SailorClient *v1.CoreAPIClient `json:"-"`

	// Token is the admin/user token fetched after logged in, it works until
	// it expires!
	Token string `json:"token"`

	// User is used for global set an email which can then be used for logging
	// in to your sailor core server
	User string `json:"user"`

	// CwdSailorFile is the sailor file in the current working directory
	CwdSailorFile types.SailorFile `json:"-"`

	// CwdSailorLockFile is the sailor lock file in the current working directory
	CwdSailorLockFile types.SailorLockFile `json:"-"`
}
