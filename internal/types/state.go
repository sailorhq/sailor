package types

import (
	"sync"
	"time"
)

type SailorMeta struct {
	Version string `json:"version"`
}

type SailorOpts struct {
	Logging        bool
	RefreshTimeout time.Duration
	BackupURL      string
}

type SailorState struct {
	sync.Mutex
	Meta    SailorMeta        `json:"meta"`
	Configs map[string]any    `json:"configs"`
	Secrets map[string]string `json:"secrets"`
}
