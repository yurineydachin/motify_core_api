package etcdV3

import (
	"github.com/coreos/etcd/clientv3"
)

// kvAPI is an interface to generate mock
//go:generate mockery -name=kvAPI -case=underscore -inpkg
type kvAPI interface {
	clientv3.KV
}

// kvAPI is an interface to generate mock
//go:generate mockery -name=watcher -case=underscore -inpkg
type watcher interface {
	clientv3.Watcher
}
