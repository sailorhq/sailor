package handlers

import (
	"fmt"
	"net/http"
	"os"

	bolt "go.etcd.io/bbolt"
)

func (sh *SailorCore) VersionHandler(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("ns")
	app := r.URL.Query().Get("app")
	_ = r.URL.Query().Get("key")

	if ns == "" || app == "" {
		fmt.Fprint(w, "namespace or app is empty")
		return
	}

	dbpath := fmt.Sprintf("./configs/%s-%s.db", ns, app)
	if f, _ := os.Stat(dbpath); f == nil {
		fmt.Fprint(w, "no such app in this namespace")
		return
	}

	db, err := bolt.Open(dbpath, 0600, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var ver string
	db.View(func(tx *bolt.Tx) error {
		// fetch current deployed version ...
		metaBucket := tx.Bucket([]byte(BUCKET_META))
		ver = string(metaBucket.Get([]byte(KEY_DEPLOYED_VERSION)))

		return nil
	})

	fmt.Fprint(w, ver)
}
