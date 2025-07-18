package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/codekidx/sailor/internal/types"
	"github.com/valyala/fasthttp"

	diffmod "github.com/sergi/go-diff/diffmatchpatch"
)

type CreateDeploymentRequest struct {
	Description string         `json:"desc"`
	Data        map[string]any `json:"data"`
}

func (sc *SailorCore) CreateDeploymentHandler(ctx *fasthttp.RequestCtx) {
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

	var deployment CreateDeploymentRequest
	err := json.Unmarshal(ctx.Request.Body(), &deployment)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	if len(deployment.Data) == 0 {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "cannot create deployment with empty data"})
		return
	}

	if deployment.Description == "" {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "description is required to create a deployment"})
		return
	}

	projectKey := fmt.Sprintf("%s-%s", ns, app)
	var resourceKey = kind
	if kind == KindMisc {
		resourceKey = name
	}
	differ := diffmod.New()

	err = sc.dbconns[projectKey].Update(func(tx *bolt.Tx) error {
		deploymentBucket := tx.Bucket([]byte(BUCKET_DEPLOYMENT)).Bucket([]byte(resourceKey))
		if deploymentBucket == nil {
			return errors.New("deployment not available, was the resource created?")
		}
		ver, _ := deploymentBucket.Cursor().Last()
		if ver == nil {
			ver = []byte("1")

			b, err := json.Marshal(&deployment.Data)
			if err != nil {
				return err
			}

			diff := differ.DiffMain("", string(b), true)
			patchList := differ.PatchMake("", string(b), diff)
			patch := differ.PatchToText(patchList)

			depBytes, err := json.Marshal(types.Deployment{
				Description: deployment.Description,
				Version:     "1",
				Deployed:    false,
				CreatedAt:   time.Now().Format(time.RFC3339),
				CreatedBy:   "--todo--",
				Diff:        patch,
			})

			if err != nil {
				return err
			}
			return deploymentBucket.Put(ver, depBytes)
		} else {
			next, _ := strconv.Atoi(string(ver))
			next += 1

			b, err := json.Marshal(&deployment.Data)
			if err != nil {
				return err
			}

			versionKey := fmt.Sprintf("%s_version", resourceKey)
			resourceJSON := buildResource(sc.dbconns[projectKey], resourceKey, versionKey)

			diff := differ.DiffMain(resourceJSON, string(b), true)
			patchList := differ.PatchMake(resourceJSON, string(b), diff)
			patch := differ.PatchToText(patchList)

			depBytes, err := json.Marshal(types.Deployment{
				Description: deployment.Description,
				Version:     fmt.Sprint(next),
				Deployed:    false,
				CreatedAt:   time.Now().Format(time.RFC3339),
				CreatedBy:   "--todo--",
				Diff:        patch,
			})

			if err != nil {
				return err
			}
			return deploymentBucket.Put(fmt.Append(nil, next), depBytes)
		}
	})

	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	enc.Encode(ResponseMessage{Message: "ok"})

}
