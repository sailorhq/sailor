package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	bolt "go.etcd.io/bbolt"

	diffmod "github.com/sergi/go-diff/diffmatchpatch"
)

type ProjectStatus string

const (
	Alive    ProjectStatus = "alive"
	Archived ProjectStatus = "archived"
)

type Project struct {
	Ns     string        `json:"ns"`
	App    string        `json:"app"`
	Status ProjectStatus `json:"status"`
}

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

	projectKey := fmt.Sprintf("%s-%s", params.Ns, params.App)
	if _, ok := sc.dbconns[projectKey]; ok {
		enc.Encode(ResponseMessage{Message: "app already exists"})
		return
	}

	db, err := bolt.Open(fmt.Sprintf("./configs/%s-%s.%s", params.Ns, params.App, DB_EXT), 0600, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	sc.dbconns[projectKey] = db

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
		if err != nil {
			return err
		}

		if err = metaBucket.Put([]byte(KEY_ACCESS_KEY), []byte(generateAccessKey(16))); err != nil {
			return err
		}

		return metaBucket.Put([]byte(KEY_SECRET_KEY), []byte(generateAccessKey(32)))
	})

	adminDB := sc.dbconns[BUCKET_ADMIN]
	adminDB.Update(func(tx *bolt.Tx) error {
		projectBucket := tx.Bucket([]byte(BUCKET_PROJECTS))
		projectBytes, err := json.Marshal(Project{Ns: params.Ns, App: params.App, Status: Alive})
		if err != nil {
			return err
		}
		return projectBucket.Put([]byte(params.Ns+"-"+params.App), projectBytes)
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

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateAccessKey(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return fmt.Sprintf("sailor-%s", string(b))
}
