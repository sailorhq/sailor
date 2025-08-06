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
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	bolt "go.etcd.io/bbolt"
)

func (c *Console) ValidateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tokenHeader := r.Header.Get("x-token")
	if tokenHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenHeader, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}
		return []byte("TODO :: access key"), nil
	})
	if err != nil || token == nil || !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	db := c.dbconns[BUCKET_ADMIN]

	enc := json.NewEncoder(w)

	var user DBUser
	err = db.View(func(tx *bolt.Tx) error {
		usersBucket := tx.Bucket([]byte(BUCKET_USERS))
		userBytes := usersBucket.Get([]byte(r.Header.Get("x-username")))
		if userBytes == nil {
			return errors.New("user not found")
		}

		err = json.Unmarshal(userBytes, &user)
		if err != nil {
			return err
		}

		// TODO :: check the status while type casting
		if user.Password != claims["data"].(string) {
			return errors.New("invalid password")
		}

		return nil
	})

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	enc.Encode(AuthResponse{
		Username:    user.Username,
		Permissions: user.Permissions,
		Roles:       user.Roles,
		Token:       tokenHeader,
	})
}
