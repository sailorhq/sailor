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

type SailorResource struct {
	Schema  map[string]any   `json:"schema"`
	Setting *ResourceSetting `json:"setting"`
}

func (sc *SailorCore) CreateResourceHandler(ctx *fasthttp.RequestCtx) {
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

	err = sc.dbconns[params.ProjectKey].Update(func(tx *bolt.Tx) error {
		resourceKey := params.Kind.ResourceKey(params.ResourceName)

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
