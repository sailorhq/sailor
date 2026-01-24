package sail

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sailorhq/sailor/internal/bige"
	"github.com/sailorhq/sailor/internal/constants"
	"github.com/sailorhq/sailor/internal/types"
	diffmod "github.com/sergi/go-diff/diffmatchpatch"
	bolt "go.etcd.io/bbolt"
)

type Sail interface {
	CreateProject(ns, app string) error
	GetProjects() ([]types.Project, error)
	GetCurrentDeployedVersion(projectKey, kind string) uint32
	GetResourceKeys(projectKey string) ([]string, error)
	GetPinnedVersion(projectKey, kind, releaseVer string) (uint32, error)
	BuildResource(
		projectKey, resourceKey, versionKey string,
		onTopOfLastDeployment bool,
		overrideMaxVersion []byte,
	) (string, uint32)
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
		return projectBucket.Put([]byte(cs.Core_CreateProjectKey(ns, app)), projectBytes)
	})
}

func (cs *CoreSail) GetProjects() (projs []types.Project, err error) {
	err = cs.Meta.View(func(tx *bolt.Tx) error {
		projectBucket := tx.Bucket([]byte(constants.BUCKET_PROJECTS))
		if projectBucket == nil {
			return errors.New("cannot find project bucket inside core _meta")
		}

		pc := projectBucket.Cursor()
		for k, projectBytes := pc.First(); k != nil; _, projectBytes = pc.Next() {
			var p types.Project
			if unmerr := json.Unmarshal(projectBytes, &p); unmerr != nil {
				continue
			}
			projs = append(projs, p)
		}

		return nil
	})
	if err != nil {
		return
	}

	return
}

// GetResourceKeys returns the list of resource keys inside a project using a project key.
func (cs *CoreSail) GetResourceKeys(projectKey string) ([]string, error) {
	var (
		projectDB *bolt.DB
		exists    bool
	)
	if projectDB, exists = cs.ProjectMap[projectKey]; !exists {
		return nil, errors.New("no such project")
	}

	var resourceKeys = []string{}
	err := projectDB.View(func(tx *bolt.Tx) error {
		resourceBucket := tx.Bucket([]byte(constants.BUCKET_RESOURCE))
		rbc := resourceBucket.Cursor()
		for k, _ := rbc.First(); k != nil; k, _ = rbc.Next() {
			resourceKeys = append(resourceKeys, string(k))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return resourceKeys, nil
}

func (cs *CoreSail) GetPinnedVersion(projectKey, kind, releaseVer string) (uint32, error) {
	var (
		projectDB *bolt.DB
		exists    bool
		ver       uint32
	)

	if projectDB, exists = cs.ProjectMap[projectKey]; !exists {
		return 0, errors.New("no such project")
	}

	err := projectDB.View(func(tx *bolt.Tx) error {
		releaseBucket := tx.Bucket([]byte(constants.BUCKET_RELEASE))
		if releaseBucket == nil {
			// there is no pinned version yet!
			return nil
		}

		kindBucket := releaseBucket.Bucket([]byte(kind))
		if kindBucket == nil {
			// there is no pinned version yet!
			return nil
		}

		pinnedVer := kindBucket.Get([]byte(releaseVer))
		if len(pinnedVer) == 0 {
			// there is no pinned version yet!
			return nil
		}

		ver = bige.UInt32FromByte(pinnedVer)

		return nil
	})
	if err != nil {
		return 0, err
	}

	return ver, nil
}

func (cs *CoreSail) BuildResource(
	projectKey, resourceKey, versionKey string,
	onTopOfLastDeployment bool,
	overrideMaxVersion []byte,
) (string, uint32) {
	var (
		configJson string
		max        []byte
		projectDB  *bolt.DB
		exists     bool
	)

	if projectDB, exists = cs.ProjectMap[projectKey]; !exists {
		return "", 0
	}

	projectDB.View(func(tx *bolt.Tx) error {
		deploymentBucket := tx.Bucket([]byte(constants.BUCKET_DEPLOYMENT)).Bucket([]byte(resourceKey))
		metaBucket := tx.Bucket([]byte(constants.BUCKET_META))

		min := bige.ByteFromUInt32(0)

		if overrideMaxVersion != nil {
			max = overrideMaxVersion
		} else {
			if onTopOfLastDeployment {
				max, _ = deploymentBucket.Cursor().Last()
			} else {
				max = metaBucket.Get([]byte(versionKey)) // also returns big-endian bytes back
			}
		}

		// if the key is still nil from deployment bucket, the resource was neither
		// deployed nor had any deployments created
		if max == nil {
			return nil
		}

		cur := deploymentBucket.Cursor()

		differ := diffmod.New()

		for depVer, depBytes := cur.Seek(min); depVer != nil && bytes.Compare(depVer, max) <= 0; depVer, depBytes = cur.Next() {
			var deployment types.Deployment
			json.Unmarshal(depBytes, &deployment)
			// fmt.Println("diff_ver", string(diff_ver), string(max))
			p, _ := differ.PatchFromText(string(deployment.Diff))
			configJson, _ = differ.PatchApply(p, configJson)
		}

		return nil
	})

	return configJson, bige.UInt32FromByte(max)
}

func (cs *CoreSail) GetCurrentDeployedVersion(projectKey, kind string) uint32 {
	var (
		projectDB *bolt.DB
		exists    bool
		ver       uint32
	)

	if projectDB, exists = cs.ProjectMap[projectKey]; !exists {
		return 0
	}

	projectDB.View(func(tx *bolt.Tx) error {
		metaBucket := tx.Bucket([]byte(constants.BUCKET_META))
		deployedVer := metaBucket.Get([]byte(fmt.Sprintf("%s_version", kind)))
		if len(deployedVer) == 0 {
			return nil
		}

		ver = bige.UInt32FromByte(deployedVer)

		return nil
	})
	return ver
}

func (cs *CoreSail) Core_CreateProjectKey(ns, app string) string {
	return fmt.Sprintf("%s_%s", ns, app)
}
