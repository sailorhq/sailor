package main

import (
	"net/http"

	"github.com/codekidx/sailor/cmd/sailor/handlers"
)

func main() {
	http.HandleFunc("/create", handlers.CreateAppHandler)
	http.HandleFunc("/update", handlers.UpdateAppMetaHandler)
	http.HandleFunc("/config", handlers.ConfigHandler)
	http.HandleFunc("/deploy", handlers.DeployHandler)
	http.HandleFunc("/rules", handlers.RuleHandler)

	http.ListenAndServe(":7766", nil)
}
