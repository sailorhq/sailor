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
package console

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	bolt "go.etcd.io/bbolt"
)

type ResourceKind string

type Project struct {
	Ns  string `json:"ns"`
	App string `json:"app"`
}

type FirstConfig struct {
	App string `json:"app"`
}

const (
	// buckets used by sailor
	BUCKET_ADMIN       = "_admin"
	BUCKET_META        = "_meta"
	BUCKET_AUDIT       = "_audit"
	BUCKET_USERS       = "users"
	BUCKET_PROJECTS    = "projects"
	BUCKET_AUDIT_TRAIL = "audit_trail"

	// buckets used by sailor apps
	BUCKET_RESOURCE   = "resource"
	BUCKET_CONFIGS    = "configs"
	BUCKET_SECRETS    = "secrets"
	BUCKET_MISC       = "misc"
	BUCKET_DEPLOYMENT = "deployments"
	BUCKET_BACKUP     = "backup"

	// kind of resource in sailor
	KindConfig = "config"
	KindSecret = "secret"
	KindMisc   = "misc"

	// keys used by sailor
	KEY_ACCESS_KEY       = "access_key"
	KEY_SECRET_KEY       = "secret_key"
	KEY_DEPLOYED_VERSION = "deploy_ver"
	KEY_RULES            = "rules"
	KEY_S3               = "s3"

	// db extension used by sailor
	DB_EXT = "sail"
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

// extractSailorParams extracts the sailor params from the request
// TODO ::it should not ideally give back errors, it should be a function
// which blindly extracts the params and returns a sailor params object
func (c *Console) extractSailorParams(r *http.Request) (*SailorParams, error) {
	ns := r.URL.Query().Get("ns")
	app := r.URL.Query().Get("app")
	deployVer := r.URL.Query().Get("deploy_ver")
	accessKey := r.URL.Query().Get("key")
	deploymentDescription := r.URL.Query().Get("d_desc")

	if ns == "" || app == "" {
		return nil, errors.New("namespace or app is empty")
	}

	username := r.Header.Get("x-username")
	password := r.Header.Get("x-password")

	return &SailorParams{
		ProjectKey:            fmt.Sprintf("%s-%s", ns, app),
		Ns:                    ns,
		App:                   app,
		AccessKey:             accessKey,
		DeployVersion:         deployVer,
		Username:              username,
		Password:              password,
		DeploymentDescription: deploymentDescription,
	}, nil
}

func (c *Console) getDBConn(params *SailorParams) (*bolt.DB, error) {
	key := fmt.Sprintf("%s-%s", params.Ns, params.App)

	if conn, ok := c.dbconns[key]; ok {
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

	c.dbconns[key] = db
	return db, nil
}
