package inmem

import (
	"sync"
	"time"

	"github.com/sergei-svistunov/gorpc/transport/cache"

	gocache "godep.lzd.co/service/cache"
	"godep.lzd.co/service/config"
	"godep.lzd.co/service/dconfig"
)

type StructCache struct {
	*gocache.StructCache
}

func NewFromFlags(set string) cache.ICache {
	if v, _ := config.GetBool("inmem-cache-enabled"); !v {
		return nil
	}
	cacheOptions := Options{
		Limit: 1200,
		TTL:   time.Second * 30,
		Set:   set,
	}
	transportCache := New(cacheOptions)
	if set != "" {
		set += "-"
	}
	dconfig.RegisterInt(set+"inmemcache-limit", "In-memory cache (set='"+cacheOptions.Set+"') limit of entities count", 1200, func(val int) {
		cacheOptions.Limit = val
		transportCache.SetOptions(cacheOptions)
	})
	dconfig.RegisterDuration(set+"inmemcache-ttl", "In-memory cache (set='"+cacheOptions.Set+"') ttl (duration)", time.Second*30, func(val time.Duration) {
		cacheOptions.TTL = val
		transportCache.SetOptions(cacheOptions)
	})
	return transportCache
}

type Cache struct {
	structCache gocache.IStructCache
	ttl         time.Duration
	mx          sync.RWMutex
	set         string
	*cache.LocalCacheLocker
}

type Options struct {
	Limit int
	TTL   time.Duration
	Set   string
}

func New(options Options) *Cache {
	inmem := &Cache{
		LocalCacheLocker: cache.NewLocalCacheLocker(),
	}
	gc, _ := gocache.NewStructCache(options.Limit)
	inmem.structCache = gc
	inmem.SetOptions(options)
	return inmem
}

func (inmem *Cache) SetOptions(options Options) {
	inmem.mx.Lock()
	defer inmem.mx.Unlock()
	if c, ok := inmem.structCache.(gocache.ILimitSetter); ok {
		c.SetLimit(options.Limit)
	}
	inmem.ttl = options.TTL
	inmem.set = options.Set
}

func (inmem *Cache) Get(key []byte) *cache.CacheEntry {
	strKey := string(key)
	inmem.mx.RLock()
	structCache := inmem.structCache
	inmem.mx.RUnlock()

	if data, ok := structCache.Get(&gocache.Key{
		Set: inmem.set,
		Pk:  strKey,
	}); ok {
		if ent, ok := data.(*cache.CacheEntry); ok {
			return ent
		}
	}
	return nil
}

func (inmem *Cache) Put(key []byte, entry *cache.CacheEntry) {
	inmem.mx.RLock()
	ttl := inmem.ttl
	structCache := inmem.structCache
	inmem.mx.RUnlock()
	structCache.Put(entry, &gocache.Key{Set: inmem.set, Pk: string(key)}, ttl)
}

func (inmem *Cache) PutWithTTL(key []byte, entry *cache.CacheEntry, ttl time.Duration) {
	inmem.mx.RLock()
	structCache := inmem.structCache
	inmem.mx.RUnlock()
	structCache.Put(entry, &gocache.Key{Set: inmem.set, Pk: string(key)}, ttl)
}
