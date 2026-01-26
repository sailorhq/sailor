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

// RegisterResourceRoutes registers all resource-related routes
func RegisterResourceRoutes(apiV1 *router.Group, core *handlers.SailorCore) {
	// Create resource
	createResourceHandler := core.Authenticated(core.CreateResourceHandler, v1.RBACConstraints{
		Roles:       []string{RoleUser},
		Permissions: []string{PermissionCreateResource},
	})
	apiV1.PUT("/resource/{namespace}/{app}/{kind}", createResourceHandler)
	apiV1.PUT("/resource/{namespace}/{app}/{kind}/{name}", createResourceHandler)

	// Get resource
	apiV1.GET("/resource/{namespace}/{app}/{kind}", core.ClientCallable(core.GetResourceHandler))
	apiV1.GET("/resource/{namespace}/{app}/{kind}/{name}", core.ClientCallable(core.GetResourceHandler))

	// Create deployment
	createDeploymentHandler := core.Authenticated(core.CreateDeploymentHandler, v1.RBACConstraints{
		Roles:       []string{RoleUser},
		Permissions: []string{PermissionCreateDeployment},
	})
	apiV1.PUT("/resource/{namespace}/{app}/{kind}/deployment", createDeploymentHandler)
	apiV1.PUT("/resource/{namespace}/{app}/{kind}/{name}/deployment", createDeploymentHandler)

	// Get deployment
	getDeploymentHandler := core.Authenticated(core.GetDeploymentHandler, v1.RBACConstraints{
		Roles:       []string{RoleUser},
		Permissions: []string{PermissionCreateDeployment},
	})
	apiV1.GET("/resource/{namespace}/{app}/{kind}/deployment", getDeploymentHandler)
	apiV1.GET("/resource/{namespace}/{app}/{kind}/{name}/deployment", getDeploymentHandler)

	// Deploy resource
	deployResourceHandler := core.Authenticated(core.DeployResourceHandler, v1.RBACConstraints{
		Roles:       []string{RoleUser},
		Permissions: []string{PermissionDeploy},
	})
	apiV1.POST("/resource/{namespace}/{app}/{kind}/deploy", deployResourceHandler)
	apiV1.POST("/resource/{namespace}/{app}/{kind}/{name}/deploy", deployResourceHandler)

	// Update resource setting
	updateResourceSetting := core.Authenticated(core.UpdateResourceSetting, v1.RBACConstraints{
		Roles:       []string{RoleUser},
		Permissions: []string{PermissionEditSetting},
	})
	apiV1.POST("/resource/{namespace}/{app}/{kind}/setting", updateResourceSetting)
	apiV1.POST("/resource/{namespace}/{app}/{kind}/{name}/setting", updateResourceSetting)

	// Get resource setting
	getResourceSetting := core.Authenticated(core.GetResourceSetting, v1.RBACConstraints{
		Roles:       []string{RoleUser},
		Permissions: []string{PermissionViewSetting},
	})
	apiV1.GET("/resource/{namespace}/{app}/{kind}/setting", getResourceSetting)
	apiV1.GET("/resource/{namespace}/{app}/{kind}/{name}/setting", getResourceSetting)

	// Update resource schema
	updateResourceSchemaHandler := core.Authenticated(core.UpdateResourceSchemaHandler, v1.RBACConstraints{
		Roles:       []string{RoleUser},
		Permissions: []string{PermissionEditSchema},
	})
	apiV1.POST("/resource/{namespace}/{app}/{kind}/schema", updateResourceSchemaHandler)
	apiV1.POST("/resource/{namespace}/{app}/{kind}/{name}/schema", updateResourceSchemaHandler)

	// Get resource schema
	getResourceSchemaHandler := core.Authenticated(core.GetResourceSchemaHandler, v1.RBACConstraints{
		Roles:       []string{RoleUser},
		Permissions: []string{PermissionViewSchema},
	})
	apiV1.GET("/resource/{namespace}/{app}/{kind}/schema", getResourceSchemaHandler)
	apiV1.GET("/resource/{namespace}/{app}/{kind}/{name}/schema", getResourceSchemaHandler)
}
