package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/codekidx/sailor/internal/types"
	"github.com/valyala/fasthttp"
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

	projectKey := fmt.Sprintf("%s-%s", ns, app)
	sc.dbconns[projectKey].Update(func(tx *bolt.Tx) error {
		deploymentBucket := tx.Bucket([]byte(BUCKET_DEPLOYMENT))
		ver, _ := deploymentBucket.Cursor().Last()
		if ver == nil {
			ver = []byte("1")

			b, err := json.Marshal(&deployment.Data)
			if err != nil {
				return err
			}

			next, _ := strconv.Atoi(string(ver))
			next += 1

			depBytes, err := json.Marshal(types.Deployment{
				Description: deployment.Description,
				Version:     strconv.Itoa(next),
				Deployed:    false,
				CreatedAt:   time.Now().Format(time.RFC3339),
				CreatedBy:   "--todo--",
				Data:        b,
			})

			if err != nil {
				return err
			}
			return deploymentBucket.Put(fmt.Append(nil, next), depBytes)
		}

		return nil
	})

}
