package v1

import (
	"encoding/json"
	"errors"
)

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

func serverMessageToErr(b []byte) error {
	var errMsg map[string]string
	if err := json.Unmarshal(b, &errMsg); err != nil {
		return err
	}
	return errors.New(errMsg["message"])
}
