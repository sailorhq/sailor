package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	bolt "go.etcd.io/bbolt"
)

func (sh *SailorCore) EditUserHandler(w http.ResponseWriter, r *http.Request) {
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
