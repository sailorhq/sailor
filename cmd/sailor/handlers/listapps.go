package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/golang-jwt/jwt/v5"
	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) ListAppsHandler(w http.ResponseWriter, r *http.Request) {
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

	db := sc.dbconns[BUCKET_ADMIN]

	var user User
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
		}

		return nil
	})

	enc := json.NewEncoder(w)
	enc.Encode(user.AllowedApps)
}
