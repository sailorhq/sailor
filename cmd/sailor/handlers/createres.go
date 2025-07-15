package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
)

type DeploySetting struct {
	K8s bool `json:"k8s"`
}

type ResourceSetting struct {
	Deploy DeploySetting
}

type Resource struct {
	Data    map[string]any   `json:"data"`
	Setting *ResourceSetting `json:"setting"`
	Version string           `json:"version"`
}

type CreateResourceRequest struct {
	Data    map[string]any   `json:"data"`
	Setting *ResourceSetting `json:"setting"`
}

func (sc *SailorCore) CreateResourceHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	ns := ctx.UserValue("namespace").(string)
	app := ctx.UserValue("app").(string)
	kind := ctx.UserValue("kind").(string)
	name := ctx.UserValue("name").(string)

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

	var resource CreateResourceRequest
	err := json.Unmarshal(ctx.Request.Body(), &resource)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	// if setting is not present add default setting
	if resource.Setting == nil {
		resource.Setting = &ResourceSetting{
			Deploy: DeploySetting{K8s: false},
		}
	}

	projectKey := fmt.Sprintf("%s-%s", ns, app)
	if _, ok := sc.dbconns[projectKey]; !ok {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "no such project in sailor"})
		return
	}

	err = sc.dbconns[projectKey].Update(func(tx *bolt.Tx) error {
		resourceBucket := tx.Bucket([]byte(BUCKET_RESOURCE))

		var resourceKey = kind
		if kind == KindMisc {
			resourceKey = name
		}

		res, err := json.Marshal(&resource)
		if err != nil {
			return err
		}

		return resourceBucket.Put([]byte(resourceKey), res)
	})

	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	enc.Encode(ResponseMessage{Message: "ok"})

}
