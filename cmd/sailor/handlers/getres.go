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

	"github.com/sailorhq/sailor/pkg/vault"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func (sc *SailorCore) GetResourceHandler(ctx *fasthttp.RequestCtx) {
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

	resourceKey := params.Kind.ResourceKey(params.ResourceName)
	versionKey := fmt.Sprintf("%s_version", resourceKey)

	builtRes := buildResource(sc.dbconns[params.ProjectKey], resourceKey, versionKey, false, nil)
	if builtRes.Content == "" {
		ctx.SetStatusCode(http.StatusNotFound)
		enc.Encode(ResponseMessage{"this resource was never deployed!"})
		return
	}

	// TODO :: move header key to a common place
	ctx.Response.Header.Set("x-resource-version", fmt.Sprintf("%d", builtRes.Version))

	switch params.Kind {
	case KindMisc:
		enc.Encode(builtRes)
		return
	case KindConfig:
		var data map[string]any
		if err := json.Unmarshal([]byte(builtRes.Content), &data); err != nil {
			sc.Log.Error("resource get has failed", zap.Error(err), zap.String("built_res", builtRes.Content))
			ctx.SetStatusCode(http.StatusNotFound)
			enc.Encode(ResponseMessage{Message: "error while parsing resource json!"})
			return
		}
		ctx.Response.Header.Set("Content-Type", "application/json")
		enc.Encode(data)
		return
	case KindSecret:
		if params.AccessKey == "" || params.SecretKey == "" {
			ctx.SetStatusCode(http.StatusBadRequest)
			enc.Encode(ResponseMessage{"key pair not provided"})
			return
		}

		var encSecrets map[string]vault.SecretRecord
		if err := json.Unmarshal([]byte(builtRes.Content), &encSecrets); err != nil {
			ctx.SetStatusCode(http.StatusInternalServerError)
			enc.Encode(ResponseMessage{err.Error()})
			return
		}

		enc.Encode(encSecrets)
	}
}
