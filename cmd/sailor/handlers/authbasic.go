package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"strings"
	"time"

	v1 "github.com/codekidx/sailor/pkg/core/v1"
	"github.com/golang-jwt/jwt/v5"
	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) AuthBasicHandler(ctx *fasthttp.RequestCtx) {
	user := string(ctx.Request.Header.Peek("x-user"))
	pass := ctx.Request.Header.Peek("x-pass")

	var token string
	var allowedApps []string
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

		allowedApps = user.AllowedApps

		return nil
	})

	enc := json.NewEncoder(ctx)

	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusUnauthorized)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	var keyPairs = make(map[string]v1.KeyPair)
	for _, app := range allowedApps {
		// this is wildcard case
		if strings.HasSuffix(app, "-*") {
			for projectKey, conn := range sc.dbconns {
				ns := strings.Split(app, "-")[0]
				if strings.HasPrefix(projectKey, ns) {
					conn.View(func(tx *bolt.Tx) error {
						metaBucket := tx.Bucket([]byte(BUCKET_META))
						keyPairs[projectKey] = v1.KeyPair{
							AccessKey: string(metaBucket.Get([]byte(KEY_ACCESS_KEY))),
							SecretKey: string(metaBucket.Get([]byte(KEY_SECRET_KEY))),
						}
						return nil
					})
				}
			}
		} else {
			if conn, ok := sc.dbconns[app]; ok {
				conn.View(func(tx *bolt.Tx) error {
					metaBucket := tx.Bucket([]byte(BUCKET_META))
					keyPairs[app] = v1.KeyPair{
						AccessKey: string(metaBucket.Get([]byte(KEY_ACCESS_KEY))),
						SecretKey: string(metaBucket.Get([]byte(KEY_SECRET_KEY))),
					}

					return nil
				})
			}
		}
	}

	ctx.Response.Header.Set("Content-Type", "application/json")
	enc.Encode(v1.LoginResponse{
		Token:    token,
		KeyPairs: keyPairs,
	})
}
