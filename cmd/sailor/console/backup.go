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

	"github.com/sailorhq/sailor/cmd/sailor/backup"
	bolt "go.etcd.io/bbolt"
)

type BackupDetails struct {
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

func (c *Console) BackupHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	var backupDetails BackupDetails
	if err := json.NewDecoder(r.Body).Decode(&backupDetails); err != nil {
		enc.Encode(ResponseMessage{Message: "Invalid request body"})
		return
	}

	err := c.dbconns[BUCKET_ADMIN].Update(func(tx *bolt.Tx) error {
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
