package handlers

import (
	"encoding/json"
	"net/http"

	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) DeployHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	enc := json.NewEncoder(w)

	params, err := sc.extractSailorParams(r)
	if err != nil {
		// TODO: log here!
		enc.Encode(ResponseMessage{Message: err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	db, err := sc.getDBConn(params)
	if err != nil {
		enc.Encode(ResponseMessage{Message: err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = db.Update(func(tx *bolt.Tx) error {
		metaBucket := tx.Bucket([]byte(BUCKET_META))
		return metaBucket.Put([]byte("deploy_ver"), []byte(params.DeployVersion))
	})

	if err != nil {
		enc.Encode(err.Error())
		return
	}

	enc.Encode("done!")

}
