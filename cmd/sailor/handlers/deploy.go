package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/codekidx/sailor/cmd/sailor/backup"
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

	projectKey := fmt.Sprintf("%s-%s", params.Ns, params.App)

	err = db.Update(func(tx *bolt.Tx) error {
		metaBucket := tx.Bucket([]byte(BUCKET_META))
		if err = metaBucket.Put([]byte(KEY_DEPLOYED_VERSION), []byte(params.DeployVersion)); err == nil {
			sc.versions[projectKey] = params.DeployVersion
		}

		deploymentBucket := tx.Bucket([]byte(BUCKET_DEPLOYMENT))
		depBytes := deploymentBucket.Get([]byte(params.DeployVersion))
		if err != nil {
			tx.Rollback()
			return err
		}

		var deployment types.Deployment
		err = json.Unmarshal(depBytes, &deployment)

		deployment.Deployed = true
		deployment.DeployedAt = time.Now().Format(time.RFC3339)
		deployment.DeployedBy = params.Username

		depBytes, err = json.Marshal(deployment)
		if err != nil {
			tx.Rollback()
			return err
		}

		if err = deploymentBucket.Put([]byte(params.DeployVersion), depBytes); err != nil {
			tx.Rollback()
			return err
		}

		// start backup
		err = backupState(projectKey, db, sc.dbconns[BUCKET_ADMIN])
		if err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	enc.Encode("done!")

}

func backupState(projectKey string, appDbConn *bolt.DB, adminDbConn *bolt.DB) error {
	var backupDetails BackupDetails
	return adminDbConn.View(func(tx *bolt.Tx) error {
		backupBucket := tx.Bucket([]byte(BUCKET_BACKUP))
		backupBytes := backupBucket.Get([]byte(KEY_S3))
		if backupBytes != nil {
			err := json.Unmarshal(backupBytes, &backupDetails)
			if err != nil {
				return err
			}
		} else {
			// TODO :: log here that consumer cannot recover if sailor is not alive!
			return nil
		}

		// TODO :: BackupDetails struct should be in a common place
		if err := backup.BackupState(projectKey, appDbConn, backupDetails.Bucket, backupDetails.Region, backupDetails.AccessKey, backupDetails.SecretKey); err != nil {
			return err
		}

		return nil
	})
}
