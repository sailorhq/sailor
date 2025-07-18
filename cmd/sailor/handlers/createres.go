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

type SchemaSetting struct {
	Strict bool `json:"strict"`
}

type ResourceSetting struct {
	Deploy DeploySetting `json:"deploy"`
	Schema SchemaSetting `json:"schema"`
}

type Resource struct {
	Data    map[string]any   `json:"data"`
	Setting *ResourceSetting `json:"setting"`
	Version string           `json:"version"`
}

type SailorResource struct {
	Data    map[string]any   `json:"data"`
	Schema  map[string]any   `json:"schema"`
	Setting *ResourceSetting `json:"setting"`
}

func (sc *SailorCore) CreateResourceHandler(ctx *fasthttp.RequestCtx) {
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

	var resource SailorResource
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
			Schema: SchemaSetting{
				Strict: false,
			},
		}
	}

	err = sc.dbconns[projectKey].Update(func(tx *bolt.Tx) error {
		var resourceKey = kind
		if kind == KindMisc {
			resourceKey = name
		}

		resourceBucket := tx.Bucket([]byte(BUCKET_RESOURCE))
		resBytes := resourceBucket.Get([]byte(resourceKey))
		if resBytes != nil {
			return fmt.Errorf("%s is already created", resourceKey)
		}

		deploymentBucket := tx.Bucket([]byte(BUCKET_DEPLOYMENT))
		if _, err := deploymentBucket.CreateBucket([]byte(resourceKey)); err != nil {
			return err
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
