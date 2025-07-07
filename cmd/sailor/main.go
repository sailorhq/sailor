package main

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/codekidx/sailor/cmd/sailor/handlers"
)

//go:embed console
var staticConsoleFS embed.FS

func main() {

	sh := handlers.NewSailorCore()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/create", sh.CreateAppHandler)
	mux.HandleFunc("/api/v1/update", sh.UpdateAppMetaHandler)
	mux.HandleFunc("/api/v1/config", sh.ConfigHandler)
	mux.HandleFunc("/api/v1/deploy", sh.DeployHandler)
	mux.HandleFunc("/api/v1/rules", sh.RuleHandler)
	mux.HandleFunc("/api/v1/version", sh.VersionHandler)
	mux.HandleFunc("/api/v1/state", sh.SailorStateHandler)

	mux.HandleFunc("/api/v1/auth", sh.AuthHandler)
	mux.HandleFunc("/api/v1/auth.validate", sh.ValidateHandler)

	mux.HandleFunc("/api/v1/admin.create.user", sh.CreateUserHandler)
	mux.HandleFunc("/api/v1/admin.edit.user", sh.EditUserHandler)
	mux.HandleFunc("/api/v1/admin.list.users", sh.ListUserHandler)
	mux.HandleFunc("/api/v1/admin.list.apps", sh.ListAppsHandler)
	mux.HandleFunc("/api/v1/admin.app.state", sh.AdminStateHandler)
	mux.HandleFunc("/api/v1/admin.secrets", sh.AddSecretHandler)
	mux.HandleFunc("/api/v1/admin.change.password", sh.ChangePasswordHandler)
	mux.HandleFunc("/api/v1/admin.backup", sh.BackupHandler)

	mux.HandleFunc("/api/v1/audit.trail", sh.AuditHandler)

	// Serve static files from the embedded console directory
	consoleFS, err := fs.Sub(staticConsoleFS, "console")
	if err != nil {
		panic(err)
	}

	// Create a file server that serves the embedded files
	fileServer := http.FileServer(http.FS(consoleFS))

	// Handle console routes - serve static files and fallback to index.html for SPA routing
	mux.HandleFunc("/console/", func(w http.ResponseWriter, r *http.Request) {
		// Remove the /console prefix from the request path
		requestPath := strings.TrimPrefix(r.URL.Path, "/console")
		if requestPath == "" {
			requestPath = "/"
		}

		// Check if the file exists in the embedded filesystem
		if _, err := fs.Stat(consoleFS, requestPath[1:]); err == nil {
			// File exists, serve it
			r.URL.Path = requestPath
			fileServer.ServeHTTP(w, r)
			return
		}

		// File doesn't exist, serve index.html for SPA routing
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/console", http.StatusFound)
	})

	http.ListenAndServe(":7766", mux)
}
