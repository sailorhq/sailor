package handlers

const (
	BUCKET_META       = "_meta"
	BUCKET_CONFIGS    = "configs"
	BUCKET_SECRETS    = "secrets"
	BUCKET_DIFFS      = "_diffs"
	BUCKET_DEPLOYMENT = "deployments"

	KEY_ACCESS_KEY       = "access_key"
	KEY_DEPLOYED_VERSION = "deploy_ver"
	KEY_RULES            = "rules"
)

type ResponseMessage struct {
	Message string
}
