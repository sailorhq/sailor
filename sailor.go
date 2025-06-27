package sailor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/codekidx/sailor/internal/types"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ENV_SAILOR_URL        = "SAILOR_URL"
	ENV_SAILOR_ACCESS_KEY = "SAILOR_ACCESS_KEY"
	ENV_SAILOR_BACKUP_URL = "SAILOR_BACKUP_URL"
)

type internalState struct {
	sync.RWMutex
	meta    types.SailorMeta
	configs map[string]any
	secrets map[string]string
}

type Sailor struct {
	addr string
	ns   string
	app  string
	opts *types.SailorOpts

	state internalState

	isConnected    bool
	sourceUnstable bool
}

var sailor = Sailor{
	addr: "",
	ns:   "",
	app:  "",
	opts: &types.SailorOpts{
		RefreshTimeout: 5 * time.Second,
	},
	state:       internalState{},
	isConnected: false,
}

func Connect(addr, ns, app string, opts ...types.SailorOpts) error {
	sailor.ns = ns
	sailor.app = app

	// check if opts are present
	if len(opts) > 0 {
		sailor.opts = &opts[0]
	} else {
		// defaults to 5 seconds
		sailor.opts = &types.SailorOpts{
			RefreshTimeout: 5 * time.Second,
		}
	}

	if addr == "" {
		envAddr := os.Getenv(ENV_SAILOR_URL)
		if envAddr != "" {
			sailor.addr = envAddr
		}
	} else {
		sailor.addr = addr
	}

	if sailor.opts.AccessKey == "" {
		envKey := os.Getenv(ENV_SAILOR_ACCESS_KEY)
		if envKey != "" {
			sailor.opts.AccessKey = envKey
		}
	}

	refresh(false)

	sailor.isConnected = true

	return nil
}

func sleepAndRefresh() {
	time.Sleep(sailor.opts.RefreshTimeout)
	refresh(true)
}

func checkStateVersion() bool {
	url := fmt.Sprintf("%s/version?ns=%s&app=%s&key=%s", sailor.addr, sailor.ns, sailor.app, sailor.opts.AccessKey)
	fmt.Println("url: ", url)
	resp, err := http.Get(url)
	if err != nil {
		return false
	}

	if resp.StatusCode != 200 {
		sailor.sourceUnstable = true

		// TODO :: log here that going to fetch from backup
		if sailor.opts.BackupURL == "" {
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
	sailor.sourceUnstable = false

	fmt.Println("string(b): ", string(b))

	return !strings.EqualFold(sailor.state.meta.Version, string(b))
}

func refresh(checkVersion bool) {
	fmt.Println("[REFRESH] trying to refresh config...")
	if checkVersion {
		if shouldRefresh := checkStateVersion(); !shouldRefresh {
			fmt.Println("[REFRESH] state version same, not updating config")
			go sleepAndRefresh()
			return
		}
	}

	url := fmt.Sprintf("%s/state?ns=%s&app=%s&key=%s", sailor.addr, sailor.ns, sailor.app, sailor.opts.AccessKey)
	resp, err := http.Get(url)
	if err != nil {
		go sleepAndRefresh()
		return
	}

	if resp.StatusCode != 200 {
		sailor.sourceUnstable = true

		// TODO :: log here that going to fetch from backup
		if sailor.opts.BackupURL == "" {
			go sleepAndRefresh()
			return
		}

		return
	}

	b, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		go sleepAndRefresh()
		return
	}

	// mark source as stable.. if it was set unstable before
	sailor.sourceUnstable = false

	var state types.SailorState
	err = json.Unmarshal(b, &state)
	if err != nil {
		// TODO :: log here that the refresh object is flunked!!
		return
	}

	sailor.state.Lock()
	sailor.state.meta = state.Meta
	sailor.state.configs = state.Configs
	sailor.state.secrets = state.Secrets
	sailor.state.Unlock()

	go sleepAndRefresh()
}

func (is *internalState) Get(key string) (value any, err error) {
	var ok bool
	if value, ok = is.configs[key]; !ok {
		err = fmt.Errorf("cannot find config %s", key)
		return value, err
	}
	return
}

func (is *internalState) GetDecode(key string, target *any) error {
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

func (is *internalState) GetSecret(key string) (value string, err error) {
	var (
		data string
		ok   bool
	)

	if data, ok = is.secrets[key]; !ok {
		return "", fmt.Errorf("secret %s not found", key)
	}

	// TODO :: get access key from sailor instance
	claims, err := getClaims(string(data), "")
	if err != nil {
		// TODO :: user friendly bug
		return "", err
	}
	return claims["data"].(string), nil
}

// Function to validate and extract map claims from the JWT string
func getClaims(secretStr string, accessKey string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(secretStr, func(token *jwt.Token) (any, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}
		return []byte(accessKey), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

func (is *internalState) Release() {
	is.RUnlock()
}

func Instance() *internalState {
	sailor.state.RLock()
	return &sailor.state
}
