package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	bolt "go.etcd.io/bbolt"
)

func (sh *SailorCore) RuleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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

	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	if ok := json.Valid(b); !ok {
		enc.Encode(ResponseMessage{Message: "invalid json for rules!"})
		return
	}

	if err = db.Update(func(tx *bolt.Tx) error {
		metaBucket := tx.Bucket([]byte(BUCKET_META))
		return metaBucket.Put([]byte(KEY_RULES), b)
	}); err != nil {
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

}
