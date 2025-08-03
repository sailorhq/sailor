package command

import v1 "github.com/codekidx/sailor/pkg/core/v1"

type CLIConfig struct {
	// Manifest is used to know details about different environment sailor is hosted in
	Manifest v1.SailorManifest

	// SailorRoot is mostly ~/.sailor
	SailorRoot string

	// Env is the current selected environment by the user
	Env string

	// SailorHost is the host of the current selected environment, splatted for ease
	// of use
	SailorHost string

	// SailorClient is the REST API client created with SailorHost
	SailorClient *v1.CoreAPIClient

	// Token is the admin/user token fetched after logged in, it works until
	// it expires!
	Token string
}
