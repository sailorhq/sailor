package handlers

import (
	"encoding/json"

	v1 "github.com/codekidx/sailor/pkg/core/v1"
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
		var setting SailorSetting
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

	enc.Encode(ResponseMessage{"ok"})
}
