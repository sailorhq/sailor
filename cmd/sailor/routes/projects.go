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

import (
	"github.com/fasthttp/router"
	"github.com/sailorhq/sailor/cmd/sailor/handlers"
	v1 "github.com/sailorhq/sailor/pkg/core/v1"
)

// RegisterProjectRoutes registers all project-related routes
func RegisterProjectRoutes(apiV1 *router.Group, core *handlers.SailorCore) {
	apiV1.PUT("/project/{namespace}/{app}", core.Authenticated(core.CreateProjectHandler, v1.RBACConstraints{
		Roles:       []string{RoleAdmin, RoleUser},
		Permissions: []string{PermissionSuperAdmin, PermissionCreateProject},
	}))
	apiV1.GET("/projects", core.Authenticated(core.GetProjects, v1.RBACConstraints{
		Roles: []string{RoleAdmin, RoleUser},
	}))
}
