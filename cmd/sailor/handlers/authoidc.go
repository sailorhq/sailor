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
	"errors"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/oauth2"
	"k8s.io/apimachinery/pkg/util/json"
)

func (sc *SailorCore) AuthOIDCHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	fingerprint := ctx.QueryArgs().Peek("fp")
	if len(fingerprint) == 0 {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "unknown caller"})
		return
	}

	ss, err := getSailorSetting(sc.dbconns[BUCKET_ADMIN])
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	if ss.OIDC == nil {
		ctx.SetStatusCode(http.StatusNotFound)
		enc.Encode(ResponseMessage{Message: "oidc sailor settings not present"})
		return
	}

	provider, err := oidc.NewProvider(ctx, ss.OIDC.IssuerURL)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	oauth2Config := oauth2.Config{
		ClientID:     ss.OIDC.ClientID,
		ClientSecret: ss.OIDC.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  ss.OIDC.RedirectURL,
		Scopes:       ss.OIDC.Scopes,
	}

	// TODO :: think about the state which should be passed as part of OIDC request
	ctx.Redirect(oauth2Config.AuthCodeURL(string(fingerprint)), http.StatusFound)
}

func getSailorSetting(adminDB *bolt.DB) (*SailorSetting, error) {
	var ss SailorSetting
	err := adminDB.View(func(tx *bolt.Tx) error {
		settingBytes := tx.Bucket([]byte(BUCKET_SETTING)).Get([]byte(KEY_SETTING))
		if settingBytes == nil {
			return errors.New("sailor settings not found.")
		}

		if err := json.Unmarshal(settingBytes, &ss); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &ss, nil
}
