package handlers

import (
	"fmt"
	"net/http"

	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) VersionHandler(w http.ResponseWriter, r *http.Request) {
	params, err := sc.extractSailorParams(r)
	if err != nil {
		// TODO: log here!
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	db, err := sc.getDBConn(params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var ver string
	db.View(func(tx *bolt.Tx) error {
		// fetch current deployed version ...
		metaBucket := tx.Bucket([]byte(BUCKET_META))
		ver = string(metaBucket.Get([]byte(KEY_DEPLOYED_VERSION)))

		return nil
	})

	fmt.Fprint(w, ver)
}
