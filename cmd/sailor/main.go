package main

import (
	"net/http"

	"github.com/codekidx/sailor/cmd/sailor/handlers"
)

func main() {

	sh := handlers.NewSailorCore()

	http.HandleFunc("/create", sh.CreateAppHandler)
	http.HandleFunc("/update", sh.UpdateAppMetaHandler)
	http.HandleFunc("/config", sh.ConfigHandler)
	http.HandleFunc("/deploy", sh.DeployHandler)
	http.HandleFunc("/rules", sh.RuleHandler)
	http.HandleFunc("/state", sh.StateHandler)

	http.ListenAndServe(":7766", nil)
}
