// backport of motify_core_api/godep_libs/go-cache for using `metrics` library
package cache

import (
	"container/list"
	"strings"
	"sync"
	"time"

	"motify_core_api/godep_libs/go-errors/errors"
	"motify_core_api/godep_libs/metrics"
	"motify_core_api/godep_libs/metrics/structcachemon"
)

// StructCache is simple storage with locking
type StructCache struct {
	entries           map[string]*list.Element
	lruList           *list.List
	ticker            *time.Ticker
	lock              sync.RWMutex
	limit             int
	quitCollectorChan chan struct{} // connection count metric
}

func NewStructCache(limit int) (*StructCache, error) {
	cache := &StructCache{
		limit:   limit,
		ticker:  time.NewTicker(5 * time.Minute), // @todo make it changeable param
		lock:    sync.RWMutex{},
		entries: make(map[string]*list.Element),
		lruList: list.New(),

		quitCollectorChan: make(chan struct{}, 1),
	}

	go cache.collector()

	return cache, nil
}

func (cache *StructCache) SetLimit(limit int) {
	if cache.limit != limit {
		cache.lock.Lock()
		defer cache.lock.Unlock()
		cache.limit = limit
		if cache.lruList.Len() > limit {
			cache.trim()
		}
	}
}

// GetWithTime returns value and create time(UTC) by key
func (cache *StructCache) GetWithTime(key *Key) (interface{}, time.Time, bool) {
	var (
		timeStart = time.Now()
		created   time.Time
		data      interface{}
	)

	cache.lock.RLock()
	el, ok := cache.entries[key.ID()]
	if ok {
		if entry, eok := el.Value.(*Entry); eok {
			created = entry.CreateDate
			if !entry.IsValid() {
				cache.lock.RUnlock()
				cache.Remove(key)
				ok = false
			} else {
				data = entry.Data
				cache.lock.RUnlock()
				cache.lock.Lock()
				cache.lruList.MoveToFront(el)
				cache.lock.Unlock()
			}
		} else {
			cache.lock.RUnlock()
			ok = false
		}
	} else {
		cache.lock.RUnlock()
	}

	structcachemon.ResponseTime.WithLabelValues("get", key.Set).Observe(metrics.SinceMs(timeStart))
	if ok {
		structcachemon.HitCount.WithLabelValues(key.Set).Inc()
	} else {
		structcachemon.MissCount.WithLabelValues(key.Set).Inc()
	}

	return data, created, ok
}

// Get returns value by key
func (cache *StructCache) Get(key *Key) (interface{}, bool) {
	data, _, ok := cache.GetWithTime(key)

	return data, ok
}

// Count data from cache
func (cache *StructCache) Count() int {
	cache.lock.RLock()
	count := cache.lruList.Len()
	cache.lock.RUnlock()
	return count
}

// Find search key by mask
func (cache *StructCache) Find(maskedKey string, limit int) []string {
	result := make([]string, 0, limit)

	cache.lock.RLock()
	for key := range cache.entries {
		if strings.Contains(strings.ToLower(key), maskedKey) {
			result = append(result, key)
			limit--
		}

		if limit == 0 {
			break
		}
	}
	cache.lock.RUnlock()

	return result
}

// Put puts data into storage
func (cache *StructCache) Put(data interface{}, key *Key, ttl time.Duration) error {
	if ttl <= 0 || cache.limit <= 0 {
		return errors.New("Cannot put element (ttl or cache limit is not assign)")
	}

	timeStart := time.Now()

	cache.lock.Lock()
	defer cache.lock.Unlock()

	entitiesCount := cache.lruList.Len()

	if entitiesCount >= cache.limit {
		cache.trim()
	}

	k := key.ID()
	if el, ok := cache.entries[k]; ok {
		cache.lruList.MoveToFront(el)
		if entry, eok := el.Value.(*Entry); eok {
			entry.EndDate = time.Now().Unix() + int64(ttl.Seconds())
			entry.Data = data
			return nil
		}
	}

	entry := CreateEntry(key, time.Now().Unix()+int64(ttl.Seconds()), data)
	el := cache.lruList.PushFront(entry)
	cache.entries[k] = el

	structcachemon.ResponseTime.WithLabelValues("put", key.Set).Observe(metrics.Ms(time.Since(timeStart)))
	structcachemon.ItemNumber.WithLabelValues(key.Set).Inc()

	return nil
}

func (cache *StructCache) Close() {
	cache.quitCollectorChan <- struct{}{}
}

// trim remove least recently used elements from cache and leave 'limit - 1' elements, to have a change to put one element
func (cache *StructCache) trim() {
	for cache.lruList.Len() >= cache.limit && cache.lruList.Len() > 0 {
		el := cache.lruList.Back()
		if el != nil {
			if entry, ok := el.Value.(*Entry); ok {
				delete(cache.entries, entry.Key.ID())
				cache.lruList.Remove(el)
				structcachemon.ItemNumber.WithLabelValues(entry.Key.Set).Dec()
			}
		}
	}
}

// Remove removes value by key
func (cache *StructCache) Remove(key *Key) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	cache.remove(key)
}

func (cache *StructCache) remove(key *Key) {
	var k = key.ID()
	if el, ok := cache.entries[k]; ok {
		delete(cache.entries, k)
		cache.lruList.Remove(el)
		structcachemon.ItemNumber.WithLabelValues(key.Set).Dec()
	}
}

func (cache *StructCache) collector() {
	for {
		select {
		case <-cache.ticker.C:
			cache.lock.RLock()
			for _, el := range cache.entries {
				if entry, ok := el.Value.(*Entry); ok {
					if !entry.IsValid() {
						cache.lock.RUnlock()
						cache.Remove(entry.Key)
						cache.lock.RLock()
					}
				}
			}
			cache.lock.RUnlock()
		case <-cache.quitCollectorChan:
			return
		}
	}
}
