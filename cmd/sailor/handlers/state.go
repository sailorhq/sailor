package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/codekidx/sailor/internal/types"
	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) StateHandler(w http.ResponseWriter, r *http.Request) {
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

	configStr := buildConfig(db)
	var builtConfig map[string]any
	json.Unmarshal([]byte(configStr), &builtConfig)
	w.Header().Set("Content-Type", "application/json")

	resp := types.SailorState{
		Configs:     builtConfig,
		Secrets:     []types.Secret{},
		Deployments: []types.Deployment{},
	}

	db.View(func(tx *bolt.Tx) error {
		// fetch current deployed version ...
		metaBucket := tx.Bucket([]byte(BUCKET_META))
		resp.Meta.Version = string(metaBucket.Get([]byte(KEY_DEPLOYED_VERSION)))
		resp.AccessKey = string(metaBucket.Get([]byte(KEY_ACCESS_KEY)))
		resp.SecretKey = string(metaBucket.Get([]byte(KEY_SECRET_KEY)))

		// fetch secrets...
		secretsBucket := tx.Bucket([]byte(BUCKET_SECRETS))
		cur := secretsBucket.Cursor()

		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			resp.Secrets = append(resp.Secrets, types.Secret{
				Name:  string(k),
				Value: string(v),
			})
		}

		deploymentsBucket := tx.Bucket([]byte(BUCKET_DEPLOYMENT))
		cur = deploymentsBucket.Cursor()
		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			var deployment types.Deployment
			json.Unmarshal(v, &deployment)

			if resp.Meta.Version == deployment.Version {
				deployment.Deployed = true
			}

			resp.Deployments = append(resp.Deployments, deployment)
		}

		return nil
	})

	enc.Encode(&resp)
}
