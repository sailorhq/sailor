package handlers

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) GetUserHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	var (
		userKey string
		ok      bool
	)
	if userKey, ok = ctx.UserValue("user").(string); !ok {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "no user provided"})
		return
	}

	if strings.EqualFold(userKey, "admin@super.sailor") {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "cannot get super sailor details"})
		return
	}

	var user DBUser
	err := sc.dbconns[BUCKET_ADMIN].View(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket([]byte(BUCKET_USERS))
		userBytes := userBucket.Get([]byte(userKey))
		if userBytes == nil {
			return errors.New("no such sailor user")
		}

		return json.Unmarshal(userBytes, &user)
	})

	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	enc.Encode(user)
}
