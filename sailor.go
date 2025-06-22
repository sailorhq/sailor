package sailor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/codekidx/sailor/internal/types"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ENV_SAILOR_URL        = "SAILOR_URL"
	ENV_SAILOR_ACCESS_KEY = "SAILOR_ACCESS_KEY"
	ENV_SAILOR_BACKUP_URL = "SAILOR_BACKUP_URL"
)

type Sailor struct {
	addr      string
	ns        string
	app       string
	accessKey string
	opts      *types.SailorOpts

	state types.SailorState

	sourceUnstable bool
}

func (s *Sailor) Connect(ns, app string) error {
	s.ns = ns
	s.app = app

	s.refresh(false)

	return nil
}

func (s *Sailor) sleepAndRefresh() {
	time.Sleep(s.opts.RefreshTimeout)
	s.refresh(true)
}

func (s *Sailor) checkStateVersion() bool {
	url := fmt.Sprintf("%s/version?ns=%s&app=%s&key=%s", s.addr, s.ns, s.app, s.accessKey)
	resp, err := http.Get(url)
	if err != nil {
		return false
	}

	if resp.StatusCode != 200 {
		s.sourceUnstable = true

		// TODO :: log here that going to fetch from backup
		if s.opts.BackupURL == "" {
			return false
		}

		return false
	}

	b, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return false
	}

	// mark source as stable.. if it was set unstable before
	s.sourceUnstable = false

	fmt.Println("string(b)", s.state.Meta.Version, string(b))
	fmt.Println("s.state.Meta.Version", s.state.Meta.Version)

	return !strings.EqualFold(s.state.Meta.Version, string(b))
}

func (s *Sailor) refresh(checkVersion bool) {
	fmt.Println("[REFRESH] trying to refresh config...")
	if checkVersion {
		if shouldRefresh := s.checkStateVersion(); !shouldRefresh {
			fmt.Println("state version same, not updating config")
			go s.sleepAndRefresh()
			return
		}
	}

	url := fmt.Sprintf("%s/state?ns=%s&app=%s&key=%s", s.addr, s.ns, s.app, s.accessKey)
	resp, err := http.Get(url)
	if err != nil {
		go s.sleepAndRefresh()
		return
	}

	if resp.StatusCode != 200 {
		s.sourceUnstable = true

		// TODO :: log here that going to fetch from backup
		if s.opts.BackupURL == "" {
			go s.sleepAndRefresh()
			return
		}

		return
	}

	b, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		go s.sleepAndRefresh()
		return
	}

	// mark source as stable.. if it was set unstable before
	s.sourceUnstable = false

	var state types.SailorState
	err = json.Unmarshal(b, &state)
	if err != nil {
		// TODO :: log here that the refresh object is flunked!!
		return
	}

	s.state.Lock()
	s.state.Meta = state.Meta
	s.state.Configs = state.Configs
	s.state.Secrets = state.Secrets
	s.state.Unlock()

	go s.sleepAndRefresh()
}

func (s *Sailor) Get(key string) (value any, err error) {
	var ok bool
	if value, ok = s.state.Configs[key]; !ok {
		err = fmt.Errorf("cannot find config %s", key)
		return value, err
	}
	return
}

func (s *Sailor) GetDecode(key string, target *any) error {
	// TODO :: maybe check if we want to use mapstructure module here!
	// var data any
	// var ok bool
	// if data, ok = s.configs[key]; !ok {
	// 	return fmt.Errorf("config key %s not found", key)
	// }

	// err := json.Unmarshal(data, target)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (s *Sailor) GetSecret(key string) (value string, err error) {
	var (
		data string
		ok   bool
	)

	if data, ok = s.state.Secrets[key]; !ok {
		return "", fmt.Errorf("secret %s not found", key)
	}

	claims, err := s.getClaims(string(data))
	if err != nil {
		// TODO :: user friendly bug
		return "", err
	}
	return claims["data"].(string), nil
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

func New(addr, key string, opts ...types.SailorOpts) *Sailor {
	if addr == "" {
		envAddr := os.Getenv(ENV_SAILOR_URL)
		if envAddr != "" {
			addr = envAddr
		}
	}

	if key == "" {
		envKey := os.Getenv(ENV_SAILOR_ACCESS_KEY)
		if envKey != "" {
			key = envKey
		}
	}

	s := Sailor{addr: addr, accessKey: key}
	if len(opts) > 0 {
		s.opts = &opts[0]
	} else {
		// defaults to 5 seconds
		s.opts = &types.SailorOpts{
			RefreshTimeout: 5 * time.Second,
		}
	}
	return &s
}
