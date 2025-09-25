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

	v1 "github.com/sailorhq/sailor/pkg/core/v1"
	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) GetResourceSchemaHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)
	enc.SetEscapeHTML(false)

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

	var schema map[string]any
	err := sc.dbconns[params.ProjectKey].View(func(tx *bolt.Tx) error {
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

		schema = sailorRes.Schema

		return nil
	})

	if len(schema) == 0 {
		ctx.SetStatusCode(http.StatusNotFound)
		return
	}

	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	ctx.Response.Header.Set("Content-Type", "application/json")
	enc.Encode(schema)
}
