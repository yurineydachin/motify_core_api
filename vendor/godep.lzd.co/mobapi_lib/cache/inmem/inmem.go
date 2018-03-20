package inmem

import (
	"sync"
	"time"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	structcache "godep.lzd.co/mobapi_lib/cache"

	"godep.lzd.co/go-config"
	"godep.lzd.co/go-dconfig"
)

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
	structCache *structcache.StructCache
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
	c := &Cache{
		LocalCacheLocker: cache.NewLocalCacheLocker(),
	}
	gc, _ := structcache.NewStructCache(options.Limit)
	c.structCache = gc
	c.SetOptions(options)
	return c
}

func (c *Cache) SetOptions(options Options) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.structCache.SetLimit(options.Limit)
	c.ttl = options.TTL
	c.set = options.Set
}

func (c *Cache) Get(key []byte) *cache.CacheEntry {
	strKey := string(key)
	c.mx.RLock()
	structCache := c.structCache
	c.mx.RUnlock()

	if data, ok := structCache.Get(&structcache.Key{
		Set: c.set,
		Pk:  strKey,
	}); ok {
		if ent, ok := data.(*cache.CacheEntry); ok {
			return ent
		}
	}
	return nil
}

func (c *Cache) Put(key []byte, entry *cache.CacheEntry) {
	c.mx.RLock()
	ttl := c.ttl
	structCache := c.structCache
	c.mx.RUnlock()
	structCache.Put(entry, &structcache.Key{Set: c.set, Pk: string(key)}, ttl)
}

func (c *Cache) PutWithTTL(key []byte, entry *cache.CacheEntry, ttl time.Duration) {
	c.mx.RLock()
	structCache := c.structCache
	c.mx.RUnlock()
	structCache.Put(entry, &structcache.Key{Set: c.set, Pk: string(key)}, ttl)
}
