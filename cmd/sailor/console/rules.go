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
	"io"
	"net/http"

	bolt "go.etcd.io/bbolt"
)

func (c *Console) RuleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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
