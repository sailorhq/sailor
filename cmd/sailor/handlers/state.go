package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/codekidx/sailor/internal/types"
	bolt "go.etcd.io/bbolt"
)

func (sh *SailorCore) StateHandler(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("ns")
	app := r.URL.Query().Get("app")
	_ = r.URL.Query().Get("key")

	enc := json.NewEncoder(w)

	if ns == "" || app == "" {
		enc.Encode(ResponseMessage{Message: "namespace or app is empty"})
		return
	}

	dbpath := fmt.Sprintf("./configs/%s-%s.db", ns, app)
	if f, _ := os.Stat(dbpath); f == nil {
		enc.Encode(ResponseMessage{Message: "no such app in this namespace"})
		return
	}

	db, err := bolt.Open(dbpath, 0600, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

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

	enc.Encode(resp)
}
