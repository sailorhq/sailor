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

	bolt "go.etcd.io/bbolt"
)

func (c *Console) ListUserHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)

	// TODO :: here we need to check it the current user role is admin and then give out the list
	db := c.dbconns[BUCKET_ADMIN]

	var users []User
	db.View(func(tx *bolt.Tx) error {
		usersBucket := tx.Bucket([]byte(BUCKET_USERS))

		usersBucket.ForEach(func(k, v []byte) error {
			// skip admin user
			if string(k) == "admin" {
				return nil
			}

			var user User
			err := json.Unmarshal(v, &user)
			if err != nil {
				return err
			}

			users = append(users, user)

			return nil
		})

		return nil
	})

	enc.Encode(users)
}
