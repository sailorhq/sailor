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
