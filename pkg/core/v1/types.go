package v1

// EnvConfig has name
type EnvConfig struct {
	Name string `json:"name"`
	Host string `json:"host"`
}

// SailorManifest is the manifest with which Sailor is configured
// Right now we only have environment mentioned here
type SailorManifest struct {
	Envs []EnvConfig `json:"envs"`
}
