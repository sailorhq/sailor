package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	bolt "go.etcd.io/bbolt"
)

func (sh *SailorCore) UpdateAppMetaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	enc := json.NewEncoder(w)

	ns := r.URL.Query().Get("ns")
	app := r.URL.Query().Get("app")
	accessKey := r.URL.Query().Get("key")

	if ns == "" || app == "" {
		enc.Encode(ResponseMessage{Message: "namespace or app is empty"})
		return
	}

	if accessKey == "" {
		enc.Encode(ResponseMessage{Message: "nothing to update"})
		return
	}

	dbpath := fmt.Sprintf("./configs/%s-%s.db", ns, app)
	if f, _ := os.Stat(dbpath); f == nil {
		enc.Encode(ResponseMessage{Message: "app not present inside this namespace, create one."})
		return
	}

	db, err := bolt.Open(dbpath, 0600, nil)
	if err != nil {
		enc.Encode(ResponseMessage{Message: "unable to access your config, contact your admin."})
		return
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		metaBucket := tx.Bucket([]byte(BUCKET_META))
		return metaBucket.Put([]byte(KEY_ACCESS_KEY), []byte(accessKey))
	})

	if err != nil {
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	enc.Encode(ResponseMessage{
		Message: fmt.Sprintf("created namespace: %s | app: %s | access_key: %v",
			ns, app, accessKey != ""),
	})
}
