package handlers

import (
	"encoding/json"
	"net/http"

	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) ListUserHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)

	// TODO :: here we need to check it the current user role is admin and then give out the list
	db := sc.dbconns[BUCKET_ADMIN]

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
