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
	"strings"

	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) GetUserHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)
	enc.SetEscapeHTML(false)

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
