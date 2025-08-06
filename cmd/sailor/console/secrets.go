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
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sailorhq/sailor/internal/types"
	bolt "go.etcd.io/bbolt"
)

type SecretRequest struct {
	Secrets        []types.Secret `json:"secrets"`
	DeletedSecrets []types.Secret `json:"deleted_secrets"`
}

func (c *Console) AddSecretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	enc := json.NewEncoder(w)

	var secretRequest SecretRequest
	err := json.NewDecoder(r.Body).Decode(&secretRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	if len(secretRequest.Secrets) == 0 && len(secretRequest.DeletedSecrets) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "no secrets provided"})
		return
	}

	params, err := c.extractSailorParams(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	if db, ok := c.dbconns[params.ProjectKey]; ok {
		err := db.Update(func(tx *bolt.Tx) error {
			secretsBucket := tx.Bucket([]byte(BUCKET_SECRETS))
			metaBucket := tx.Bucket([]byte(BUCKET_META))
			secretKey := metaBucket.Get([]byte(KEY_SECRET_KEY))

			for _, secret := range secretRequest.Secrets {
				token, err := jwtfyit(secret.Value, string(secretKey), nil)
				if err != nil {
					return err
				}
				err = secretsBucket.Put([]byte(secret.Name), []byte(token))
				if err != nil {
					return err
				}
			}

			for _, secret := range secretRequest.DeletedSecrets {
				err = secretsBucket.Delete([]byte(secret.Name))
				if err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			enc.Encode(ResponseMessage{Message: err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		enc.Encode(ResponseMessage{Message: "secret added"})
		return
	}

	w.WriteHeader(http.StatusNotFound)
	enc.Encode(ResponseMessage{Message: "app not present in this namespace"})
}

func jwtfyit(data string, accessKey string, expiry *time.Time) (string, error) {
	// Create a new token
	token := jwt.New(jwt.SigningMethodHS256)
	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["data"] = data

	if expiry != nil {
		claims["exp"] = *expiry
	}

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(accessKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
