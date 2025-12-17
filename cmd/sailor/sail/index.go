package sail

import (
	"encoding/json"
	"fmt"

	"github.com/sailorhq/sailor/internal/constants"
	"github.com/sailorhq/sailor/internal/types"
	bolt "go.etcd.io/bbolt"
)

type Sail interface {
	CreateProject(ns, app string) error
}

type CoreSail struct {
	Meta  *bolt.DB
	Admin *bolt.DB
	Audit *bolt.DB

	ProjectMap map[string]*bolt.DB
}

func (cs *CoreSail) CreateProject(ns, app string) error {
	return cs.Meta.Update(func(tx *bolt.Tx) error {
		projectBucket, err := tx.CreateBucketIfNotExists([]byte(constants.BUCKET_PROJECTS))
		if err != nil {
			return err
		}
		projectBytes, err := json.Marshal(types.Project{Ns: ns, App: app})
		if err != nil {
			return err
		}
		return projectBucket.Put([]byte(cs.createProjectKey(ns, app)), projectBytes)
	})
}

func (cs *CoreSail) createProjectKey(ns, app string) string {
	return fmt.Sprintf("%s_%s", ns, app)
}
