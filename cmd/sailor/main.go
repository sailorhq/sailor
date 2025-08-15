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
package main

import (
	"os"

	"github.com/sailorhq/sailor/cmd/sailor/handlers"
	v1 "github.com/sailorhq/sailor/pkg/core/v1"
	"go.uber.org/zap"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

// var staticConsoleFS embed.FS

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

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

func main() {

	core := handlers.NewSailorCore()

	if core == nil {
		panic("core of sailor was unable to start! check for errors...")
	}

	core.Log.Info("sailor core initialized")

	r := router.New()
	apiV1 := r.Group("/api/v1")

	core.Log.Info("initializing core routes")

	// CORE
	//
	// this block contains APIs which is responsible of core working of sailor
	// like obtaining a sailor token or providing the access key and secret key
	// to an authorized party!
	apiV1.GET("/setting", core.Authenticated(core.GetSailorSettingHandler, v1.RBACConstraints{
		Roles:       []string{RoleAdmin},
		Permissions: []string{PermissionSuperAdmin},
	}))

	apiV1.POST("/setting", core.Authenticated(core.SailorSettingHandler, v1.RBACConstraints{
		Roles:       []string{RoleAdmin},
		Permissions: []string{PermissionSuperAdmin},
	}))

	apiV1.POST("/setting/manifest", core.Authenticated(core.UpdateManifestHandler, v1.RBACConstraints{
		Roles:       []string{RoleAdmin},
		Permissions: []string{PermissionSuperAdmin},
	}))

	apiV1.GET("/setting/manifest", core.GetManifestHandler)

	// AUTH
	//
	// this block contains APIs which uses OIDC module to authenticate user and fetch a
	// outh2 valid token
	apiV1.GET("/auth/oidc", core.AuthOIDCHandler)
	apiV1.ANY("/auth/callback", core.AuthCallbackHandler)
	apiV1.GET("/auth/token", core.GetTokenHandler)

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

	// PROJECT
	//
	// this block contains APIs that work on project. A project is a collection of
	// resources, and a project should contain a namespace and an app value.
	apiV1.PUT("/project/{namespace}/{app}", core.Authenticated(core.CreateProjectHandler, v1.RBACConstraints{
		Roles:       []string{RoleAdmin, RoleUser},
		Permissions: []string{PermissionSuperAdmin, PermissionCreateProject},
	}))

	// RESOURCE
	//
	// this block contains APIs that work and operate on resources inside a project.
	// These APIs handles three types of resources:
	//
	// config - a key value pair with value as {any} type
	// secret - a key value pair with value as string and is encrypted
	// misc   - a text value which cannot be interpretted at the moment, but the
	//          user knows how to make sense of it.
	createResourceHandler := core.Authenticated(core.CreateResourceHandler, v1.RBACConstraints{
		Roles:       []string{RoleUser},
		Permissions: []string{PermissionCreateResource},
	})
	apiV1.PUT("/resource/{namespace}/{app}/{kind}", createResourceHandler)
	apiV1.PUT("/resource/{namespace}/{app}/{kind}/{name}", createResourceHandler)

	apiV1.GET("/resource/{namespace}/{app}/{kind}", core.ClientCallable(core.GetResourceHandler))
	apiV1.GET("/resource/{namespace}/{app}/{kind}/{name}", core.ClientCallable(core.GetResourceHandler))

	createDeploymentHandler := core.Authenticated(core.CreateDeploymentHandler, v1.RBACConstraints{
		Roles:       []string{RoleUser},
		Permissions: []string{PermissionCreateDeployment},
	})
	apiV1.PUT("/resource/{namespace}/{app}/{kind}/deployment", createDeploymentHandler)
	apiV1.PUT("/resource/{namespace}/{app}/{kind}/{name}/deployment", createDeploymentHandler)

	getDeploymentHandler := core.Authenticated(core.GetDeploymentHandler, v1.RBACConstraints{
		Roles:       []string{RoleUser},
		Permissions: []string{PermissionCreateDeployment},
	})
	apiV1.GET("/resource/{namespace}/{app}/{kind}/deployment", getDeploymentHandler)
	apiV1.GET("/resource/{namespace}/{app}/{kind}/{name}/deployment", getDeploymentHandler)

	deployResourceHandler := core.Authenticated(core.DeployResourceHandler, v1.RBACConstraints{
		Roles:       []string{RoleUser},
		Permissions: []string{PermissionDeploy},
	})
	apiV1.POST("/resource/{namespace}/{app}/{kind}/deploy", deployResourceHandler)
	apiV1.POST("/resource/{namespace}/{app}/{kind}/{name}/deploy", deployResourceHandler)

	updateResourceSetting := core.Authenticated(core.UpdateResourceSetting, v1.RBACConstraints{
		Roles:       []string{RoleUser},
		Permissions: []string{PermissionEditSetting},
	})
	apiV1.POST("/resource/{namespace}/{app}/{kind}/setting", updateResourceSetting)
	apiV1.POST("/resource/{namespace}/{app}/{kind}/{name}/setting", updateResourceSetting)

	getResourceSetting := core.Authenticated(core.GetResourceSetting, v1.RBACConstraints{
		Roles:       []string{RoleUser},
		Permissions: []string{PermissionViewSetting},
	})
	apiV1.GET("/resource/{namespace}/{app}/{kind}/setting", getResourceSetting)
	apiV1.GET("/resource/{namespace}/{app}/{kind}/{name}/setting", getResourceSetting)

	updateResourceSchemaHandler := core.Authenticated(core.UpdateResourceSchemaHandler, v1.RBACConstraints{
		Roles:       []string{RoleUser},
		Permissions: []string{PermissionEditSchema},
	})
	apiV1.POST("/resource/{namespace}/{app}/{kind}/schema", updateResourceSchemaHandler)
	apiV1.POST("/resource/{namespace}/{app}/{kind}/{name}/schema", updateResourceSchemaHandler)

	getResourceSchemaHandler := core.Authenticated(core.GetResourceSchemaHandler, v1.RBACConstraints{
		Roles:       []string{RoleUser},
		Permissions: []string{PermissionViewSchema},
	})
	apiV1.GET("/resource/{namespace}/{app}/{kind}/schema", getResourceSchemaHandler)
	apiV1.GET("/resource/{namespace}/{app}/{kind}/{name}/schema", getResourceSchemaHandler)

	getAuditLogHanlder := core.Authenticated(core.GetAuditEvents, v1.RBACConstraints{
		Roles: []string{RoleAdmin},
	})
	apiV1.GET("/audit", getAuditLogHanlder)

	// it is best to take SAILOR_PORT through ENV because people have their own
	// deployment flow
	port := os.Getenv("SAILOR_PORT")
	if port == "" {
		port = ":7766"
	}
	core.Log.Info("[🐧] starting core sailor server", zap.String("port", port))

	if err := fasthttp.ListenAndServe(port, r.Handler); err != nil {
		core.Log.Error("unable to start sailor core", zap.Error(err))
	}
}
