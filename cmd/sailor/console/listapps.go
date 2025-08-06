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
	"slices"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sailorhq/sailor/internal/types"
	bolt "go.etcd.io/bbolt"
)

func (c *Console) ListAppsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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

	db := c.dbconns[BUCKET_ADMIN]

	var user DBUser
	var adminBackupState types.AdminBackupState
	err = db.View(func(tx *bolt.Tx) error {
		usersBucket := tx.Bucket([]byte(BUCKET_USERS))
		username := r.Header.Get("x-username")
		userBytes := usersBucket.Get([]byte(username))

		if userBytes == nil {
			return errors.New("user not found, let your admin create an account for you")
		}

		err = json.Unmarshal(userBytes, &user)

		// if user is admin, add all projects to allowed apps
		if slices.Contains(user.Roles, "admin") {
			projectsBucket := tx.Bucket([]byte(BUCKET_PROJECTS))
			projectsBucket.ForEach(func(k []byte, v []byte) error {
				var project Project
				if err := json.Unmarshal(v, &project); err != nil {
					return err
				}

				user.AllowedApps = append(user.AllowedApps, fmt.Sprintf("%s-%s", project.Ns, project.App))

				return nil
			})

			backupBucket := tx.Bucket([]byte(BUCKET_BACKUP))
			backupBytes := backupBucket.Get([]byte(KEY_S3))
			if backupBytes != nil {
				err := json.Unmarshal(backupBytes, &adminBackupState)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	enc := json.NewEncoder(w)
	enc.Encode(types.ListAppsResponse{
		Apps:             user.AllowedApps,
		AdminBackupState: &adminBackupState,
	})
}
