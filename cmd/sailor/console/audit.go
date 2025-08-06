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
	"time"

	bolt "go.etcd.io/bbolt"
)

type AuditEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Username  string    `json:"username"`
	Namespace string    `json:"namespace"`
	App       string    `json:"app"`
	Action    string    `json:"action"`
	Details   any       `json:"details"`
}

func (c *Console) AuditHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		c.getAuditEvents(w)
	case http.MethodPost:
		c.addAuditEvent(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (c *Console) getAuditEvents(w http.ResponseWriter) {
	db := c.dbconns[BUCKET_AUDIT]

	var events = []AuditEvent{}
	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(BUCKET_AUDIT_TRAIL)).Cursor()

		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			var ae AuditEvent
			json.Unmarshal(v, &ae)
			events = append(events, ae)
		}
		return nil
	})

	enc := json.NewEncoder(w)
	enc.Encode(events)
}

func (c *Console) addAuditEvent(w http.ResponseWriter, r *http.Request) {
	db := c.dbconns[BUCKET_AUDIT]

	var ae AuditEvent
	json.NewDecoder(r.Body).Decode(&ae)

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_AUDIT_TRAIL))
		json, err := json.Marshal(ae)
		if err != nil {
			return err
		}
		return b.Put([]byte(ae.Timestamp.Format(time.RFC3339)), json)
	})

	enc := json.NewEncoder(w)
	enc.Encode(ResponseMessage{Message: "Audit event added"})
	w.WriteHeader(http.StatusOK)
}
