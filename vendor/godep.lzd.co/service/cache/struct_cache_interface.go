// backport of godep.lzd.co/go-cache for using `metrics` library
package cache

import (
	"time"
)

// IStructCache defines required interface for caching module (to be able to store any kind of data, mostly in memory)
type IStructCache interface {
	Get(key *Key) (data interface{}, ok bool)
	GetWithTime(key *Key) (data interface{}, dt time.Time, ok bool)
	Put(data interface{}, key *Key, ttl time.Duration) error
	Remove(key *Key)
	Count() int
	Close()
}

type ILimitSetter interface {
	SetLimit(int)
}
