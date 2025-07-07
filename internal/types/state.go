package types

import (
	"time"
)

type SailorMeta struct {
	Version string `json:"version"`
}

type SailorOpts struct {
	Logging        bool
	AvoidRefresh   bool
	RefreshTimeout time.Duration
	BackupURL      string
	AccessKey      string
	SecretKey      string
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

type AdminBackupState struct {
	Bucket string `json:"bucket"`
}

type ListAppsResponse struct {
	Apps             []string          `json:"apps"`
	AdminBackupState *AdminBackupState `json:"admin_backup_state"`
}

type AdminSailorState struct {
	Meta        SailorMeta     `json:"meta"`
	Configs     map[string]any `json:"configs"`
	Secrets     []Secret       `json:"secrets"`
	AccessKey   string         `json:"access_key"`
	SecretKey   string         `json:"secret_key"`
	Rules       string         `json:"rules"`
	Policy      string         `json:"policy"`
	Deployments []Deployment   `json:"deployments"`
}

type SailorState struct {
	Version string            `json:"config_ver"`
	Config  []byte            `json:"config"`
	Secrets map[string][]byte `json:"secrets"`
}
