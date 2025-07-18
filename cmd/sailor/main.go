package main

import (
	"github.com/codekidx/sailor/cmd/sailor/handlers"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

// var staticConsoleFS embed.FS

func main() {

	core := handlers.NewSailorCore()

	if core == nil {
		panic("core of sailor was unable to start! check for errors...")
	}

	r := router.New()

	// --- SRD - Sailor Resource Definition ---
	// KIND OF RESOURCE:
	// - CONFIGS
	// - SECRETS
	// - MISC - (can only be fetched when asked with a resource name)

	// r.PUT("/api/v1/resource/{namespace}/{app}", core.CreateResourceHandler) // PUT
	// mux.HandleFunc("/api/v1/resource/all", sh.SailorStateHandler)               // GET

	r.PUT("/api/v1/project/{namespace}/{app}", core.CreateProjectHandler)

	r.PUT("/api/v1/resource/{namespace}/{app}/{kind}", core.CreateResourceHandler) // PUT -- this is for config and secrets which doesn't require resource name

	r.PUT("/api/v1/resource/{namespace}/{app}/{kind}/{name}", core.CreateResourceHandler) // PUT -- this is for misc resource

	r.GET("/api/v1/resource/{namespace}/{app}/{kind}", core.GetResourceHandler)

	r.GET("/api/v1/resource/{namespace}/{app}/{kind}/{name}", core.GetResourceHandler)

	r.PUT("/api/v1/resource/{namespace}/{app}/{kind}/deployment", core.CreateDeploymentHandler) // PUT

	r.PUT("/api/v1/resource/{namespace}/{app}/{kind}/{name}/deployment", core.CreateDeploymentHandler) // PUT

	r.POST("/api/v1/resource/{namespace}/{app}/{kind}/deploy", core.DeployResourceHandler) // POST

	r.POST("/api/v1/resource/{namespace}/{app}/{kind}/{name}/deploy", core.DeployResourceHandler) // POST

	r.POST("/api/v1/resource/{namespace}/{app}/{kind}/setting", core.UpdateResourceSetting)
	r.POST("/api/v1/resource/{namespace}/{app}/{kind}/{name}/setting", core.UpdateResourceSetting)

	r.GET("/api/v1/resource/{namespace}/{app}/{kind}/setting", core.GetResourceSetting)
	r.GET("/api/v1/resource/{namespace}/{app}/{kind}/{name}/setting", core.GetResourceSetting)

	// mux.HandleFunc("/api/v1/resource/{kind}/{name}", sh.SailorStateHandler)         // GET
	// mux.HandleFunc("/api/v1/resource/{kind}/{name}/version", sh.SailorStateHandler) // GET
	// resource setting is used to tell Sailor how the resource should be handled
	// and where all it should be deployed
	// { deploy: { k8s: true } }
	// mux.HandleFunc("/api/v1/resource/{kind}/{name}/setting", sh.SailorStateHandler) // GET
	// deploy will update to sail file but will Rollback incase k8s deployment is mentioned
	// but it was not successful
	// mux.HandleFunc("/api/v1/resource/{kind}/{name}/deploy", sh.SailorStateHandler) // GET

	// RULES AND SCHEMAS - FOR A RESOURCE
	// mux.HandleFunc("/api/v1/resource/{kind}/{name}/schema", sh.SailorStateHandler)        // GET - (defaults to .strict = true)
	// mux.HandleFunc("/api/v1/resource/{kind}/{name}/schema/create", sh.SailorStateHandler) // POST

	// mux.HandleFunc("/api/v1/project/{namespace}/{app}/setting", sh.CreateAppHandler) // GET
	// POST - will always merge over the default settings map
	// mux.HandleFunc("/api/v1/project/{namespace}/{app}/setting/edit", sh.CreateAppHandler)

	// These are admin based actions, this should require a proper authorization mechanism
	// mux.HandleFunc("/api/v1/_sailor/bucket/{name}", sh.AddSecretHandler) // PUT
	// mux.HandleFunc("/api/v1/_sailor/user/{name}", sh.AddSecretHandler) // POST

	// mux.HandleFunc("/api/v1/update", sh.UpdateAppMetaHandler)
	// mux.HandleFunc("/api/v1/config", sh.ConfigHandler)   // replaced by - /api/v1/resource/{kind}/{name}
	// mux.HandleFunc("/api/v1/deploy", sh.DeployHandler)   // replaced by - /api/v1/resource/{kind}/{name}/deploy
	// mux.HandleFunc("/api/v1/rules", sh.RuleHandler)      // replaced by - /api/v1/resource/{kind}/{name}/schema../create
	// mux.HandleFunc("/api/v1/version", sh.VersionHandler) // replaced by - /api/v1/resource/{kind}/{name}/version
	// !!DEPRECATION - since we are trying to create a resource modal, we will not need a state API
	// which will give everything all at once, the client needs to ask for resources it needs.
	// mux.HandleFunc("/api/v1/state", sh.SailorStateHandler)

	// TODO :: check if we can add OIDC or any SSO login for easy integration for organizations
	// mux.HandleFunc("/api/v1/auth", sh.AuthHandler)
	// mux.HandleFunc("/api/v1/auth.validate", sh.ValidateHandler)

	// CONSOLE APIs...
	// TODO :: need to use the same format as our resource model
	// mux.HandleFunc("/api/v1/console.create.user", sh.CreateUserHandler)
	// mux.HandleFunc("/api/v1/console.edit.user", sh.EditUserHandler)
	// mux.HandleFunc("/api/v1/console.list.users", sh.ListUserHandler)
	// mux.HandleFunc("/api/v1/console.list.apps", sh.ListAppsHandler)
	// mux.HandleFunc("/api/v1/console.app.state", sh.AdminStateHandler)
	// mux.HandleFunc("/api/v1/console.secrets", sh.AddSecretHandler)
	// mux.HandleFunc("/api/v1/console.change.password", sh.ChangePasswordHandler)
	// mux.HandleFunc("/api/v1/console.backup", sh.BackupHandler)

	// mux.HandleFunc("/api/v1/audit.trail", sh.AuditHandler)

	// Serve static files from the embedded console directory
	// consoleFS, err := fs.Sub(staticConsoleFS, "console")
	// if err != nil {
	// 	panic(err)
	// }

	// Create a file server that serves the embedded files
	// _ = http.FileServer(http.FS(consoleFS))

	// Handle console routes - serve static files and fallback to index.html for SPA routing
	// mux.HandleFunc("/console/", func(w http.ResponseWriter, r *http.Request) {
	// Remove the /console prefix from the request path
	// 	requestPath := strings.TrimPrefix(r.URL.Path, "/console")
	// 	if requestPath == "" {
	// 		requestPath = "/"
	// 	}

	// 	// Check if the file exists in the embedded filesystem
	// 	if _, err := fs.Stat(consoleFS, requestPath[1:]); err == nil {
	// 		// File exists, serve it
	// 		r.URL.Path = requestPath
	// 		fileServer.ServeHTTP(w, r)
	// 		return
	// 	}

	// 	// File doesn't exist, serve index.html for SPA routing
	// 	r.URL.Path = "/"
	// 	fileServer.ServeHTTP(w, r)
	// })

	// mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	http.Redirect(w, r, "/console", http.StatusFound)
	// })

	fasthttp.ListenAndServe(":7766", r.Handler)
}
