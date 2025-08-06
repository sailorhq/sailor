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
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) CreateUserHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	db := sc.dbconns[BUCKET_ADMIN]

	var user User
	if err := json.Unmarshal(ctx.Request.Body(), &user); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "input format not accepted"})
		return
	}

	if user.Email == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "email is required"})
		return
	}

	var pwd string
	err := db.Update(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket([]byte(BUCKET_USERS))
		if userBucket == nil {
			return fmt.Errorf("user bucket not found")
		}

		userBytes := userBucket.Get([]byte(user.Email))
		if userBytes != nil {
			return fmt.Errorf("user already exists")
		}

		pwd = generateUniqueKey("su", 8)
		user := DBUser{
			Email:       user.Email,
			Username:    user.Email,
			Password:    sha256.Sum256([]byte(pwd)), // su- prefix stands for sailor user
			Roles:       user.Roles,
			Permissions: user.Permissions,
			AllowedApps: user.AllowedApps,
		}

		userBytes, err := json.Marshal(user)
		if err != nil {
			return err
		}

		return userBucket.Put([]byte(user.Email), userBytes)
	})

	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	enc.Encode(map[string]any{
		"note": "sailor user password are only generated once, make sure pass is not misplaced.",
		"pass": pwd,
	})
}
