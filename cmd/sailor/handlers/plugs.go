package handlers

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
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
