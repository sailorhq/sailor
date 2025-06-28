package main

import (
	"net/http"

	"github.com/codekidx/sailor/cmd/sailor/handlers"
	"github.com/rs/cors"
)

func main() {

	sh := handlers.NewSailorCore()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/create", sh.CreateAppHandler)
	mux.HandleFunc("/api/v1/update", sh.UpdateAppMetaHandler)
	mux.HandleFunc("/api/v1/config", sh.ConfigHandler)
	mux.HandleFunc("/api/v1/deploy", sh.DeployHandler)
	mux.HandleFunc("/api/v1/rules", sh.RuleHandler)
	mux.HandleFunc("/api/v1/state", sh.StateHandler)
	mux.HandleFunc("/api/v1/version", sh.VersionHandler)

	mux.HandleFunc("/api/v1/auth", sh.AuthHandler)
	mux.HandleFunc("/api/v1/auth.validate", sh.ValidateHandler)

	mux.HandleFunc("/api/v1/admin.create.user", sh.CreateUserHandler)

	handler := cors.AllowAll().Handler(mux)
	http.ListenAndServe(":7766", handler)
}
