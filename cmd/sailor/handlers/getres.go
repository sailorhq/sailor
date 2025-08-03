package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func (sc *SailorCore) GetResourceHandler(ctx *fasthttp.RequestCtx) {
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

	resourceKey := params.Kind.ResourceKey(params.ResourceName)
	versionKey := fmt.Sprintf("%s_version", resourceKey)

	resStr := buildResource(sc.dbconns[params.ProjectKey], resourceKey, versionKey, false)
	if resStr == "" {
		ctx.SetStatusCode(http.StatusNotFound)
		enc.Encode(ResponseMessage{"this resource was never deployed!"})
		return
	}

	if params.Kind.IsMisc() {
		enc.Encode(resStr)
		return
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(resStr), &data); err != nil {
		sc.Log.Error("resource get has failed", zap.Error(err), zap.String("built_res", resStr))
		ctx.SetStatusCode(http.StatusNotFound)
		enc.Encode(ResponseMessage{Message: "error while parsing resource json!"})
		return
	}

	ctx.Response.Header.Set("Content-Type", "application/json")
	enc.Encode(data)
}
