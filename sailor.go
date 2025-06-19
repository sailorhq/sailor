package sailor

import (
	"encoding/json"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type Sailor struct {
	addr      string
	ns        string
	app       string
	accessKey string

	secrets map[string]string
	configs map[string][]byte
}

func (s *Sailor) Connect(ns, app string) error {
	s.ns = ns
	s.app = app

	// TODO :: try to connect to the websocket path for this namespace and app
	// TODO :: need to handle key set as well, via websocket

	// TODO :: on fail check for S3/Redis fallback

	return nil
}

func (s *Sailor) Get(key string) (value []byte, err error) {
	var ok bool
	if value, ok = s.configs[key]; !ok {
		err = fmt.Errorf("cannot find config %s", key)
		return value, err
	}
	return
}

func (s *Sailor) GetDecode(key string, target *any) error {
	var data []byte
	var ok bool
	if data, ok = s.configs[key]; !ok {
		return fmt.Errorf("config key %s not found", key)
	}

	err := json.Unmarshal(data, target)
	if err != nil {
		return err
	}
	return nil
}

func (s *Sailor) GetSecret(key string) (value string, err error) {
	var (
		data string
		ok   bool
	)

	if data, ok = s.secrets[key]; !ok {
		return "", fmt.Errorf("secret %s not found", key)
	}

	claims, err := s.getClaims(string(data))
	if err != nil {
		// TODO :: user friendly bug
		return "", err
	}
	return claims["data"].(string), nil
}

func New(addr, key string) *Sailor {
	return &Sailor{addr: addr, accessKey: key}
}

// Function to validate and extract map claims from the JWT string
func (s *Sailor) getClaims(secretStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(secretStr, func(token *jwt.Token) (any, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}
		return []byte(s.accessKey), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
