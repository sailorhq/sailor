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
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/golang-jwt/jwt/v5"
	v1 "github.com/sailorhq/sailor/pkg/core/v1"
	"github.com/valyala/fasthttp"
)

type Project struct {
	Ns  string `json:"ns"`
	App string `json:"app"`
}

type FirstConfig struct {
	App string `json:"app"`
}

func (sc *SailorCore) CreateProjectHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	params := extractSailorParams(ctx)

	if params.Ns == "" || params.App == "" {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "namespace or app should not be empty"})
		return
	}

	if _, ok := sc.dbconns[params.ProjectKey]; ok {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "project already exists"})
		return
	}

	db, err := bolt.Open(fmt.Sprintf("./configs/%s.%s", params.ProjectKey, DB_EXT), 0600, nil)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	sc.dbconns[params.ProjectKey] = db

	var accessKey string
	var secretKey string
	err = db.Update(func(tx *bolt.Tx) error {
		var metaBucket *bolt.Bucket
		metaBucket, err := tx.CreateBucket([]byte(BUCKET_META))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucket([]byte(BUCKET_DEPLOYMENT))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucket([]byte(BUCKET_RESOURCE))
		if err != nil {
			return err
		}

		accessKey = generateUniqueKey("sailor", 16)
		if err = metaBucket.Put([]byte(KEY_ACCESS_KEY), []byte(accessKey)); err != nil {
			return err
		}

		secretKey = generateUniqueKey("secret", 32)
		return metaBucket.Put([]byte(KEY_SECRET_KEY), []byte(secretKey))
	})

	adminDB := sc.dbconns[BUCKET_ADMIN]
	adminDB.Update(func(tx *bolt.Tx) error {
		projectBucket := tx.Bucket([]byte(BUCKET_PROJECTS))
		projectBytes, err := json.Marshal(Project{Ns: params.Ns, App: params.App})
		if err != nil {
			return err
		}
		return projectBucket.Put([]byte(params.ProjectKey), projectBytes)
	})

	if err != nil {
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	claims := ctx.UserValue("__sailor_claims").(jwt.MapClaims)
	go sc.addAuditEvent(&AuditEvent{
		Namespace: params.Ns,
		App:       params.App,
		Username:  claims["email"].(string),
		Action:    "create_project",
		Timestamp: time.Now(),
	})

	enc.Encode(v1.ProjectResponse{
		Key:       params.ProjectKey,
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateUniqueKey(prefix string, length int) string {

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return fmt.Sprintf("%s-%s", prefix, string(b))
}
