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

	v1 "github.com/codekidx/sailor/pkg/core/v1"
	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) UpdateResourceSetting(ctx *fasthttp.RequestCtx) {
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

	err = sc.dbconns[params.ProjectKey].Update(func(tx *bolt.Tx) error {
		resourceKey := params.Kind.ResourceKey(params.ResourceName)

		resourceBucket := tx.Bucket([]byte(BUCKET_RESOURCE))
		resBytes := resourceBucket.Get([]byte(resourceKey))
		if resBytes == nil {
			return fmt.Errorf("%s is not created", resourceKey)
		}

		var sailorRes v1.SailorResource
		if err := json.Unmarshal(resBytes, &sailorRes); err != nil {
			return err
		}

		sailorRes.Setting = resource.Setting

		if resBytes, err = json.Marshal(&sailorRes); err != nil {
			return err
		}

		return resourceBucket.Put([]byte(resourceKey), resBytes)
	})

	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	enc.Encode(ResponseMessage{Message: "ok"})
}
