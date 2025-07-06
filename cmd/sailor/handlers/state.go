package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/codekidx/sailor/internal/types"
	bolt "go.etcd.io/bbolt"
)

func (sh *SailorCore) SailorStateHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)

	params, err := sh.extractSailorParams(r)
	if err != nil {
		enc.Encode(ResponseMessage{Message: err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	db, err := sh.getDBConn(params)
	if err != nil {
		enc.Encode(ResponseMessage{Message: err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var state = types.SailorState{
		Secrets: make(map[string][]byte),
	}
	err = db.View(func(tx *bolt.Tx) error {
		metaBucket := tx.Bucket([]byte("_meta"))
		secretsBucket := tx.Bucket([]byte("secrets"))

		secretsBucket.ForEach(func(k, v []byte) error {
			state.Secrets[string(k)] = v
			return nil
		})

		state.Version = string(metaBucket.Get([]byte("deploy_ver")))

		return nil
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	builtConfig := buildConfig(db)
	state.Config = []byte(builtConfig)

	enc.Encode(&state)
}
