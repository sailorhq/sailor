package sailor

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	ENV_SAILOR_SECRET_KEY = "SAILOR_SECRET_KEY"
	ENV_SAILOR_BACKUP_URL = "SAILOR_BACKUP_URL"
)

const (
	state_api_path   = "api/v1/state"
	version_api_path = "api/v1/version"
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

	sailorClient *http.Client
}

var sailor = Sailor{
	addr:        "",
	ns:          "",
	app:         "",
	state:       internalState{},
	isConnected: false,
	sailorClient: &http.Client{
		Timeout: 10 * time.Second,
	},
}

func Connect(addr, ns, app string, opts ...types.SailorOpts) error {
	sailor.ns = ns
	sailor.app = app

	// check if opts are present
	if len(opts) > 0 {
		sailor.opts = &opts[0]

		// if refresh timeout is not set, set it to 5 seconds by default
		if sailor.opts.RefreshTimeout == 0 {
			sailor.opts.RefreshTimeout = 5 * time.Second
		}
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

	if sailor.opts.SecretKey == "" {
		envKey := os.Getenv(ENV_SAILOR_SECRET_KEY)
		if envKey != "" {
			sailor.opts.SecretKey = envKey
		}
	}

	if sailor.opts.AccessKey == "" || sailor.opts.SecretKey == "" {
		return fmt.Errorf("access key or secret key is not set, sailor cannot connect")
	}

	refresh(true)

	sailor.isConnected = true

	return nil
}

func sleepAndRefresh() {
	time.Sleep(sailor.opts.RefreshTimeout)
	refresh(false)
}

func checkStateVersion() bool {
	url := fmt.Sprintf("%s/%s?ns=%s&app=%s", sailor.addr, version_api_path, sailor.ns, sailor.app)
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

	return !strings.EqualFold(sailor.state.meta.Version, string(b))
}

func refresh(isInitiating bool) {
	if !isInitiating {
		if shouldRefresh := checkStateVersion(); !shouldRefresh {
			sailor.log("sailor state version is same, not updating config")
			go sleepAndRefresh()
			return
		}
	}

	sailor.log("refreshing sailor state...")
	url := fmt.Sprintf("%s/%s?ns=%s&app=%s", sailor.addr, state_api_path, sailor.ns, sailor.app)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		sailor.log(fmt.Sprintf("error creating request: %s", err.Error()))
		go sleepAndRefresh()
		return
	}

	req.Header.Set("x-access-key", sailor.opts.AccessKey)
	req.Header.Set("x-secret-key", sailor.opts.SecretKey)

	resp, err := sailor.sailorClient.Do(req)
	if err != nil {
		sailor.log(fmt.Sprintf("error fetching state: %s", err.Error()))
		go sleepAndRefresh()
		return
	}

	if resp.StatusCode != 200 {
		sailor.log("marking sailor source as unstable")
		sailor.sourceUnstable = true

		sailor.log("sailor is not reachable, fetching from backup")

		if sailor.opts.BackupURL == "" {
			sailor.log("no backup url set, waiting for sailor to be reachable")
			go sleepAndRefresh()
			return
		}

		return
	}

	b, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		sailor.logf("cannot read response from sailor: %s", err.Error())
		go sleepAndRefresh()
		return
	}

	// mark source as stable.. if it was set unstable before
	sailor.sourceUnstable = false

	var state types.SailorState
	err = json.Unmarshal(b, &state)
	if err != nil {
		sailor.logf("sailor response is not in an expected format: %s", err.Error())
		return
	}

	sailor.state.setState(state)

	if !sailor.opts.AvoidRefresh {
		go sleepAndRefresh()
		return
	}

	sailor.log("avoiding refresh as per options provided...")
}

func (is *internalState) setState(state types.SailorState) {
	is.Lock()
	var configs map[string]any
	err := json.Unmarshal(state.Config, &configs)
	if err != nil {
		// TODO :: log here that the config is flunked!!
		return
	}

	is.configs = configs
	is.secrets = make(map[string]string)
	for k, v := range state.Secrets {
		is.secrets[k] = string(v)
	}

	// finally set the version so that we surely know that the state is set properly
	// without errors
	is.meta = types.SailorMeta{
		Version: state.Version,
	}
	is.Unlock()
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

func (s *Sailor) logf(msg string, args ...any) {
	if sailor.opts.Logging {
		log.Printf(msg, args...)
	}
}

func (s *Sailor) log(msg string) {
	if sailor.opts.Logging {
		log.Println(msg)
	}
}

func (is *internalState) Release() {
	is.RUnlock()
}

func Instance() *internalState {
	sailor.state.RLock()
	return &sailor.state
}

func Refresh() {
	refresh(true)
}
