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
	"path/filepath"
	"regexp"
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

	err := validateK8sObjectNamingConvention(params.Ns, "namespace")
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	err = validateK8sObjectNamingConvention(params.Ns, "app")
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	if _, ok := sc.dbconns[params.ProjectKey]; ok {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "project already exists"})
		return
	}

	db, err := bolt.Open(filepath.Join(SAILS_FOLDER_PATH, fmt.Sprintf("%s%s", params.ProjectKey, DB_EXT)), 0600, nil)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	sc.dbconns[params.ProjectKey] = db

	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucket([]byte(BUCKET_META))
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

		return nil
	})

	if sc.SailorSail.CreateProject(params.Ns, params.App) != nil {
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	claims := ctx.UserValue("__sailor_claims").(jwt.MapClaims)
	go sc.addAuditEvent(&v1.AuditEvent{
		Namespace: params.Ns,
		App:       params.App,
		Username:  claims["email"].(string),
		Action:    "create_project",
		Timestamp: time.Now(),
	})

	enc.Encode(v1.ProjectResponse{
		Key: params.ProjectKey,
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

func validateK8sObjectNamingConvention(name string, paramName string) error {
	if len(name) < 3 || len(name) > 63 {
		return fmt.Errorf("%s must be between 3 and 63 characters", paramName)
	}
	if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(name) {
		return fmt.Errorf("%s must contain only lowercase letters, numbers, and hyphens", paramName)
	}
	return nil
}
