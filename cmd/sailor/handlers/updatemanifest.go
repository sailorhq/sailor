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
	"time"

	"github.com/golang-jwt/jwt/v5"
	v1 "github.com/sailorhq/sailor/pkg/core/v1"
	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) UpdateManifestHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	var manifest v1.SailorManifest
	if err := json.Unmarshal(ctx.Request.Body(), &manifest); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{err.Error()})
		return
	}

	err := sc.dbconns[BUCKET_ADMIN].Update(func(tx *bolt.Tx) error {
		settingBytes := tx.Bucket([]byte(BUCKET_SETTING)).Get([]byte(KEY_SETTING))
		var setting v1.SailorSetting
		if err := json.Unmarshal(settingBytes, &setting); err != nil {
			return err
		}

		setting.Manifest = manifest
		b, err := json.Marshal(&setting)
		if err != nil {
			return err
		}
		return tx.Bucket([]byte(BUCKET_SETTING)).Put([]byte(KEY_SETTING), b)
	})

	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		enc.Encode(ResponseMessage{err.Error()})
		return
	}

	// set the updated manifest inside our in-memory lookup
	// FIXME :: i think it does not matter to fetch settings from boltdb everytime because it's
	// fast and we don't go on updating manifests every second, so just remove settings and manifest
	// from the SailorCore struct!!
	sc.setting.Manifest = manifest

	claims := ctx.UserValue("__sailor_claims").(jwt.MapClaims)
	go sc.addAuditEvent(&v1.AuditEvent{
		Username:  claims["email"].(string),
		Action:    "update_manifest",
		Timestamp: time.Now(),
		Details:   manifest,
	})

	enc.Encode(ResponseMessage{"ok"})
}
