package command

import v1 "github.com/codekidx/sailor/pkg/core/v1"

type CLIConfig struct {
	// Manifest is used to know details about different environment sailor is hosted in
	Manifest v1.SailorManifest `json:"manifest"`

	// SailorRoot is mostly ~/.sailor
	SailorRoot string `json:"-"`

	// Env is the current selected environment by the user
	Env string `json:"env"`

	// SailorHost is the host of the current selected environment, splatted for ease
	// of use
	SailorHost string `json:"-"`

	// SailorClient is the REST API client created with SailorHost
	SailorClient *v1.CoreAPIClient `json:"-"`

	// Token is the admin/user token fetched after logged in, it works until
	// it expires!
	Token string `json:"token"`

	// KeyPairs is used to fetch a resource from namespace and app
	// @key = combination of ns and app separated by '-'
	// @value = has access key and secret key of the projects
	KeyPairs map[string]v1.KeyPair `json:"key_pairs"`
}
