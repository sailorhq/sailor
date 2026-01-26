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
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	plugrpc "github.com/sailorhq/plug/sdk/proto"
	"github.com/sailorhq/sailor/internal/bige"
	"github.com/sailorhq/sailor/internal/types"
	v1 "github.com/sailorhq/sailor/pkg/core/v1"
	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
)

type DeployResourceRequest struct {
	Version string `json:"version"`
}

func (sc *SailorCore) DeployResourceHandler(ctx *fasthttp.RequestCtx) {
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

	var deployment DeployResourceRequest
	err := json.Unmarshal(ctx.Request.Body(), &deployment)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	if deployment.Version == "" {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "version is required"})
		return
	}

	intver, err := strconv.ParseUint(deployment.Version, 10, 32)
	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "version is not a number"})
		return
	}

	if _, ok := sc.dbconns[params.ProjectKey]; !ok {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "sailor project was not created"})
		return
	}

	claims := ctx.UserValue("__sailor_claims").(jwt.MapClaims)

	err = sc.dbconns[params.ProjectKey].Update(func(tx *bolt.Tx) error {
		resourceKey := params.Kind.ResourceKey(params.ResourceName)
		versionKey := fmt.Sprintf("%s_version", resourceKey)
		bigeVer := bige.ByteFromUInt32(uint32(intver))

		deploymentBucket := tx.Bucket([]byte(BUCKET_DEPLOYMENT)).Bucket([]byte(resourceKey))
		depBytes := deploymentBucket.Get(bigeVer)
		if depBytes == nil {
			return fmt.Errorf("no deployment with version: %s", deployment.Version)
		}

		metaBucket := tx.Bucket([]byte(BUCKET_META))
		if err := metaBucket.Put([]byte(versionKey), bigeVer); err != nil {
			return err
		}

		resourceBucket := tx.Bucket([]byte(BUCKET_RESOURCE))
		var resource v1.SailorResource
		resBytes := resourceBucket.Get([]byte(resourceKey))
		if err := json.Unmarshal(resBytes, &resource); err != nil {
			return err
		}
		if resBytes, err = json.Marshal(&resource); err != nil {
			return err
		}

		// start by deploying to plugs first
		// TODO: check if we need this
		// resourceName := fmt.Sprintf("%s-%s", params.App, resourceKey)
		var builtRes = buildResource(sc.dbconns[params.ProjectKey], resourceKey, versionKey, false, bigeVer)

		sigerr := sc.plugman.FireDeploy(&plugrpc.DeployRequest{
			Ns:          params.Ns,
			App:         params.App,
			Kind:        string(params.Kind),
			ResourceKey: resourceKey,
			Version:     uint32(intver),
			Content:     []byte(builtRes.Content),
		})
		// sigerr is not nil only when the FailurePolicy is "Fail"
		if sigerr != nil {
			return sigerr
		}

		// add to deployment history bucket
		depHistory, err := tx.CreateBucketIfNotExists(fmt.Appendf(nil, FMT_BUCKET_DEPLOYMENT_HISTORY, deployment.Version))
		if err != nil {
			return err
		}
		histbytes, err := json.Marshal(types.DeploymentHistory{
			Version:    deployment.Version,
			DeployedBy: claims["email"].(string),
			DeployedAt: time.Now().Format(time.RFC3339),
		})
		if err != nil {
			return err
		}

		return depHistory.Put(bige.ByteFromInt64(time.Now().Unix()), histbytes)
	})

	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	go sc.addAuditEvent(&v1.AuditEvent{
		Namespace: params.Ns,
		App:       params.App,
		Username:  claims["email"].(string),
		Action:    "deploy",
		Timestamp: time.Now(),
		Details: map[string]any{
			"deployed_version": deployment.Version,
		},
	})

	enc.Encode(ResponseMessage{Message: "ok"})
}
