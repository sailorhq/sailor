package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	bolt "go.etcd.io/bbolt"

	diffmod "github.com/sergi/go-diff/diffmatchpatch"
)

type FirstConfig struct {
	App string `json:"app"`
}

func CreateAppHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
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

	dbpath := fmt.Sprintf("./configs/%s-%s.db", ns, app)
	if f, _ := os.Stat(dbpath); f != nil {
		enc.Encode(ResponseMessage{Message: "app already present in this namespace"})
		return
	}

	db, err := bolt.Open(dbpath, 0600, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		var metaBucket *bolt.Bucket
		metaBucket, err = tx.CreateBucket([]byte(BUCKET_META))
		if err == nil {
			metaBucket.Put([]byte(KEY_RULES), []byte("{\"app\": \"required\"}"))
		}

		diffBucket, err := tx.CreateBucket([]byte(BUCKET_DIFFS))
		if err == nil {
			differ := diffmod.New()
			firstConfBytes, err := json.Marshal(FirstConfig{App: app})
			if err != nil {
				return err
			}

			diff := differ.DiffMain("", string(firstConfBytes), true)
			patchList := differ.PatchMake("", string(firstConfBytes), diff)
			patchh := differ.PatchToText(patchList)
			fmt.Println("pa: ", patchh)
			if err = diffBucket.Put([]byte("1"), []byte(patchh)); err != nil {
				return err
			} else {
				metaBucket.Put([]byte(KEY_DEPLOYED_VERSION), []byte("1"))
			}
		}

		_, err = tx.CreateBucket([]byte(BUCKET_SECRETS))

		if accessKey != "" {
			err = metaBucket.Put([]byte(KEY_ACCESS_KEY), []byte(accessKey))
		}
		return err
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
