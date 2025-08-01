package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	bolt "go.etcd.io/bbolt"

	v1 "github.com/codekidx/sailor/pkg/core/v1"
	"github.com/valyala/fasthttp"
)

type Project struct {
	Ns  string `json:"ns"`
	App string `json:"app"`
}

type FirstConfig struct {
	App string `json:"app"`
}

func (sc *SailorCore) CreateProjectHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	params := extractSailorParams(ctx)

	if params.Ns == "" || params.App == "" {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "namespace or app should not be empty"})
		return
	}

	if _, ok := sc.dbconns[params.ProjectKey]; ok {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "project already exists"})
		return
	}

	db, err := bolt.Open(fmt.Sprintf("./configs/%s.%s", params.ProjectKey, DB_EXT), 0600, nil)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	sc.dbconns[params.ProjectKey] = db

	err = db.Update(func(tx *bolt.Tx) error {
		var metaBucket *bolt.Bucket
		metaBucket, err := tx.CreateBucket([]byte(BUCKET_META))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucket([]byte(BUCKET_DEPLOYMENT))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucket([]byte(BUCKET_RESOURCE))
		if err != nil {
			return err
		}

		if err = metaBucket.Put([]byte(KEY_ACCESS_KEY), []byte(generateUniqueKey("sailor", 16))); err != nil {
			return err
		}

		return metaBucket.Put([]byte(KEY_SECRET_KEY), []byte(generateUniqueKey("secret", 32)))
	})

	adminDB := sc.dbconns[BUCKET_ADMIN]
	adminDB.Update(func(tx *bolt.Tx) error {
		projectBucket := tx.Bucket([]byte(BUCKET_PROJECTS))
		projectBytes, err := json.Marshal(Project{Ns: params.Ns, App: params.App})
		if err != nil {
			return err
		}
		return projectBucket.Put([]byte(params.ProjectKey), projectBytes)
	})

	if err != nil {
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	enc.Encode(v1.ProjectResponse{
		Key: params.ProjectKey,
	})
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateUniqueKey(prefix string, length int) string {

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return fmt.Sprintf("%s-%s", prefix, string(b))
}
