// sailor
// Copyright (C) 2025 SailorHQ and Ashish Shekar (codekidX)

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package console

import (
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sailorhq/sailor/internal/types"
	bolt "go.etcd.io/bbolt"
)

func (c *Console) AdminStateHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)

	params, err := c.extractSailorParams(r)
	if err != nil {
		// TODO: log here!
		enc.Encode(ResponseMessage{Message: err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	db, err := c.getDBConn(params)
	if err != nil {
		enc.Encode(ResponseMessage{Message: err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	configStr := "" //buildResource(db, "", KEY_DEPLOYED_VERSION)
	var builtConfig map[string]any
	json.Unmarshal([]byte(configStr), &builtConfig)
	w.Header().Set("Content-Type", "application/json")

	resp := types.AdminSailorState{
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
		resp.Rules = string(metaBucket.Get([]byte(KEY_RULES)))

		// fetch secrets...
		secretsBucket := tx.Bucket([]byte(BUCKET_SECRETS))
		cur := secretsBucket.Cursor()

		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			decoded, err := jwt.Parse(string(v), func(token *jwt.Token) (interface{}, error) {
				return []byte(resp.SecretKey), nil
			})
			if err != nil {
				return err
			}

			resp.Secrets = append(resp.Secrets, types.Secret{
				Name:  string(k),
				Value: decoded.Claims.(jwt.MapClaims)["data"].(string),
			})
		}

		deploymentsBucket := tx.Bucket([]byte(BUCKET_DEPLOYMENT))
		cur = deploymentsBucket.Cursor()
		for k, v := cur.Last(); k != nil; k, v = cur.Prev() {
			var deployment types.Deployment
			json.Unmarshal(v, &deployment)

			if resp.Meta.Version == deployment.Version {
				deployment.Deployed = true
			} else {
				deployment.Deployed = false
			}

			resp.Deployments = append(resp.Deployments, deployment)
		}

		return nil
	})

	enc.Encode(&resp)
}
