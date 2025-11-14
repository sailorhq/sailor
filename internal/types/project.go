package types

type SailorFile struct {
	Project ProjectDetails          `json:"project"`
	Config  ResourceFile            `json:"config,omitempty"`
	Secret  ResourceFile            `json:"secret,omitempty"`
	Misc    map[string]ResourceFile `json:"misc,omitempty"`
}

type ProjectDetails struct {
	Namespace string `json:"ns"`
	App       string `json:"app"`
}

type ResourceFile struct {
	File   string          `json:"file"`
	Name   string          `json:"name,omitempty"`
	Schema *map[string]any `json:"schema,omitempty"`
}

type SailorLockFile struct {
	Environments map[string]ResourceVersion `json:"envs"`
}

type LockVersion struct {
	Version int    `json:"version"`
	Hash    string `json:"hash,omitempty"`
}

type ResourceVersion struct {
	Config *LockVersion           `json:"config,omitempty"`
	Secret *LockVersion           `json:"secret,omitempty"`
	Misc   map[string]LockVersion `json:"misc,omitempty"`
}
