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
	"fmt"
	"net/http"
	"strings"

	bolt "go.etcd.io/bbolt"
)

func (c *Console) EditUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	enc := json.NewEncoder(w)

	db := c.dbconns[BUCKET_ADMIN]

	err := db.Update(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket([]byte(BUCKET_USERS))
		if userBucket == nil {
			return fmt.Errorf("user bucket not found")
		}

		userBytes := userBucket.Get([]byte(r.URL.Query().Get("username")))
		if userBytes == nil {
			return fmt.Errorf("user not found")
		}

		var user DBUser
		json.Unmarshal(userBytes, &user)

		permissions := r.URL.Query().Get("permissions")
		allowedApps := r.URL.Query().Get("allowed_apps")
		user.Roles = []string{r.URL.Query().Get("role")}
		if strings.TrimSpace(permissions) != "" {
			user.Permissions = strings.Split(permissions, "|")
		} else {
			user.Permissions = []string{}
		}
		if strings.TrimSpace(allowedApps) != "" {
			user.AllowedApps = strings.Split(allowedApps, "|")
		} else {
			user.AllowedApps = []string{}
		}

		userBytes, err := json.Marshal(user)
		if err != nil {
			return err
		}

		return userBucket.Put([]byte(r.URL.Query().Get("username")), userBytes)
	})

	if err != nil {
		enc.Encode(ResponseMessage{Message: err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte("User edited successfully"))
}
