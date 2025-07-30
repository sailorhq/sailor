package handlers

import (
	"encoding/json"
	"errors"

	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
)

// GetTokenHandler gives sailor specific token to the user, to interact with
// the resource APIs, the /token call happens after oidc callback
func (sc *SailorCore) GetTokenHandler(ctx *fasthttp.RequestCtx) {
	// OPT :: the token call should happen after login within 5minutes by default
	// but depending upon your requirements & security measures/flow you should
	// ideally change this using sailor settings, which should be added later
	enc := json.NewEncoder(ctx)
	username := string(ctx.Request.Header.Peek("x-username"))
	if username == "" {
		// TODO :: should there be source IP here?
		sc.Log.Error("token request was called without username")
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "invalid request"})
		return
	}

	var token string
	var err error
	err = sc.dbconns[BUCKET_ADMIN].View(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket([]byte(BUCKET_USERS))
		if userBucket == nil {
			// this case should not happen because while creating _admin sail
			// we create user bucket, so it is technically a unrecoverable error
			return errors.New("cannot find users volume to save into")
		}

		userBytes := userBucket.Get([]byte(username))
		if userBytes == nil {
			return errors.New("sailor user not created")
		}

		var user DBUser
		if err = json.Unmarshal(userBytes, &user); err != nil {
			return errors.New("sailor user is not parsable")
		}

		token = user.Token
		return nil
	})

	if err != nil {
		sc.Log.Error("encountered an error sailor user token creation", zap.Error(err))
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	enc.Encode(map[string]any{"token": token})
}
