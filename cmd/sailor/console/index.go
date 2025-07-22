package console

import bolt "go.etcd.io/bbolt"

type Console struct {
	dbconns map[string]*bolt.DB
}
