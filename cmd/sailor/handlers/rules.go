package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	bolt "go.etcd.io/bbolt"
)

func (sc *SailorCore) RuleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	if ok := json.Valid(b); !ok {
		enc.Encode(ResponseMessage{Message: "invalid json for rules!"})
		return
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		metaBucket := tx.Bucket([]byte(BUCKET_META))
		return metaBucket.Put([]byte(KEY_RULES), b)
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

}
