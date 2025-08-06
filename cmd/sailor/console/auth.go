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
package console

import (
	"encoding/json"
	"errors"
	"net/http"

	bolt "go.etcd.io/bbolt"
)

type AuthResponse struct {
	Username    string   `json:"username"`
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
	Token       string   `json:"token"`
}

func (c *Console) AuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	enc := json.NewEncoder(w)

	username := r.Header.Get("x-username")
	password := r.Header.Get("x-password")

	db := c.dbconns[BUCKET_ADMIN]

	var user DBUser
	err := db.View(func(tx *bolt.Tx) error {
		usersBucket := tx.Bucket([]byte(BUCKET_USERS))
		userBytes := usersBucket.Get([]byte(username))

		if userBytes == nil {
			return errors.New("user not found, let your admin create an account for you")
		}

		err := json.Unmarshal(userBytes, &user)
		if err != nil {
			return err
		}

		if user.Password != password {
			return errors.New("invalid password")
		}

		return nil
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	// expiry := time.Now().Add(time.Hour * 24)
	token, err := jwtfyit(user.Password, "TODO :: access key", nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	enc.Encode(AuthResponse{
		Username:    user.Username,
		Permissions: user.Permissions,
		Roles:       user.Roles,
		Token:       token,
	})
}
