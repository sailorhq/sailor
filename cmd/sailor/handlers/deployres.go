package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/codekidx/sailor/internal/types"
	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
)

type DeployResourceRequest struct {
	Version string `json:"version"`
}

func (sc *SailorCore) DeployResourceHandler(ctx *fasthttp.RequestCtx) {
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

	var deployment DeployResourceRequest
	err := json.Unmarshal(ctx.Request.Body(), &deployment)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	if deployment.Version == "" {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "version is required"})
		return
	}

	projectKey := fmt.Sprintf("%s-%s", ns, app)
	err = sc.dbconns[projectKey].Update(func(tx *bolt.Tx) error {
		resourceBucket := tx.Bucket([]byte(BUCKET_RESOURCE))
		deploymentBucket := tx.Bucket([]byte(BUCKET_DEPLOYMENT))
		depBytes := deploymentBucket.Get([]byte(deployment.Version))
		if depBytes == nil {
			return fmt.Errorf("no deployment with version: %s", deployment.Version)
		}

		var resourceKey = kind
		if kind == KindMisc {
			resourceKey = name
		}

		versionKey := fmt.Sprintf("%s_version", resourceKey)
		metaBucket := tx.Bucket([]byte(BUCKET_META))
		metaBucket.Put([]byte(versionKey), []byte(deployment.Version))

		var deploymentInfo types.Deployment
		if err := json.Unmarshal(depBytes, &deploymentInfo); err != nil {
			return err
		}

		var resource CreateResourceRequest
		resBytes := resourceBucket.Get([]byte(resourceKey))
		if err := json.Unmarshal(resBytes, &resource); err != nil {
			return err
		}

		var data map[string]any
		if err := json.Unmarshal(deploymentInfo.Data, &data); err != nil {
			return err
		}
		resource.Data = data

		if resBytes, err = json.Marshal(&resource); err != nil {
			return err
		}

		if err := resourceBucket.Put([]byte(resourceKey), resBytes); err != nil {
			return err
		}

		// TODO :: here we will get the resource setting and do k8s deployment
		// if k8s is marked true .. if k8s deployment fails then we will
		// rollback all the operations under tx and return error to the user
		return nil
	})

	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	enc.Encode(ResponseMessage{Message: "ok"})
}
