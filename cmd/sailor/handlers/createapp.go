package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	bolt "go.etcd.io/bbolt"

	diffmod "github.com/sergi/go-diff/diffmatchpatch"
)

type FirstConfig struct {
	App string `json:"app"`
}

func (sc *SailorCore) CreateAppHandler(w http.ResponseWriter, r *http.Request) {
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
		var metaBucket *bolt.Bucket
		metaBucket, err := tx.CreateBucket([]byte(BUCKET_META))
		if err == nil {
			metaBucket.Put([]byte(KEY_RULES), []byte(`{
    "app": "required"
}`))
		}

		diffBucket, err := tx.CreateBucket([]byte(BUCKET_DIFFS))
		if err == nil {
			differ := diffmod.New()
			firstConfBytes, err := json.Marshal(FirstConfig{App: params.App})
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

		if params.AccessKey != "" {
			err = metaBucket.Put([]byte(KEY_ACCESS_KEY), []byte(params.AccessKey))
		}
		return err
	})

	if err != nil {
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	enc.Encode(ResponseMessage{
		Message: fmt.Sprintf("created namespace: %s | app: %s | access_key: %v",
			params.Ns, params.App, params.AccessKey != ""),
	})
}
