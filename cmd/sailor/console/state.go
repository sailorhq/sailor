package console

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/codekidx/sailor/internal/types"
	bolt "go.etcd.io/bbolt"
)

func (c *Console) SailorStateHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)

	ak := strings.TrimSpace(r.Header.Get("x-access-key"))
	sk := strings.TrimSpace(r.Header.Get("x-secret-key"))

	if ak == "" || sk == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	params, err := c.extractSailorParams(r)
	if err != nil {
		enc.Encode(ResponseMessage{Message: err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	db, err := c.getDBConn(params)
	if err != nil {
		enc.Encode(ResponseMessage{Message: err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var state = types.SailorState{
		Secrets: make(map[string][]byte),
	}
	err = db.View(func(tx *bolt.Tx) error {
		metaBucket := tx.Bucket([]byte("_meta"))
		secretsBucket := tx.Bucket([]byte("secrets"))

		accessKey := metaBucket.Get([]byte(KEY_ACCESS_KEY))
		secretKey := metaBucket.Get([]byte(KEY_SECRET_KEY))

		if string(accessKey) != ak || string(secretKey) != sk {
			return errors.New("invalid access key or secret key")
		}

		secretsBucket.ForEach(func(k, v []byte) error {
			state.Secrets[string(k)] = v
			return nil
		})

		state.Version = string(metaBucket.Get([]byte("deploy_ver")))

		return nil
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	// builtConfig := buildResource(db, "", KEY_DEPLOYED_VERSION)
	// state.Config = []byte(builtConfig)

	enc.Encode(&state)
}
