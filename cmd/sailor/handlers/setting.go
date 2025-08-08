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
	"net/http"
	"slices"

	v1 "github.com/sailorhq/sailor/pkg/core/v1"
	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) SailorSettingHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	var ss v1.SailorSetting
	if err := json.Unmarshal(ctx.Request.Body(), &ss); err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	if ss.OIDC != nil {
		if ss.OIDC.ClientID == "" || ss.OIDC.ClientSecret == "" ||
			ss.OIDC.IssuerURL == "" || ss.OIDC.RedirectURL == "" || len(ss.OIDC.Scopes) == 0 {
			ctx.SetStatusCode(http.StatusBadRequest)
			enc.Encode(ResponseMessage{Message: "oidc validations failed, some required fields are empty"})
			return
		}

		if !slices.Contains(ss.OIDC.Scopes, "email") {
			ctx.SetStatusCode(http.StatusBadRequest)
			enc.Encode(ResponseMessage{Message: "oidc:scopes must contain email"})
			return
		}
	}

	err := sc.dbconns[BUCKET_ADMIN].Update(func(tx *bolt.Tx) error {
		var currSetting v1.SailorSetting
		currSettingBytes := tx.Bucket([]byte(BUCKET_SETTING)).Get([]byte(KEY_SETTING))
		if err := json.Unmarshal(currSettingBytes, &currSetting); err != nil {
			return nil
		}

		currSetting.OIDC = ss.OIDC
		currSetting.Webhook = ss.Webhook
		b, err := json.Marshal(&currSetting)
		if err != nil {
			return err
		}
		return tx.Bucket([]byte(BUCKET_SETTING)).Put([]byte(KEY_SETTING), b)
	})

	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	// we update the sailor settings inside our own core memory lookup
	sc.setting = &ss

	enc.Encode(ResponseMessage{Message: "ok"})
}

func (sc *SailorCore) GetSailorSettingHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)
	ctx.Response.Header.Set("Content-Type", "application/json")
	enc.Encode(sc.setting)
}
