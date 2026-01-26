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

// RegisterAuthRoutes registers all authentication-related routes
func RegisterAuthRoutes(apiV1 *router.Group, core *handlers.SailorCore) {
	apiV1.GET("/auth/oidc", core.AuthOIDCHandler)
	apiV1.ANY("/auth/callback", core.AuthCallbackHandler)
	apiV1.GET("/auth/token", core.GetTokenHandler)
	apiV1.GET("/keypair/{namespace}/{app}", core.ClientCallable(core.GetKeyPairHandler))

	apiV1.POST("/auth/basic", core.AuthBasicHandler)
	apiV1.POST("/auth/rbac", core.Authenticated(core.AuthRBACHandler, v1.RBACConstraints{
		Roles:       []string{RoleAdmin},
		Permissions: []string{PermissionSuperAdmin},
	}))

	apiV1.PUT("/auth/user", core.Authenticated(core.CreateUserHandler, v1.RBACConstraints{
		Roles:       []string{RoleAdmin},
		Permissions: []string{PermissionSuperAdmin},
	}))
	apiV1.GET("/auth/user/{user}", core.Authenticated(core.GetUserHandler, v1.RBACConstraints{
		Roles:       []string{RoleAdmin},
		Permissions: []string{PermissionSuperAdmin},
	}))
}
