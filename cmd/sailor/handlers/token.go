// sailor
// Copyright (C) 2025 SailorHQ and Ashish Shekar (codekidX)

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
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
	username := string(ctx.Request.Header.Peek("x-user"))
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
