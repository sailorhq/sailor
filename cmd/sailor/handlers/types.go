package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	bolt "go.etcd.io/bbolt"
)

const (
	// buckets used by sailor
	BUCKET_ADMIN    = "_admin"
	BUCKET_META     = "_meta"
	BUCKET_AUDIT    = "_audit"
	BUCKET_USERS    = "users"
	BUCKET_PROJECTS = "projects"

	// buckets used by sailor apps
	BUCKET_CONFIGS    = "configs"
	BUCKET_SECRETS    = "secrets"
	BUCKET_DIFFS      = "_diffs"
	BUCKET_DEPLOYMENT = "deployments"

	// keys used by sailor
	KEY_ACCESS_KEY       = "access_key"
	KEY_SECRET_KEY       = "secret_key"
	KEY_DEPLOYED_VERSION = "deploy_ver"
	KEY_RULES            = "rules"

	// db extension used by sailor
	DB_EXT = "sail"
)

type User struct {
	Email       string   `json:"email"`
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
	AllowedApps []string `json:"allowed_apps"`
}

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

	if err := sc.initInternalDatabase(BUCKET_ADMIN); err != nil {
		return nil
	}

	err := sc.dbconns[BUCKET_ADMIN].Update(func(tx *bolt.Tx) error {
		projectsBucket, err := tx.CreateBucketIfNotExists([]byte(BUCKET_PROJECTS))
		if err != nil {
			return err
		}

		projectsBucket.ForEach(func(k []byte, v []byte) error {
			var project Project
			if err := json.Unmarshal(v, &project); err != nil {
				return err
			}

			projectKey := fmt.Sprintf("%s-%s", project.Ns, project.App)
			projectDbPath := fmt.Sprintf("./configs/%s-%s.%s", project.Ns, project.App, DB_EXT)
			db, err := bolt.Open(projectDbPath, 0600, nil)
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
		return nil
	})

	if err != nil {
		return nil
	}

	// load audit bucket
	if err = sc.initInternalDatabase(BUCKET_AUDIT); err != nil {
		return nil
	}

	return &sc
}

func (sc *SailorCore) initInternalDatabase(dbName string) error {
	dbPath := fmt.Sprintf("./configs/%s.%s", dbName, DB_EXT)
	if f, _ := os.Stat(dbPath); f == nil {
		db, err := bolt.Open(dbPath, 0600, nil)
		if err != nil {
			// TODO :: log error
			return nil
		}

		sc.dbconns[dbName] = db

		// custom handling for admin db
		if dbName == BUCKET_ADMIN {
			db.Update(func(tx *bolt.Tx) error {
				usersBucket, err := tx.CreateBucket([]byte(BUCKET_USERS))
				if err != nil {
					return err
				}

				user := User{
					Email:       "admin@sailor.com",
					Username:    "admin",
					Password:    "admin", // TODO :: hash password
					Permissions: []string{"admin"},
					Roles:       []string{"admin"},
				}
				json, err := json.Marshal(user)
				if err != nil {
					return err
				}

				return usersBucket.Put([]byte("admin"), json)
			})
		}

		return nil
	} else {
		db, err := bolt.Open(dbPath, 0600, nil)
		if err != nil {
			// TODO :: log error
			return nil
		}

		sc.dbconns[dbName] = db
	}

	return nil
}

// extractSailorParams extracts the sailor params from the request
// TODO ::it should not ideally give back errors, it should be a function
// which blindly extracts the params and returns a sailor params object
func (sc *SailorCore) extractSailorParams(r *http.Request) (*SailorParams, error) {
	ns := r.URL.Query().Get("ns")
	app := r.URL.Query().Get("app")
	deployVer := r.URL.Query().Get("deploy_ver")
	accessKey := r.URL.Query().Get("key")

	if ns == "" || app == "" {
		return nil, errors.New("namespace or app is empty")
	}

	username := r.Header.Get("x-username")
	password := r.Header.Get("x-password")

	return &SailorParams{
		ProjectKey:    fmt.Sprintf("%s-%s", ns, app),
		Ns:            ns,
		App:           app,
		AccessKey:     accessKey,
		DeployVersion: deployVer,
		Username:      username,
		Password:      password,
	}, nil
}

func (sc *SailorCore) getDBConn(params *SailorParams) (*bolt.DB, error) {
	key := fmt.Sprintf("%s-%s", params.Ns, params.App)

	if conn, ok := sc.dbconns[key]; ok {
		return conn, nil
	}

	dbpath := fmt.Sprintf("./configs/%s-%s.%s", params.Ns, params.App, DB_EXT)
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
	ProjectKey    string
	Ns            string
	App           string
	AccessKey     string
	Body          []byte
	DeployVersion string

	// auth params
	Username string
	Password string
}

type ResponseMessage struct {
	Message string
}
