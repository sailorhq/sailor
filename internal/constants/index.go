package constants

const (
	// BUCKET_META contains meta information about a project.
	// Each project will have a meta bucket and it will contain:
	//
	// 1. access key
	// 2. secret key
	// 3. current deployed version of all resources
	BUCKET_META = "_meta"

	// BUCKET_PROJECTS will have list of all projects inside a
	// sailor instance, this bucket lives inside [BUCKET_ADMIN]
	BUCKET_PROJECTS = "projects"

	// BUCKET_SETTING is collection of sailor wide settings this
	// buckect lives inside [BUCKET_ADMIN]
	BUCKET_SETTING = "settings"

	// KEY_SETTING is used to save sailor wide settings this
	// key lives inside [BUCKET_ADMIN]
	KEY_SETTING = "settings"

	// BUCKET_RESOURCE contains all the resources present in a project
	BUCKET_RESOURCE = "resource"

	// BUCKET_RELEASE contains all the release tags provided while creating a deployment
	BUCKET_RELEASE = "release"

	// BUCKET_DEPLOYMENT contains buckets for each resource
	// 		- config
	// 		- secret
	// 		- {key}-misc
	// and each sub-bucket will contain list of deployments per resource
	BUCKET_DEPLOYMENT = "deployments"
)

// -- RESOURCE TYPES INSIDE SAILOR --
const (
	KindConfig = "config"
	KindSecret = "secret"
	KindMisc   = "misc"
)
