package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) AuthBasicHandler(ctx *fasthttp.RequestCtx) {
	user := string(ctx.Request.Header.Peek("x-user"))
	pass := ctx.Request.Header.Peek("x-pass")

	var token string
	err := sc.dbconns[BUCKET_ADMIN].View(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket([]byte(BUCKET_USERS))
		userBytes := userBucket.Get([]byte(user))
		if userBytes == nil {
			return errors.New("no such user")
		}

		var user DBUser
		json.Unmarshal(userBytes, &user)

		if sha256.Sum256(pass) != user.Password {
			return errors.New("invalid credentials")
		}

		var err error
		token, err = jwtFromClaims(jwt.MapClaims{
			"iss":          "sailor",
			"sub":          "sailor-basic",
			"aud":          "sailor",
			"exp":          time.Now().Add(24 * time.Hour).Unix(),
			"iat":          time.Now().Unix(),
			"email":        user.Email,
			"roles":        user.Roles,
			"permissions":  user.Permissions,
			"allowed_apps": user.AllowedApps,
		}, sc.setting.TokenKey)
		if err != nil {
			return errors.New("error while creating token")
		}

		return nil
	})

	enc := json.NewEncoder(ctx)

	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusUnauthorized)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	ctx.Response.Header.Set("Content-Type", "application/json")
	enc.Encode(map[string]any{
		"token": token,
	})
}
