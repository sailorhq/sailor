package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "Invalid request!"})
		return
	}

	username := r.Header.Get("x-username")
	password := r.Header.Get("x-password")

	db := sc.dbconns[BUCKET_ADMIN]

	err := db.Update(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket([]byte(BUCKET_USERS))
		userBytes := userBucket.Get([]byte(username))
		if userBytes == nil {
			return fmt.Errorf("User not found")
		}

		var user DBUser
		err := json.Unmarshal(userBytes, &user)
		if err != nil {
			return err
		}

		if !slices.Contains(user.Roles, "admin") {
			return fmt.Errorf("User is not an admin")
		}

		user.Password = sha256.Sum256([]byte(password))
		adminBytes, err := json.Marshal(user)
		if err != nil {
			return err
		}

		return userBucket.Put([]byte(username), adminBytes)
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}
}
