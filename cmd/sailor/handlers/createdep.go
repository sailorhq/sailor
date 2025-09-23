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
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sailorhq/sailor/internal/bige"
	"github.com/sailorhq/sailor/internal/types"
	v1 "github.com/sailorhq/sailor/pkg/core/v1"
	"github.com/sailorhq/sailor/pkg/vault"
	"github.com/valyala/fasthttp"

	diffmod "github.com/sergi/go-diff/diffmatchpatch"
)

type CreateDeploymentRequest struct {
	Description string            `json:"desc"`
	ConfigData  map[string]any    `json:"config_data"`
	MiscData    string            `json:"misc_data"`
	SecretData  map[string]string `json:"secret_data"`
}

func (sc *SailorCore) CreateDeploymentHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	params := extractSailorParams(ctx)

	if params.Ns == "" || params.App == "" || params.Kind == "" {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "namespace or app should not be empty"})
		return
	}

	if params.Kind != KindConfig && params.Kind != KindSecret && params.Kind != KindMisc {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "unknown resource kind"})
		return
	}

	if params.Kind == KindMisc && params.ResourceName == "" {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "missing resource name"})
		return
	}

	var deployment CreateDeploymentRequest
	err := json.Unmarshal(params.Body, &deployment)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	if deployment.Description == "" {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "description is required to create a deployment"})
		return
	}

	if params.Kind == KindMisc && deployment.MiscData == "" {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "cannot create deployment with empty string_data for kind misc"})
		return
	}

	if len(deployment.ConfigData) == 0 && len(deployment.SecretData) == 0 && len(deployment.MiscData) == 0 {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "cannot create deployment with empty data"})
		return
	}

	if _, ok := sc.dbconns[params.ProjectKey]; !ok {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "sailor project was not created"})
		return
	}

	resourceKey := params.Kind.ResourceKey(params.ResourceName)
	differ := diffmod.New()

	var version uint32 = 1
	claims := ctx.UserValue("__sailor_claims").(jwt.MapClaims)
	err = sc.dbconns[params.ProjectKey].Update(func(tx *bolt.Tx) error {
		resourceBucket := tx.Bucket([]byte(BUCKET_RESOURCE))
		resBytes := resourceBucket.Get([]byte(resourceKey))
		if resBytes == nil {
			return fmt.Errorf("%s resource not created", resourceKey)
		}

		var resource v1.SailorResource
		if err := json.Unmarshal(resBytes, &resource); err != nil {
			return err
		}

		// TODO :: make schema work for secrets as well
		// if strict schema validation is enabled then we check it!
		if resource.Setting.Schema.Strict && params.Kind.IsConfig() {
			if err := hasRuleForAllKeys(deployment.ConfigData, resource.Schema, "$root"); err != nil {
				return err
			}
			if errs := validateWithRules(deployment.ConfigData, resource.Schema); len(errs) > 0 {
				// TODO :: we need to define a proper error struct with the validation
				// `errs` as []string!
				return errors.New(strings.Join(errs, ","))
			}
		}

		deploymentBucket := tx.Bucket([]byte(BUCKET_DEPLOYMENT)).Bucket([]byte(resourceKey))
		if deploymentBucket == nil {
			return errors.New("deployment not available, was the resource created?")
		}

		var resourceData = deployment.MiscData
		if !params.Kind.IsMisc() {
			var b []byte
			switch params.Kind {
			case KindConfig:
				b, err = json.Marshal(&deployment.ConfigData)
			case KindSecret:
				metaBucket := tx.Bucket([]byte(BUCKET_META))
				secretKey := metaBucket.Get([]byte(KEY_SECRET_KEY))
				accessKey := metaBucket.Get([]byte(KEY_ACCESS_KEY))
				kek, err := vault.DeriveKEK(string(secretKey), accessKey)
				if err != nil {
					return err
				}

				var secrets = make(map[string]vault.SecretRecord, len(deployment.SecretData))
				for sk, sv := range deployment.SecretData {
					dek, err := vault.GenerateDEK()
					if err != nil {
						return err
					}

					encryptedSv, _, err := vault.EncryptWithDEK(sv, dek)
					if err != nil {
						return err
					}
					encDEK, err := vault.EncryptDEK(dek, kek)
					if err != nil {
						return err
					}

					secrets[sk] = vault.SecretRecord{
						EncryptedSecret: encryptedSv,
						EncryptedDEK:    encDEK,
					}
				}

				b, err = json.Marshal(&secrets)
			}
			if err != nil {
				return err
			}
			resourceData = string(b)
		}

		ver, _ := deploymentBucket.Cursor().Last()
		if ver == nil {
			ver = bige.ByteFromUInt32(1)

			diff := differ.DiffMain("", resourceData, true)
			patchList := differ.PatchMake("", resourceData, diff)
			patch := differ.PatchToText(patchList)

			depBytes, err := json.Marshal(types.Deployment{
				Description: deployment.Description,
				Version:     "1",
				Deployed:    false,
				CreatedAt:   time.Now().Format(time.RFC3339),
				CreatedBy:   claims["email"].(string),
				Diff:        patch,
			})

			if err != nil {
				return err
			}
			return deploymentBucket.Put(ver, depBytes)
		} else {
			last := bige.UInt32FromByte(ver)
			next := last + 1
			version = next

			versionKey := fmt.Sprintf("%s_version", resourceKey)
			resourceJSON := buildResource(sc.dbconns[params.ProjectKey], resourceKey, versionKey, true, nil)

			diff := differ.DiffMain(resourceJSON, resourceData, true)
			patchList := differ.PatchMake(resourceJSON, resourceData, diff)
			patch := differ.PatchToText(patchList)

			depBytes, err := json.Marshal(types.Deployment{
				Description: deployment.Description,
				Version:     fmt.Sprint(next),
				Deployed:    false,
				CreatedAt:   time.Now().Format(time.RFC3339),
				CreatedBy:   claims["email"].(string),
				Diff:        patch,
			})

			if err != nil {
				return err
			}
			return deploymentBucket.Put(bige.ByteFromUInt32(next), depBytes)
		}
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
		Action:    "create_deployment",
		Timestamp: time.Now(),
		Details: map[string]any{
			"kind":    params.Kind,
			"version": version,
		},
	})

	// TODO :: maybe we give proper deployment creation response afterwards..
	enc.Encode(map[string]uint32{
		"version": version,
	})
}
