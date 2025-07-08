package sailor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
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
	meta    types.SailorMeta
	configs atomic.Value
	secrets atomic.Value
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

	err := fetchState()
	if err != nil {
		return err
	}

	sailor.isConnected = true

	if !sailor.opts.DisableRefresh {
		go sleepAndRefresh()
	}

	return nil
}

func sleepAndRefresh() {
	time.Sleep(sailor.opts.RefreshTimeout)
	refresh()
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

	return !strings.EqualFold(sailor.state.meta.Version, string(b))
}

func refresh() {
	if shouldRefresh := checkStateVersion(); !shouldRefresh {
		sailor.log("sailor state version is same, not updating config")
		go sleepAndRefresh()
		return
	}

	sailor.log("refreshing sailor state...")
	err := fetchState()
	if err != nil {
		sailor.logf("error fetching state: %s", err.Error())
		sailor.sourceUnstable = true
	}

	sailor.logf("config: %+v\n", *sailor.state.configs.Load().(*map[string]any))
	sailor.logf("secrets: %+v\n", *sailor.state.secrets.Load().(*map[string][]byte))
	sailor.logf("version: %s\n", sailor.state.meta.Version)

	if !sailor.opts.DisableRefresh {
		go sleepAndRefresh()
		return
	}

	sailor.log("avoiding refresh as per options provided...")
}

func fetchState() error {
	url := fmt.Sprintf("%s/%s?ns=%s&app=%s", sailor.addr, state_api_path, sailor.ns, sailor.app)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		sailor.log(fmt.Sprintf("error creating request: %s", err.Error()))
		return err
	}

	req.Header.Set("x-access-key", sailor.opts.AccessKey)
	req.Header.Set("x-secret-key", sailor.opts.SecretKey)

	resp, err := sailor.sailorClient.Do(req)
	if err != nil {
		sailor.log(fmt.Sprintf("error fetching state: %s", err.Error()))
		return err
	}

	if resp.StatusCode != 200 {
		sailor.log("marking sailor source as unstable")
		sailor.sourceUnstable = true

		sailor.log("sailor is not reachable, fetching from backup")

		if sailor.opts.BackupURL == "" {
			msg := "no backup url set, waiting for sailor to be reachable"
			sailor.log(msg)
			return errors.New(msg)
		}

		return errors.New("unable to fetch state from sailor")
	}

	b, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		sailor.logf("cannot read response from sailor: %s", err.Error())
		return err
	}

	var state types.SailorState
	err = json.Unmarshal(b, &state)
	if err != nil {
		sailor.logf("sailor response is not in an expected format: %s", err.Error())
		return err
	}

	sailor.state.setState(state)

	return nil
}

func (is *internalState) setState(state types.SailorState) {
	var configs map[string]any
	err := json.Unmarshal(state.Config, &configs)
	if err != nil {
		// TODO :: log here that the config is flunked!!
		return
	}

	is.configs.Store(&configs)
	is.secrets.Store(&state.Secrets)

	// finally set the version so that we surely know that the state is set properly
	// without errors
	is.meta = types.SailorMeta{
		Version: state.Version,
	}
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

type ISailor interface {
	Get(key string) (any, error)
	GetSecret(key string) (string, error)
}

type SailorInstance struct {
	config  *map[string]any
	secrets *map[string][]byte
}

func (s *SailorInstance) Get(key string) (any, error) {
	var data any
	var ok bool
	if data, ok = (*s.config)[key]; !ok {
		return nil, fmt.Errorf("config key %s not found", key)
	}
	return data, nil
}

func (s *SailorInstance) GetSecret(key string) (string, error) {
	var data []byte
	var ok bool
	if data, ok = (*s.secrets)[key]; !ok {
		return "", fmt.Errorf("secret key %s not found", key)
	}

	claims, err := getClaims(string(data), sailor.opts.SecretKey)
	if err != nil {
		// TODO :: user friendly bug
		return "", err
	}
	return claims["data"].(string), nil
}

func Instance() ISailor {
	return &SailorInstance{
		config:  sailor.state.configs.Load().(*map[string]any),
		secrets: sailor.state.secrets.Load().(*map[string][]byte),
	}
}

func Refresh() {
	refresh()
}
