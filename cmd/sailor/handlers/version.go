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
package handlers

import (
	"fmt"
	"net/http"
)

// VersionHandler ...
// TODO :: need to create version API for efficient pulls
func (sc *SailorCore) VersionHandler(w http.ResponseWriter, r *http.Request) {
	// key := fmt.Sprintf("%s-%s", params.Ns, params.App)
	// if version, ok := sc.versions[key]; ok {
	// 	fmt.Fprint(w, version)
	// 	return
	// }

	// db, err := sc.getDBConn(params)
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	var ver string
	// db.View(func(tx *bolt.Tx) error {
	// 	// fetch current deployed version ...
	// 	metaBucket := tx.Bucket([]byte(BUCKET_META))
	// 	ver = string(metaBucket.Get([]byte(KEY_DEPLOYED_VERSION)))

	// 	return nil
	// })

	fmt.Fprint(w, ver)
}
