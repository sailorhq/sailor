package types

type SailorFile struct {
	Project ProjectDetails          `json:"project"`
	Config  ResourceFile            `json:"config"`
	Secret  ResourceFile            `json:"secret"`
	Misc    map[string]ResourceFile `json:"misc"`
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

type ResourceVersion struct {
	Config string `json:"config"`
	Secret string `json:"secret"`
	Misc   string `json:"misc"`
}
