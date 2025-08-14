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
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	v1 "github.com/sailorhq/sailor/pkg/core/v1"
	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) CreateResourceHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	params := extractSailorParams(ctx)

	if params.Ns == "" || params.App == "" || params.Kind == "" {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "namespace or app should not be empty"})
		return
	}

	if !params.Kind.IsOneOf(KindConfig, KindMisc, KindSecret) {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "unknown resource kind"})
		return
	}

	if params.Kind.IsMisc() && params.ResourceName == "" {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "missing resource name"})
		return
	}

	if _, ok := sc.dbconns[params.ProjectKey]; !ok {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "no such project in sailor"})
		return
	}

	var resource v1.SailorResource
	err := json.Unmarshal(ctx.Request.Body(), &resource)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	// if setting is not present add default setting
	if resource.Setting == nil {
		resource.Setting = &v1.ResourceSetting{
			Deploy: v1.DeploySetting{K8s: false},
			Schema: v1.SchemaSetting{
				Strict: false,
			},
		}
	}

	err = sc.dbconns[params.ProjectKey].Update(func(tx *bolt.Tx) error {
		resourceKey := params.Kind.ResourceKey(params.ResourceName)

		resourceBucket := tx.Bucket([]byte(BUCKET_RESOURCE))
		resBytes := resourceBucket.Get([]byte(resourceKey))
		if resBytes != nil {
			return fmt.Errorf("%s is already created", resourceKey)
		}

		deploymentBucket := tx.Bucket([]byte(BUCKET_DEPLOYMENT))
		if _, err := deploymentBucket.CreateBucket([]byte(resourceKey)); err != nil {
			return err
		}

		res, err := json.Marshal(&resource)
		if err != nil {
			return err
		}

		return resourceBucket.Put([]byte(resourceKey), res)
	})

	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	claims := ctx.UserValue("__sailor_claims").(jwt.MapClaims)
	go sc.addAuditEvent(&AuditEvent{
		Namespace: params.Ns,
		App:       params.App,
		Username:  claims["email"].(string),
		Action:    "create_resource",
		Timestamp: time.Now(),
		Details: map[string]any{
			"kind": params.Kind,
		},
	})

	enc.Encode(ResponseMessage{Message: "ok"})

}
