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
