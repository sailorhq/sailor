package handlers

import (
	"github.com/sailorhq/sailor/internal/types"
	v1 "github.com/sailorhq/sailor/pkg/core/v1"
)

type MockSail struct {
	GetSailorSettingFunc func() (*v1.SailorSetting, error)
	CreateProjectFunc    func(ns, app string) error
	GetProjectsFunc      func() ([]types.Project, error)
	GetCurrentDeployedVersionFunc func(projectKey, kind string) uint32
	GetResourceKeysFunc  func(projectKey string) ([]string, error)
	GetPinnedVersionFunc func(projectKey, kind, releaseVer string) (uint32, error)
	BuildResourceFunc    func(projectKey, resourceKey, versionKey string, onTopOfLastDeployment bool, overrideMaxVersion []byte) (string, uint32)
}

func (m *MockSail) GetSailorSetting() (*v1.SailorSetting, error) {
	if m.GetSailorSettingFunc != nil {
		return m.GetSailorSettingFunc()
	}
	return nil, nil
}

func (m *MockSail) CreateProject(ns, app string) error {
	if m.CreateProjectFunc != nil {
		return m.CreateProjectFunc(ns, app)
	}
	return nil
}

func (m *MockSail) GetProjects() ([]types.Project, error) {
	if m.GetProjectsFunc != nil {
		return m.GetProjectsFunc()
	}
	return nil, nil
}

func (m *MockSail) GetCurrentDeployedVersion(projectKey, kind string) uint32 {
	if m.GetCurrentDeployedVersionFunc != nil {
		return m.GetCurrentDeployedVersionFunc(projectKey, kind)
	}
	return 0
}

func (m *MockSail) GetResourceKeys(projectKey string) ([]string, error) {
	if m.GetResourceKeysFunc != nil {
		return m.GetResourceKeysFunc(projectKey)
	}
	return nil, nil
}

func (m *MockSail) GetPinnedVersion(projectKey, kind, releaseVer string) (uint32, error) {
	if m.GetPinnedVersionFunc != nil {
		return m.GetPinnedVersionFunc(projectKey, kind, releaseVer)
	}
	return 0, nil
}

func (m *MockSail) BuildResource(projectKey, resourceKey, versionKey string, onTopOfLastDeployment bool, overrideMaxVersion []byte) (string, uint32) {
	if m.BuildResourceFunc != nil {
		return m.BuildResourceFunc(projectKey, resourceKey, versionKey, onTopOfLastDeployment, overrideMaxVersion)
	}
	return "", 0
}
