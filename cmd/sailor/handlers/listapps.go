package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (sc *SailorCore) ListAppsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	projectDetails := make(map[string][]string)

	// TODO :: we should ideally not use dbcons for this
	for projectKey := range sc.dbconns {
		if strings.HasPrefix(projectKey, "_") {
			continue
		}

		projectElements := strings.SplitN(projectKey, "-", 2)
		ns := projectElements[0]
		app := projectElements[1]
		if _, ok := projectDetails[ns]; !ok {
			projectDetails[ns] = []string{app}
			continue
		}
		projectDetails[ns] = append(projectDetails[ns], app)
	}

	enc := json.NewEncoder(w)
	enc.Encode(projectDetails)
}
