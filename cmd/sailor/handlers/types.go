package handlers

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	bolt "go.etcd.io/bbolt"
)

const (
	BUCKET_META       = "_meta"
	BUCKET_CONFIGS    = "configs"
	BUCKET_SECRETS    = "secrets"
	BUCKET_DIFFS      = "_diffs"
	BUCKET_DEPLOYMENT = "deployments"

	KEY_ACCESS_KEY       = "access_key"
	KEY_DEPLOYED_VERSION = "deploy_ver"
	KEY_RULES            = "rules"
)

type SailorCore struct {
	// TODO :: change value to socket type
	dbconns map[string]*bolt.DB

	versions map[string]string
}

func NewSailorCore() *SailorCore {
	sc := SailorCore{
		dbconns:  make(map[string]*bolt.DB),
		versions: make(map[string]string),
	}

	err := filepath.Walk("./configs", func(path string, info fs.FileInfo, e error) error {
		if info.IsDir() || e != nil {
			return nil
		}

		fileName := filepath.Base(path)
		projectKey := strings.ReplaceAll(fileName, filepath.Ext(fileName), "")

		db, err := bolt.Open(path, 0600, nil)
		if err != nil {
			return err
		}

		sc.dbconns[projectKey] = db

		db.View(func(tx *bolt.Tx) error {
			metaBucket := tx.Bucket([]byte(BUCKET_META))
			version := metaBucket.Get([]byte(KEY_DEPLOYED_VERSION))
			sc.versions[projectKey] = string(version)
			return nil
		})

		return nil
	})

	if err != nil {
		return nil
	}

	return &sc
}

func (sc *SailorCore) extractSailorParams(r *http.Request) (*SailorParams, error) {
	ns := r.URL.Query().Get("ns")
	app := r.URL.Query().Get("app")
	deployVer := r.URL.Query().Get("deploy_ver")
	accessKey := r.URL.Query().Get("key")

	if ns == "" || app == "" {
		return nil, errors.New("namespace or app is empty")
	}

	return &SailorParams{
		Ns:            ns,
		App:           app,
		AccessKey:     accessKey,
		DeployVersion: deployVer,
	}, nil
}

func (sc *SailorCore) getDBConn(params *SailorParams) (*bolt.DB, error) {
	key := fmt.Sprintf("%s-%s", params.Ns, params.App)

	if conn, ok := sc.dbconns[key]; ok {
		return conn, nil
	}

	dbpath := fmt.Sprintf("./configs/%s-%s.sail", params.Ns, params.App)
	if f, _ := os.Stat(dbpath); f == nil {
		return nil, errors.New("app not present in this namespace")
	}

	db, err := bolt.Open(dbpath, 0600, nil)
	if err != nil {
		panic(err)
	}

	sc.dbconns[key] = db
	return db, nil
}

type SailorParams struct {
	Ns            string
	App           string
	AccessKey     string
	Body          []byte
	DeployVersion string
}

type ResponseMessage struct {
	Message string
}
