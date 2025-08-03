package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"

	v1 "github.com/codekidx/sailor/pkg/core/v1"
	"github.com/valyala/fasthttp"
	"go.etcd.io/bbolt"
)

func (sc *SailorCore) GetDeploymentHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	params := extractSailorParams(ctx)

	if params.Ns == "" || params.App == "" || params.Kind == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "namespace or app should not be empty"})
		return
	}

	if params.Kind != KindConfig && params.Kind != KindSecret && params.Kind != KindMisc {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "unknown resource kind"})
		return
	}

	if params.Kind == KindMisc && params.ResourceName == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "missing resource name"})
		return
	}

	var version string
	if params.RequestedVersion == "" {
		version = "all"
	} else if params.RequestedVersion == "latest" {
		version = "latest"
	} else {
		// try parsing into a version number
		if _, err := strconv.Atoi(params.RequestedVersion); err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			enc.Encode(ResponseMessage{Message: "requested version is neither latest nor a number"})
			return
		}
	}

	var depResp = v1.DeploymentResponse{
		Deployments: []v1.Deployment{},
	}
	err := sc.dbconns[params.ProjectKey].View(func(tx *bbolt.Tx) error {
		resourceKey := params.Kind.ResourceKey(params.ResourceName)
		deploymentBucket := tx.Bucket([]byte(BUCKET_DEPLOYMENT)).Bucket([]byte(resourceKey))
		metaBucket := tx.Bucket([]byte(BUCKET_META))

		if version == "all" {
			cur := deploymentBucket.Cursor()
			for k, depBytes := cur.Last(); k != nil; k, depBytes = cur.Prev() {
				var dep v1.Deployment
				if err := json.Unmarshal(depBytes, &dep); err != nil {
					return err
				}

				depResp.Deployments = append(depResp.Deployments, dep)
			}
		} else if version == "latest" {
			versionKey := fmt.Sprintf("%s_version", resourceKey)
			latest := metaBucket.Get([]byte(versionKey))
			depBytes := deploymentBucket.Get(latest)
			if depBytes == nil {
				return fmt.Errorf("%s not found", version)
			}

			var dep v1.Deployment
			if err := json.Unmarshal(depBytes, &dep); err != nil {
				return err
			}
			depResp.Deployments = append(depResp.Deployments, dep)
		} else {
			depBytes := deploymentBucket.Get([]byte(version))
			if depBytes == nil {
				return fmt.Errorf("%s not found", version)
			}

			var dep v1.Deployment
			if err := json.Unmarshal(depBytes, &dep); err != nil {
				return err
			}
			depResp.Deployments = append(depResp.Deployments, dep)
		}
		return nil
	})

	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	enc.Encode(depResp)
}
