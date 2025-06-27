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

type SailorState struct {
	Meta    SailorMeta        `json:"meta"`
	Configs map[string]any    `json:"configs"`
	Secrets map[string]string `json:"secrets"`
}
