// sailor
// Copyright (C) 2025 SailorHQ and Ashish Shekar (codekidX)

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sailorhq/sailor/cmd/sailor/sail"
	"github.com/sailorhq/sailor/internal/bige"
	"github.com/sailorhq/sailor/internal/signal"
	"github.com/sailorhq/sailor/internal/types"
	v1 "github.com/sailorhq/sailor/pkg/core/v1"
	diffmod "github.com/sergi/go-diff/diffmatchpatch"
	"github.com/valyala/fasthttp"

	"github.com/go-playground/validator/v10"
)

type ResourceKind string

func (rk ResourceKind) IsOneOf(kinds ...ResourceKind) bool {
	return slices.Contains(kinds, rk)
}

func (rk ResourceKind) IsConfig() bool {
	return rk == KindConfig
}

func (rk ResourceKind) IsSecret() bool {
	return rk == KindSecret
}

func (rk ResourceKind) IsMisc() bool {
	return rk == KindMisc
}

func (rk ResourceKind) ResourceKey(name string) string {
	if rk == KindMisc {
		return name
	}
	return string(rk)
}

const (
	SAILS_FOLDER_PATH = "./sails"

	// -- ADMIN --
	BUCKET_ADMIN = "_admin"

	// BUCKET_PROJECTS will have list of all projects inside a
	// sailor instance, this bucket lives inside [BUCKET_ADMIN]
	BUCKET_PROJECTS = "projects"

	// BUCKET_USERS is used to store all the users and their
	// roles and permissions, this bucket lives inside [BUCKET_ADMIN]
	BUCKET_USERS = "users"

	// BUCKET_SETTING is collection of sailor wide settings this
	// buckect lives inside [BUCKET_ADMIN]
	BUCKET_SETTING = "settings"

	// KEY_SETTING is used to save sailor wide settings this
	// key lives inside [BUCKET_ADMIN]
	KEY_SETTING = "settings"

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

	// BUCKET_RELEASE contains all the release tags provided while creating a deployment
	BUCKET_RELEASE = "release"

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
	DB_EXT = ".sail"

	// decoded sailor claims from request
	SAILOR_CLAIMS_KEY = "__sailor_claims"
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
	Password    [32]byte `json:"password"`
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
	AllowedApps []string `json:"allowed_apps"`
	Token       string   `json:"token"`
	Fingerprint string   `json:"fp"`
}

// User is the user object returned to the client
type User struct {
	Email       string   `json:"email"`
	Password    string   `json:"pass"`
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
	AllowedApps []string `json:"allowed_apps"`
}

type SailorCore struct {
	dbconns map[string]*bolt.DB

	versions map[string]string

	setting *v1.SailorSetting

	Log *zap.Logger

	// TODO :: we need to abstract all sail operations behind this
	// single interface
	SailorSail sail.Sail

	plugman *signal.PlugManager
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
	log := initializeLogger()
	sc := SailorCore{
		dbconns:  make(map[string]*bolt.DB),
		versions: make(map[string]string),
		Log:      log,
		SailorSail: &sail.CoreSail{
			ProjectMap: make(map[string]*bolt.DB),
		},
		plugman: signal.NewPlugManager(log),
	}

	if f, _ := os.Stat(SAILS_FOLDER_PATH); f == nil {
		sc.Log.Info("sails folder not found, creating it...")
		if err := os.Mkdir(SAILS_FOLDER_PATH, 0755); err != nil {
			sc.Log.Fatal("unable to create sails/ folder, sailor cannot run without it, please create it manually", zap.Error(err))
		}
	}

	if err := sc.initInternalDatabase(BUCKET_ADMIN); err != nil {
		sc.Log.Error("error while initializing admin sail", zap.Error(err))
		return nil
	}

	if err := sc.initInternalDatabase(BUCKET_META); err != nil {
		sc.Log.Error("error while initializing meta sail", zap.Error(err))
		return nil
	}

	sc.Log.Info("trying to get sailor settings")
	ss, err := getSailorSetting(sc.dbconns[BUCKET_ADMIN])
	if err != nil {
		// sailor exits here... because tokenkey is required for authentication
		sc.Log.Fatal("unable to load settings", zap.Error(err))
	}
	sc.setting = ss

	err = sc.loadSailFiles()
	if err != nil {
		sc.Log.Error("error while loading sails", zap.Error(err))
		return nil
	}

	// load audit bucket
	if err = sc.initInternalDatabase(BUCKET_AUDIT); err != nil {
		sc.Log.Error("error while initializing audit sail", zap.Error(err))
		return nil
	}

	// load all plugs
	sc.plugman.Load(ss.Rxs)

	return &sc
}

func (sc *SailorCore) loadSailFiles() (err error) {
	// TODO :: add ability to set folder from ENV variable

	err = sc.dbconns[BUCKET_META].View(func(tx *bolt.Tx) error {
		projectBucket := tx.Bucket([]byte(BUCKET_PROJECTS))
		if projectBucket == nil {
			return errors.New("projects bucket not found")
		}

		projerr := projectBucket.ForEach(func(k, _ []byte) error {
			projectKey := string(k)
			if err := sc.loadProject(SAILS_FOLDER_PATH, projectKey); err != nil {
				return err
			}
			sc.Log.Info("loaded project...", zap.String("name", projectKey))

			return nil
		})
		if projerr != nil {
			return projerr
		}
		return nil
	})

	sc.performProjectRecon(SAILS_FOLDER_PATH)

	return nil
}

func (sc *SailorCore) loadProject(sailFolder, projectKey string) error {
	dbPath := filepath.Join(sailFolder, fmt.Sprintf("%s%s", projectKey, DB_EXT))

	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		sc.Log.Error("error while loading sail", zap.String("path", dbPath), zap.Error(err))
		return err
	}

	sc.dbconns[projectKey] = db
	sc.SailorSail.(*sail.CoreSail).ProjectMap[projectKey] = db
	return nil
}

// perfromProjectRecon performs a recon activity to check all the projects are loaded
// and if any project is not loaded, it will do log and load it inside our _meta bucket
func (sc *SailorCore) performProjectRecon(sailFolder string) {
	sc.Log.Info("starting project recon...")

	err := filepath.Walk(sailFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || filepath.Ext(path) != DB_EXT {
			return nil
		}

		// ignore internal sail files
		if strings.HasPrefix(info.Name(), "_") {
			return nil
		}

		projectKey := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
		if _, ok := sc.dbconns[projectKey]; ok {
			return nil
		}

		sc.Log.Info("project is being loaded during recon", zap.String("project", projectKey))

		if err := sc.loadProject(sailFolder, projectKey); err != nil {
			return err
		}

		sc.Log.Info("project is loaded during recon", zap.String("project", projectKey))

		// in order to not stat and load during the next start of sailor core we will add it to
		// _meta bucket
		splitted := strings.SplitN(projectKey, "_", 2)
		if len(splitted) != 2 {
			sc.Log.Warn("project key is not in expected format", zap.String("project", projectKey))
			return nil
		}

		if err := sc.SailorSail.CreateProject(splitted[0], splitted[1]); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		sc.Log.Error("error while reconing projects", zap.Error(err))
	}

	sc.Log.Info("recon process ended")
}

func (sc *SailorCore) initInternalDatabase(dbName string) error {
	dbPath := fmt.Sprintf("./sails/%s%s", dbName, DB_EXT)
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

				// initialize basic setting, it creates a token key for creating
				// sailor token
				settingBucket, err := tx.CreateBucket([]byte(BUCKET_SETTING))
				if err != nil {
					return err
				}
				setting := v1.SailorSetting{
					AccessKey: generateUniqueKey("sailor", 16),
					SecretKey: generateUniqueKey("secret", 32),
					TokenKey:  generateUniqueKey("token", 16),
				}
				settingb, _ := json.Marshal(&setting)
				if err := settingBucket.Put([]byte(KEY_SETTING), settingb); err != nil {
					return err
				}
				sc.setting = &setting

				// initialize user related things...
				usersBucket, err := tx.CreateBucket([]byte(BUCKET_USERS))
				if err != nil {
					return err
				}

				user := DBUser{
					Email:       "admin@super.sailor",
					Username:    "admin",
					Password:    sha256.Sum256([]byte("admin")),
					Permissions: []string{"super:*"},
					Roles:       []string{"admin"},
				}
				json, err := json.Marshal(user)
				if err != nil {
					return err
				}

				return usersBucket.Put([]byte(user.Email), json)
			})

			// TODO :: there might be a better way
			sc.SailorSail.(*sail.CoreSail).Admin = db
		case BUCKET_AUDIT:
			db.Update(func(tx *bolt.Tx) error {
				_, err := tx.CreateBucketIfNotExists([]byte(BUCKET_AUDIT_TRAIL))
				return err
			})

			// TODO :: there might be a better way
			sc.SailorSail.(*sail.CoreSail).Audit = db
		case BUCKET_META:
			db.Update(func(tx *bolt.Tx) error {
				_, err := tx.CreateBucketIfNotExists([]byte(BUCKET_PROJECTS))
				return err
			})
		}

		sc.SailorSail.(*sail.CoreSail).Meta = db
		return nil
	} else {
		db, err := bolt.Open(dbPath, 0600, nil)
		if err != nil {
			sc.Log.Error(fmt.Sprintf("error while loading %s sail", dbName), zap.Error(err))
			return nil
		}

		sc.dbconns[dbName] = db

		// if sail files are already there load it inside our interface
		switch dbName {
		case BUCKET_ADMIN:
			sc.SailorSail.(*sail.CoreSail).Admin = db
		case BUCKET_AUDIT:
			sc.SailorSail.(*sail.CoreSail).Audit = db
		case BUCKET_META:
			sc.SailorSail.(*sail.CoreSail).Meta = db
		}
	}

	sc.Log.Info(fmt.Sprintf("initialized %s sail", dbName))

	return nil
}

func (sc *SailorCore) GetConsoleHosts() []string {
	if sc.setting.Console != nil {
		return sc.setting.Console.Hosts
	}
	return []string{}
}

type SailorParams struct {
	ProjectKey string
	Ns         string
	App        string

	// ResourceName is used incase of Misc Resource
	ResourceName string
	Kind         ResourceKind
	Body         []byte

	// authentication
	Token string

	// deployment related
	RequestedVersion string
}

func extractSailorParams(ctx *fasthttp.RequestCtx) SailorParams {
	sp := SailorParams{
		Body:  ctx.Request.Body(),
		Token: string(ctx.Request.Header.Peek("x-token")),
	}

	if ns, ok := ctx.UserValue("namespace").(string); ok {
		sp.Ns = ns
	}

	if app, ok := ctx.UserValue("app").(string); ok {
		sp.App = app
	}

	if sp.Ns != "" && sp.App != "" {
		sp.ProjectKey = fmt.Sprintf("%s_%s", sp.Ns, sp.App)
	}

	// kind is not available while creating a project
	if k, ok := ctx.UserValue("kind").(string); ok {
		sp.Kind = ResourceKind(k)
	}

	// name produces nil value if the path does not contain name, so
	// we explicity check if name is castable to string
	if n, ok := ctx.UserValue("name").(string); ok {
		sp.ResourceName = n
	}

	return sp
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
			return hasRuleForAllKeys(nestedMap, subMap[key].(map[string]any), keyPath)
		}
	}
	return nil
}

// TODO :: make it a sailor core method
func validateWithRules(data, rules map[string]any) []string {
	validate := validator.New()

	errMap := validate.ValidateMap(data, rules)
	if len(errMap) == 0 {
		return nil
	}

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

type BuiltResource struct {
	Content string `json:"content"`
	Version uint32 `json:"version"`
}

func buildResource(db *bolt.DB, resourceKey, versionKey string,
	onTopOfLastDeployment bool,
	overrideMaxVersion []byte) BuiltResource {
	var configJson string
	var max []byte

	db.View(func(tx *bolt.Tx) error {
		deploymentBucket := tx.Bucket([]byte(BUCKET_DEPLOYMENT)).Bucket([]byte(resourceKey))
		metaBucket := tx.Bucket([]byte(BUCKET_META))

		min := bige.ByteFromUInt32(0)

		if overrideMaxVersion != nil {
			max = overrideMaxVersion
		} else {
			if onTopOfLastDeployment {
				max, _ = deploymentBucket.Cursor().Last()
			} else {
				max = metaBucket.Get([]byte(versionKey)) // also returns big-endian bytes back
			}
		}

		// if the key is still nil from deployment bucket, the reosurce was neither
		// deployed nor had any deployments created
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

	return BuiltResource{
		Content: configJson,
		Version: bige.UInt32FromByte(max),
	}
}

// TODO :: the accesskey here is sailor level, so it should be set by Super Admin!
func jwtFromClaims(claims jwt.MapClaims, accessKey string) (string, error) {
	// Create a new token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, getMapClaims(claims))
	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(accessKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func getMapClaims(fromClaims map[string]any) jwt.MapClaims {
	mc := jwt.MapClaims{}
	for k, v := range fromClaims {
		mc[k] = v
	}
	return mc
}

func createSailorURI(ns, app, ak, sk, host string) string {
	return fmt.Sprintf("sailor://%s:%s@%s/%s/%s", ak, sk, host, ns, app)
}
