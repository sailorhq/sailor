package handlers

import (
	bolt "go.etcd.io/bbolt"
)

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

type SailorCore struct {
	// TODO :: change value to socket type
	dbconns map[string]*bolt.DB
}

func NewSailorCore() *SailorCore {
	return &SailorCore{
		dbconns: make(map[string]*bolt.DB),
	}
}

type ResponseMessage struct {
	Message string
}
