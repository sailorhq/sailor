package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/codekidx/sailor/internal/types"
	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) StateHandler(w http.ResponseWriter, r *http.Request) {
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

	configStr := buildConfig(db)
	var builtConfig map[string]any
	json.Unmarshal([]byte(configStr), &builtConfig)
	w.Header().Set("Content-Type", "application/json")

	resp := types.SailorState{
		Configs: builtConfig,
	}

	var secrets = make(map[string]string)
	db.View(func(tx *bolt.Tx) error {
		// fetch current deployed version ...
		metaBucket := tx.Bucket([]byte(BUCKET_META))
		resp.Meta.Version = string(metaBucket.Get([]byte(KEY_DEPLOYED_VERSION)))

		// fetch secrets...
		secretsBucket := tx.Bucket([]byte(BUCKET_SECRETS))
		cur := secretsBucket.Cursor()

		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			secrets[string(k)] = string(v)
		}

		return nil
	})

	resp.Secrets = secrets

	enc.Encode(&resp)
}
