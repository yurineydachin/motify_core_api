package cache

import (
	"sync"
	"time"

	"container/list"

	"errors"
	"godep.lzd.co/metrics"
	"godep.lzd.co/metrics/structcachemon"
)

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

func (cache *StructCache) Get(key *Key) (interface{}, bool) {
	var (
		timeStart = time.Now()
		data      interface{}
	)
	defer func() {
		structcachemon.ResponseTime.WithLabelValues("get", key.Set).Observe(metrics.Ms(time.Since(timeStart)))
	}()

	cache.lock.RLock()
	el, ok := cache.entries[key.ID()]
	if ok {
		if entry, eok := el.Value.(*Entry); eok {
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

	if ok {
		structcachemon.HitCount.WithLabelValues(key.Set).Inc()
	} else {
		structcachemon.MissCount.WithLabelValues(key.Set).Inc()
	}

	return data, ok
}

func (cache *StructCache) Count() int {
	cache.lock.RLock()
	count := cache.lruList.Len()
	cache.lock.RUnlock()
	return count
}

func (cache *StructCache) Put(data interface{}, key *Key, ttl time.Duration) error {
	if ttl <= 0 || cache.limit <= 0 {
		return errors.New("Cannot put element (ttl or cache limit is not assign)")
	}

	timeStart := time.Now()
	defer func() {
		structcachemon.ResponseTime.WithLabelValues("put", key.Set).Observe(metrics.Ms(time.Since(timeStart)))
	}()

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

	structcachemon.ItemNumber.WithLabelValues(key.Set).Inc()

	return nil
}

func (cache *StructCache) Close() {
	cache.quitCollectorChan <- struct{}{}
}

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
