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
package backup

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/sailorhq/sailor/internal/types"
	diffmod "github.com/sergi/go-diff/diffmatchpatch"
	bolt "go.etcd.io/bbolt"
)

func BackupState(projectKey string, appDbConn *bolt.DB, bucket, region, accessKey, secretKey string) error {
	// TODO :: make sure all keys that we access from are obtained from a common place
	// this included the bucket names and key names
	var backupState = types.SailorState{
		Secrets: make(map[string][]byte),
	}
	err := appDbConn.View(func(tx *bolt.Tx) error {
		metaBucket := tx.Bucket([]byte("_meta"))
		secretsBucket := tx.Bucket([]byte("secrets"))

		secretsBucket.ForEach(func(k, v []byte) error {
			backupState.Secrets[string(k)] = v
			return nil
		})

		backupState.Version = string(metaBucket.Get([]byte("deploy_ver")))

		return nil
	})

	if err != nil {
		return err
	}

	builtConfig := buildConfig(appDbConn)
	backupState.Config = []byte(builtConfig)

	stateFilePath := fmt.Sprintf("state/%s.json", projectKey)
	b, err := json.Marshal(backupState)
	if err != nil {
		return err
	}
	return UploadToS3(b, bucket, region, accessKey, secretKey, stateFilePath)
}

// HACK :: buildconfig should be at a common place, first of all do we need build config
// at all, because this was used where we diff the changes with previous version and keep
// and then build them
func buildConfig(db *bolt.DB) string {
	var configJson string
	db.View(func(tx *bolt.Tx) error {
		deploymentBucket := tx.Bucket([]byte("deployments"))
		metaBucket := tx.Bucket([]byte("_meta"))

		min := []byte("0")
		max := metaBucket.Get([]byte("deploy_ver"))
		cur := deploymentBucket.Cursor()

		differ := diffmod.New()

		for depVer, depBytes := cur.Seek(min); depVer != nil && bytes.Compare(depVer, max) <= 0; depVer, depBytes = cur.Next() {
			var deployment types.Deployment
			json.Unmarshal(depBytes, &deployment)
			// fmt.Println("diff_ver", string(diff_ver), string(max))
			p, _ := differ.PatchFromText(string(deployment.Diff))
			configJson, _ = differ.PatchApply(p, configJson)
		}

		return nil
	})

	return configJson
}
