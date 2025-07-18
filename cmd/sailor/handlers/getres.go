package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/valyala/fasthttp"
)

func (sc *SailorCore) GetResourceHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	ns := ctx.UserValue("namespace").(string)
	app := ctx.UserValue("app").(string)
	kind := ctx.UserValue("kind").(string)
	var name string
	if n, ok := ctx.UserValue("name").(string); ok {
		name = n
	}

	if ns == "" || app == "" || kind == "" {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "namespace or app should not be empty"})
		return
	}

	if kind != KindConfig && kind != KindSecret && kind != KindMisc {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "unknown resource kind"})
		return
	}

	if kind == KindMisc && name == "" {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "missing resource name"})
		return
	}

	projectKey := fmt.Sprintf("%s-%s", ns, app)
	if _, ok := sc.dbconns[projectKey]; !ok {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "no such project in sailor"})
		return
	}

	var resourceKey = kind
	if kind == KindMisc {
		resourceKey = name
	}
	versionKey := fmt.Sprintf("%s_version", resourceKey)

	resStr := buildResource(sc.dbconns[projectKey], resourceKey, versionKey)
	if resStr == "" {
		ctx.SetStatusCode(http.StatusNotFound)
		return
	}

	if kind == KindMisc {
		enc.Encode(resStr)
		return
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(resStr), &data); err != nil {
		ctx.SetStatusCode(http.StatusNotFound)
		enc.Encode(ResponseMessage{Message: "error while parsing resource json!"})
		return
	}

	ctx.Response.Header.Set("Content-Type", "application/json")
	enc.Encode(data)
}
