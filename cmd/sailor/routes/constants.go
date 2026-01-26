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
package routes

// Role constants
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// Permission constants
const (
	PermissionSuperAdmin       = "super:*"
	PermissionCreateProject    = "create_proj"
	PermissionCreateResource   = "create_res"
	PermissionCreateDeployment = "create_dep"
	PermissionDeploy           = "deploy"
	PermissionViewSetting      = "view_setting"
	PermissionEditSetting      = "edit_setting"
	PermissionViewSchema       = "view_schema"
	PermissionEditSchema       = "edit_schema"
)
