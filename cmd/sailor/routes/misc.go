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

// RegisterAuditRoutes registers audit-related routes
func RegisterAuditRoutes(apiV1 *router.Group, core *handlers.SailorCore) {
	apiV1.GET("/audit", core.Authenticated(core.GetAuditEvents, v1.RBACConstraints{
		Roles: []string{RoleAdmin},
	}))
}

// RegisterK8sHookRoutes registers Kubernetes webhook routes
func RegisterK8sHookRoutes(hooksK8s *router.Group, core *handlers.SailorCore) {
	hooksK8s.POST("/admission", core.K8sAdmissionHookHandler)
}
