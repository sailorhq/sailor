package handlers

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func (sc *SailorCore) GetProjects(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)
	projects, err := sc.SailorSail.GetProjects()
	if err != nil {
		sc.Log.Error("unable to get projects list from meta bucket", zap.Error(err))
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		enc.Encode(ResponseMessage{"error while getting sailor projects"})
		return
	}

	enc.Encode(projects)
}
