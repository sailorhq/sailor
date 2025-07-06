package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/codekidx/sailor/cmd/sailor/backup"
	bolt "go.etcd.io/bbolt"
)

type BackupDetails struct {
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

func (sc *SailorCore) BackupHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	var backupDetails BackupDetails
	if err := json.NewDecoder(r.Body).Decode(&backupDetails); err != nil {
		enc.Encode(ResponseMessage{Message: "Invalid request body"})
		return
	}

	err := sc.dbconns[BUCKET_ADMIN].Update(func(tx *bolt.Tx) error {
		backupBucket, err := tx.CreateBucketIfNotExists([]byte(BUCKET_BACKUP))
		if err != nil {
			return err
		}

		backupBytes, err := json.Marshal(backupDetails)
		if err != nil {
			return err
		}
		return backupBucket.Put([]byte(KEY_S3), backupBytes)
	})
	if err != nil {
		enc.Encode(ResponseMessage{Message: "Failed to update backup details"})
		return
	}

	err = backup.BackupRawSails(
		backupDetails.Bucket, backupDetails.Region, backupDetails.AccessKey, backupDetails.SecretKey)
	if err != nil {
		enc.Encode(ResponseMessage{Message: "Failed to backup sails"})
		return
	}
}
