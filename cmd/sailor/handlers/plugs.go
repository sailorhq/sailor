package handlers

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
	plugrpc "github.com/sailorhq/plug/sdk/proto"
)

type PlugResponse struct {
	Plugs            []string `json:"plugs"`
	AvaliableSignals []Keyed  `json:"available_signals"`
}

type Keyed struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

func (sc *SailorCore) GetSailorPlugsHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	var response = PlugResponse{
		AvaliableSignals: []Keyed{
			{"Create Project Signal", "project_create"},
			{"Create Resource Signal", "resource_create"},
			{"Create Deployment Signal", "deploy_create"},
			{"Deploy Signal", "deploy"}},
		Plugs: []string{},
	}
	if len(sc.setting.Rxs) == 0 {
		enc.Encode(response)
		return
	}

	for _, rx := range sc.setting.Rxs {
		response.Plugs = append(response.Plugs, rx.Name)
	}

	enc.Encode(response)
}

func (sc *SailorCore) PostProjectCreateSignalHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	var req plugrpc.ProjectCreateRequest
	if err := json.Unmarshal(ctx.Request.Body(), &req); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "invalid request body"})
		return
	}

	if req.Ns == "" || req.App == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "missing required fields"})
		return
	}

	if err := sc.plugman.FireProjectCreate(&req); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	ctx.SetStatusCode(fasthttp.StatusNoContent)
}

func (sc *SailorCore) PostDeploySignalHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	var req plugrpc.DeployRequest
	if err := json.Unmarshal(ctx.Request.Body(), &req); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "invalid request body"})
		return
	}

	if req.Ns == "" || req.App == "" || req.Kind == "" || req.ResourceKey == "" || req.Version == 0 || len(req.Content) == 0 {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "missing required fields"})
		return
	}

	if err := sc.plugman.FireDeploy(&req); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	ctx.SetStatusCode(fasthttp.StatusNoContent)
}
