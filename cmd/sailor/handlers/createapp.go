package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	bolt "go.etcd.io/bbolt"

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

	ns := ctx.UserValue("namespace").(string)
	app := ctx.UserValue("app").(string)

	if ns == "" || app == "" {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "namespace or app should not be empty"})
		return
	}

	projectKey := fmt.Sprintf("%s-%s", ns, app)
	db, err := bolt.Open(fmt.Sprintf("./configs/%s.%s", projectKey, DB_EXT), 0600, nil)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	sc.dbconns[projectKey] = db

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

		if err = metaBucket.Put([]byte(KEY_ACCESS_KEY), []byte(generateAccessKey("sailor", 16))); err != nil {
			return err
		}

		return metaBucket.Put([]byte(KEY_SECRET_KEY), []byte(generateAccessKey("secret", 32)))
	})

	adminDB := sc.dbconns[BUCKET_ADMIN]
	adminDB.Update(func(tx *bolt.Tx) error {
		projectBucket := tx.Bucket([]byte(BUCKET_PROJECTS))
		projectBytes, err := json.Marshal(Project{Ns: ns, App: app})
		if err != nil {
			return err
		}
		return projectBucket.Put([]byte(projectKey), projectBytes)
	})

	if err != nil {
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	enc.Encode(ResponseMessage{Message: fmt.Sprintf("created namespace: %s | app: %s", ns, app)})
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateAccessKey(prefix string, length int) string {

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return fmt.Sprintf("%s-%s", prefix, string(b))
}
