package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	bolt "go.etcd.io/bbolt"
)

type AuthResponse struct {
	Username    string   `json:"username"`
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
	Token       string   `json:"token"`
}

func (sc *SailorCore) AuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	enc := json.NewEncoder(w)

	username := r.Header.Get("x-username")
	password := r.Header.Get("x-password")

	db := sc.dbconns[BUCKET_ADMIN]

	var user User
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
		enc.Encode(ResponseMessage{Message: err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := jwtfyit(user.Password, "TODO :: access key", time.Now().Add(time.Hour*24).Format(time.RFC3339))
	if err != nil {
		enc.Encode(ResponseMessage{Message: err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	enc.Encode(AuthResponse{
		Username:    user.Username,
		Permissions: user.Permissions,
		Roles:       user.Roles,
		Token:       token,
	})
}
