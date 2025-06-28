package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) ValidateHandler(w http.ResponseWriter, r *http.Request) {
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

	db := sc.dbconns[BUCKET_ADMIN]

	enc := json.NewEncoder(w)

	var user User
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
		enc.Encode(ResponseMessage{Message: err.Error()})
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	enc.Encode(AuthResponse{
		Username:    user.Username,
		Permissions: user.Permissions,
		Roles:       user.Roles,
		Token:       tokenHeader,
	})
	w.WriteHeader(http.StatusOK)
}
