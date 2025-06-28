package handlers

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func (sc *SailorCore) AddSecretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func jwtfyit(data string, accessKey string, expiry string) (string, error) {
	// Create a new token
	token := jwt.New(jwt.SigningMethodHS256)
	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["data"] = data

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(accessKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
