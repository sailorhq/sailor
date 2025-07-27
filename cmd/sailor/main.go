package main

import (
	"os"

	"github.com/codekidx/sailor/cmd/sailor/console"
	"github.com/codekidx/sailor/cmd/sailor/handlers"
	"go.uber.org/zap"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

// var staticConsoleFS embed.FS

func main() {

	core := handlers.NewSailorCore()

	if core == nil {
		panic("core of sailor was unable to start! check for errors...")
	}

	core.Log.Info("sailor core initialized")

	r := router.New()
	apiV1 := r.Group("/api/v1")
	consoleV1 := r.Group("/console/v1")

	console.Initialize(consoleV1)

	core.Log.Info("initializing core routes")

	// CORE
	//
	// this block contains APIs which is responsible of core working of sailor
	// like obtaining a sailor token or providing the access key and secret key
	// to an authorized party!
	apiV1.POST("/setting", core.SailorSettingHandler)
	apiV1.GET("/auth", core.AuthHandler)
	apiV1.POST("/auth/callback", core.AuthCallbackHandler)

	// PROJECT
	//
	// this block contains APIs that work on project. A project is a collection of
	// resources, and a project should contain a namespace and an app value.
	apiV1.PUT("/project/{namespace}/{app}", core.CreateProjectHandler)

	// RESOURCE
	//
	// this block contains APIs that work and operate on resources inside a project.
	// These APIs handles three types of resources:
	//
	// config - a key value pair with value as {any} type
	// secret - a key value pair with value as string and is encrypted
	// misc   - a text value which cannot be interpretted at the moment, but the
	//          user knows how to make sense of it.
	apiV1.PUT("/resource/{namespace}/{app}/{kind}", core.CreateResourceHandler)
	apiV1.PUT("/resource/{namespace}/{app}/{kind}/{name}", core.CreateResourceHandler)

	apiV1.GET("/resource/{namespace}/{app}/{kind}", core.GetResourceHandler)
	apiV1.GET("/resource/{namespace}/{app}/{kind}/{name}", core.GetResourceHandler)

	apiV1.PUT("/resource/{namespace}/{app}/{kind}/deployment", core.CreateDeploymentHandler)
	apiV1.PUT("/resource/{namespace}/{app}/{kind}/{name}/deployment", core.CreateDeploymentHandler)

	apiV1.POST("/resource/{namespace}/{app}/{kind}/deploy", core.DeployResourceHandler)
	apiV1.POST("/resource/{namespace}/{app}/{kind}/{name}/deploy", core.DeployResourceHandler)

	apiV1.POST("/resource/{namespace}/{app}/{kind}/setting", core.UpdateResourceSetting)
	apiV1.POST("/resource/{namespace}/{app}/{kind}/{name}/setting", core.UpdateResourceSetting)

	apiV1.GET("/resource/{namespace}/{app}/{kind}/setting", core.GetResourceSetting)
	apiV1.GET("/resource/{namespace}/{app}/{kind}/{name}/setting", core.GetResourceSetting)

	apiV1.POST("/resource/{namespace}/{app}/{kind}/schema", core.UpdateResourceSchemaHandler)
	apiV1.POST("/resource/{namespace}/{app}/{kind}/{name}/schema", core.UpdateResourceSchemaHandler)

	apiV1.GET("/resource/{namespace}/{app}/{kind}/schema", core.GetResourceSchemaHandler)
	apiV1.GET("/resource/{namespace}/{app}/{kind}/{name}/schema", core.GetResourceSchemaHandler)

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
