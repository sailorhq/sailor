package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/codekidx/sailor/internal/types"
	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) DeployHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	enc := json.NewEncoder(w)

	params, err := sc.extractSailorParams(r)
	if err != nil {
		// TODO: log here!
		enc.Encode(ResponseMessage{Message: err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	db, err := sc.getDBConn(params)
	if err != nil {
		enc.Encode(ResponseMessage{Message: err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = db.Update(func(tx *bolt.Tx) error {
		metaBucket := tx.Bucket([]byte(BUCKET_META))
		if err = metaBucket.Put([]byte(KEY_DEPLOYED_VERSION), []byte(params.DeployVersion)); err == nil {
			sc.versions[fmt.Sprintf("%s-%s", params.Ns, params.App)] = params.DeployVersion
		}

		deploymentBucket := tx.Bucket([]byte(BUCKET_DEPLOYMENT))
		depBytes := deploymentBucket.Get([]byte(params.DeployVersion))
		if err != nil {
			return err
		}

		var deployment types.Deployment
		err = json.Unmarshal(depBytes, &deployment)

		deployment.Deployed = true
		deployment.DeployedAt = time.Now().Format(time.RFC3339)
		deployment.DeployedBy = params.Username

		depBytes, err = json.Marshal(deployment)
		if err != nil {
			return err
		}

		return deploymentBucket.Put([]byte(params.DeployVersion), depBytes)
	})

	if err != nil {
		enc.Encode(err.Error())
		return
	}

	enc.Encode("done!")

}
