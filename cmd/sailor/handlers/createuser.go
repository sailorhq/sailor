package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	bolt "go.etcd.io/bbolt"
)

func (sh *SailorCore) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	enc := json.NewEncoder(w)

	db := sh.dbconns[BUCKET_ADMIN]

	err := db.Update(func(tx *bolt.Tx) error {
		userBucket := tx.Bucket([]byte(BUCKET_USERS))
		if userBucket == nil {
			return fmt.Errorf("user bucket not found")
		}

		userBytes := userBucket.Get([]byte(r.URL.Query().Get("username")))
		if userBytes != nil {
			return fmt.Errorf("user already exists")
		}

		user := User{
			Username:    r.URL.Query().Get("username"),
			Password:    r.URL.Query().Get("password"),
			Roles:       []string{r.URL.Query().Get("role")},
			Permissions: strings.Split(r.URL.Query().Get("permissions"), "|"),
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

	w.Write([]byte("User created successfully"))
}
