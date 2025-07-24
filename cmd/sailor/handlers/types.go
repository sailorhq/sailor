package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	bolt "go.etcd.io/bbolt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/codekidx/sailor/internal/types"
	diffmod "github.com/sergi/go-diff/diffmatchpatch"

	"github.com/go-playground/validator/v10"
)

type ResourceKind string

const (
	// -- ADMIN --
	BUCKET_ADMIN = "_admin"

	// BUCKET_PROJECTS will have list of all projects inside a
	// sailor instance, this bucket lives inside [BUCKET_ADMIN]
	BUCKET_PROJECTS = "projects"

	// BUCKET_USERS is used to store all the users and their
	// roles and permissions, this bucket lives inside [BUCKET_ADMIN]
	BUCKET_USERS = "users"

	// -- PROJECT --
	// BUCKET_META contains meta information about a project.
	// Each project will have a meta bucket and it will contain:
	//
	// 1. access key
	// 2. secret key
	// 3. current deployed version of all resources
	BUCKET_META = "_meta"

	// KEY_ACCESS_KEY has the access key for a sailor project
	// this key lives inside [BUCKET_META]
	KEY_ACCESS_KEY = "access_key"

	// KEY_SECRET_KEY has the secret key for a sailor project
	// which is mainly used to encrypt and decrypt a secret
	// resource
	KEY_SECRET_KEY = "secret_key"

	// BUCKET_RESOURCE contains all the resources present in a project
	BUCKET_RESOURCE = "resource"

	// BUCKET_DEPLOYMENT contains buckets for each resource
	// 		- config
	// 		- secret
	// 		- {key}-misc
	// and each sub-bucket will contain list of deployments per resource
	BUCKET_DEPLOYMENT = "deployments"

	// --- AUDIT ---
	// BUCKET_AUDIT contains audit trail of each action taken
	// by an user
	BUCKET_AUDIT = "_audit"

	// BUCKET_AUDIT_TRAIL is a bucket inside the AUDIT sail
	BUCKET_AUDIT_TRAIL = "audit_trail"

	BUCKET_BACKUP = "backup"

	KEY_S3 = "s3"

	// db extension used by sailor
	DB_EXT = "sail"
)

// -- RESOURCE TYPES INSIDE SAILOR --
const (
	KindConfig = "config"
	KindSecret = "secret"
	KindMisc   = "misc"
)

// DBUser is the user object stored in the database
type DBUser struct {
	Email       string   `json:"email"`
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
	AllowedApps []string `json:"allowed_apps"`
}

// User is the user object returned to the client
type User struct {
	Username    string   `json:"username"`
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
	AllowedApps []string `json:"allowed_apps"`
}

type SailorCore struct {
	// TODO :: change value to socket type
	dbconns map[string]*bolt.DB

	versions map[string]string

	kube *kubernetes.Clientset

	Log *zap.Logger
}

func initializeLogger() *zap.Logger {
	var logger *zap.Logger
	var productionLogging = os.Getenv("SAILOR_PROD_LOGGING")
	if productionLogging == "" {
		logCfg := zap.NewDevelopmentConfig()
		logCfg.EncoderConfig.EncodeTime = func(t time.Time, pae zapcore.PrimitiveArrayEncoder) {
			pae.AppendString(t.Format(time.UnixDate))
		}
		logCfg.DisableStacktrace = true
		logCfg.DisableCaller = true

		logger, _ = logCfg.Build()
	} else {
		logger, _ = zap.NewProduction()
	}

	return logger
}

func NewSailorCore() *SailorCore {
	sc := SailorCore{
		dbconns:  make(map[string]*bolt.DB),
		versions: make(map[string]string),
		Log:      initializeLogger(),
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

			sc.Log.Info("loaded project...", zap.String("name", projectKey))

			// if err := backup.BackupState(projectKey, db); err != nil {
			// 	return err
			// }

			// fmt.Println("uploaded state to s3 for: ", projectKey)

			// db.View(func(tx *bolt.Tx) error {
			// 	metaBucket := tx.Bucket([]byte(BUCKET_META))
			// 	version := metaBucket.Get([]byte(KEY_DEPLOYED_VERSION))
			// 	sc.versions[projectKey] = string(version)
			// 	return nil
			// })

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

	// Load in-cluster config
	config, err := rest.InClusterConfig()
	if err == nil {
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			sc.Log.Warn("error while initializing kubernetes client", zap.String("error", err.Error()))
			return nil
		}

		sc.kube = clientset
	} else {
		sc.Log.Warn("cannot get cluster config, is sailor running inside kubernetes?", zap.String("error", err.Error()))
		sc.Log.Warn("deploying resources to kubernetes will not work")
	}

	return &sc
}

func (sc *SailorCore) initInternalDatabase(dbName string) error {
	dbPath := fmt.Sprintf("./configs/%s.%s", dbName, DB_EXT)
	if f, _ := os.Stat(dbPath); f == nil {
		db, err := bolt.Open(dbPath, 0600, nil)
		if err != nil {
			sc.Log.Error(fmt.Sprintf("error during first load %s sail", dbName), zap.Error(err))
			return nil
		}

		sc.dbconns[dbName] = db

		// custom handling for admin db
		switch dbName {
		case BUCKET_ADMIN:
			db.Update(func(tx *bolt.Tx) error {
				_, err := tx.CreateBucketIfNotExists([]byte(BUCKET_BACKUP))
				if err != nil {
					return err
				}

				usersBucket, err := tx.CreateBucket([]byte(BUCKET_USERS))
				if err != nil {
					return err
				}

				user := DBUser{
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
		case BUCKET_AUDIT:
			db.Update(func(tx *bolt.Tx) error {
				_, err := tx.CreateBucketIfNotExists([]byte(BUCKET_AUDIT_TRAIL))
				return err
			})
		}

		return nil
	} else {
		db, err := bolt.Open(dbPath, 0600, nil)
		if err != nil {
			sc.Log.Error(fmt.Sprintf("error while loading %s sail", dbName), zap.Error(err))
			return nil
		}

		sc.dbconns[dbName] = db
	}

	sc.Log.Info(fmt.Sprintf("initialized %s sail", dbName))

	return nil
}

type SailorParams struct {
	ProjectKey    string
	Ns            string
	App           string
	AccessKey     string
	Body          []byte
	DeployVersion string

	DeploymentDescription string

	// auth params
	Username string
	Password string
}

type ResponseMessage struct {
	Message string `json:"message"`
}

// TODO :: make it a sailor core method
func hasRuleForAllKeys(mainMap, subMap map[string]any, parent string) error {
	// OPT :: instead of throwing error one by one , we can get all the missing rule keys
	// and then form a single error at one time
	for key := range mainMap {
		keyPath := fmt.Sprintf("%s.%s", parent, key)
		if _, ok := subMap[key]; !ok {
			return fmt.Errorf("rule for %s not present in schema", keyPath)
		} else if nestedMap, ok := mainMap[key].(map[string]any); ok {
			return hasRuleForAllKeys(nestedMap,
				subMap[key].(map[string]any),
				keyPath)
		}
	}
	return nil
}

// TODO :: make it a sailor core method
func validateWithRules(data, rules map[string]any) []string {
	validate := validator.New()

	fmt.Println("rules: ", rules)

	errMap := validate.ValidateMap(data, rules)
	if len(errMap) == 0 {
		return nil
	}

	fmt.Println("errMap: ", errMap)

	var messages = []string{}
	for field, verr := range errMap {
		if validationErrors, ok := verr.(validator.ValidationErrors); ok {
			for _, fieldErr := range validationErrors {
				// Default error message
				message := fmt.Sprintf("validation for '%s' failed on the '%s' rule\n", field, fieldErr.Tag())
				messages = append(messages, message)
			}
		} else {
			// Other error
			messages = append(messages, fmt.Sprintf("%s\n", verr))
		}
	}

	return messages
}

func buildResource(db *bolt.DB, resourceKey, versionKey string) string {
	var configJson string
	db.View(func(tx *bolt.Tx) error {
		deploymentBucket := tx.Bucket([]byte(BUCKET_DEPLOYMENT)).Bucket([]byte(resourceKey))
		metaBucket := tx.Bucket([]byte(BUCKET_META))

		min := []byte("0")
		max := metaBucket.Get([]byte(versionKey))
		// max is the current deployed version, if this is nil then
		// it means that this resource was never deployed
		if max == nil {
			return nil
		}

		cur := deploymentBucket.Cursor()

		differ := diffmod.New()

		for depVer, depBytes := cur.Seek(min); depVer != nil && bytes.Compare(depVer, max) <= 0; depVer, depBytes = cur.Next() {
			var deployment types.Deployment
			json.Unmarshal(depBytes, &deployment)
			// fmt.Println("diff_ver", string(diff_ver), string(max))
			p, _ := differ.PatchFromText(string(deployment.Diff))
			configJson, _ = differ.PatchApply(p, configJson)
		}

		return nil
	})

	return configJson
}
