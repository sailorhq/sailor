package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) UpdateResourceSetting(ctx *fasthttp.RequestCtx) {
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

	err = sc.dbconns[projectKey].Update(func(tx *bolt.Tx) error {
		var resourceKey = kind
		if kind == KindMisc {
			resourceKey = name
		}

		resourceBucket := tx.Bucket([]byte(BUCKET_RESOURCE))
		resBytes := resourceBucket.Get([]byte(resourceKey))
		if resBytes == nil {
			return fmt.Errorf("%s is not created", resourceKey)
		}

		var sailorRes SailorResource
		if err := json.Unmarshal(resBytes, &sailorRes); err != nil {
			return err
		}

		sailorRes.Setting = resource.Setting

		if resBytes, err = json.Marshal(&sailorRes); err != nil {
			return err
		}

		return resourceBucket.Put([]byte(resourceKey), resBytes)
	})

	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	enc.Encode(ResponseMessage{Message: "ok"})
}
