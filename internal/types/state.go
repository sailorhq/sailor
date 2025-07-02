package types

import (
	"time"
)

type SailorMeta struct {
	Version string `json:"version"`
}

type SailorOpts struct {
	Logging        bool
	RefreshTimeout time.Duration
	BackupURL      string
	AccessKey      string
}

type Deployment struct {
	Description string `json:"description"`
	Version     string `json:"version"`
	Deployed    bool   `json:"deployed"`
	Diff        string `json:"diff"`
	CreatedAt   string `json:"created_at"`
	CreatedBy   string `json:"created_by"`
	DeployedAt  string `json:"deployed_at"`
	DeployedBy  string `json:"deployed_by"`
}

type Secret struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SailorState struct {
	Meta        SailorMeta     `json:"meta"`
	Configs     map[string]any `json:"configs"`
	Secrets     []Secret       `json:"secrets"`
	AccessKey   string         `json:"access_key"`
	SecretKey   string         `json:"secret_key"`
	Rules       string         `json:"rules"`
	Policy      string         `json:"policy"`
	Deployments []Deployment   `json:"deployments"`
}
