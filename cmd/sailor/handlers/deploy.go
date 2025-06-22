package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	bolt "go.etcd.io/bbolt"
)

func (sh *SailorCore) DeployHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	enc := json.NewEncoder(w)

	ns := r.URL.Query().Get("ns")
	app := r.URL.Query().Get("app")
	deploy_ver := r.URL.Query().Get("deploy_ver")
	_ = r.URL.Query().Get("key")

	if ns == "" || app == "" {
		enc.Encode(ResponseMessage{Message: "namespace or app is empty"})
		return
	}

	if deploy_ver == "" {
		enc.Encode(ResponseMessage{Message: "deployment version cannot be empty"})
		return
	}

	dbpath := fmt.Sprintf("./configs/%s-%s.db", ns, app)
	if f, _ := os.Stat(dbpath); f == nil {
		enc.Encode(ResponseMessage{Message: "app not present in this namespace"})
		return
	}

	db, err := bolt.Open(dbpath, 0600, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		metaBucket := tx.Bucket([]byte(BUCKET_META))
		return metaBucket.Put([]byte("deploy_ver"), []byte(deploy_ver))
	})

	if err != nil {
		enc.Encode(err.Error())
		return
	}

	enc.Encode("done!")

}
