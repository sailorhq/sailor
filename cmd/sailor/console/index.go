package console

import (
	"github.com/fasthttp/router"
	bolt "go.etcd.io/bbolt"
)

type Console struct {
	dbconns map[string]*bolt.DB
}

func Initialize(r *router.Group) error {
	return nil
}
