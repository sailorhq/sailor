package handlers

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

func (sc *SailorCore) GetManifestHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)
	ss, err := getSailorSetting(sc.dbconns[BUCKET_ADMIN])
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		enc.Encode(ResponseMessage{"sailor settings not found"})
		return
	}

	enc.Encode(ss.Manifest)
}
