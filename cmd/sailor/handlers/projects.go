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

func (sc *SailorCore) GetResourceList(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)
	var (
		projectKey string
	)
	if key, ok := ctx.UserValue("projectKey").(string); ok {
		projectKey = key
	}
	resources, err := sc.SailorSail.GetResourceKeys(projectKey)
	if err != nil {
		sc.Log.Error("unable to get resource list from meta bucket", zap.Error(err))
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		enc.Encode(ResponseMessage{"error while getting sailor resources"})
		return
	}

	enc.Encode(resources)
}
