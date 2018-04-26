package inmem_test

import (
	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/mobapi_lib/cache/inmem"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	key      = "myKey"
	workTime = 500 * time.Millisecond
	ttl      = 1 * time.Second
	value    = "boo"
)

func get(c *inmem.Cache, k []byte) (interface{}, bool) {
	c.Lock(k)
	defer c.Unlock(k)
	if entry := c.Get(k); entry != nil {
		return entry.Body, true
	}
	time.Sleep(workTime)
	c.Put(k, &cache.CacheEntry{Body: value})
	return value, false
}

func TestConsistentGet(t *testing.T) {
	opts := inmem.Options{
		Limit: 1000,
		TTL:   24 * time.Hour,
	}
	cache := inmem.New(opts)
	for i := 0; i < 10000; i++ {
		v, hit := get(cache, []byte(key))
		if i == 0 && hit {
			t.Error("Should miss for first get")
		}
		if i > 0 && !hit {
			t.Error("Should hit")
		}
		if v.(string) != value {
			t.Errorf("Should be eq. '%v'", value)
		}
	}
}

func _TestParallelGet(t *testing.T) {
	opts := inmem.Options{
		Limit: 1000,
		TTL:   ttl,
	}
	cache := inmem.New(opts)

	var (
		hit_totals  int64
		miss_totals int64
	)

	var wg sync.WaitGroup

	for i := 0; i < 100000; i++ {
		wg.Add(1)
		go func(c int) {
			defer wg.Done()
			v, hit := get(cache, []byte(key+strconv.Itoa(c%1500)))
			if hit {
				atomic.AddInt64(&hit_totals, 1)
			} else {
				atomic.AddInt64(&miss_totals, 1)
			}
			if v.(string) != value {
				t.Errorf("Should be eq. '%v'", value)
			}
		}(i)
	}
	wg.Wait()

	log.Printf("hit: %d, miss: %d", hit_totals, miss_totals)
}

func _TestDeadlock(t *testing.T) {
	opts := inmem.Options{
		Limit: 1000,
		TTL:   ttl,
	}
	cache := inmem.New(opts)
	concurrency := 1500

	a := make(chan bool, concurrency)
	stop := make(chan bool, 1)

	go func() {
		for i := 0; ; i++ {
			a <- true
			go func(c int) {
				result := make(chan interface{}, 1)
				go func() {
					v, _ := get(cache, []byte(key+strconv.Itoa(c%1500)))
					result <- v
				}()
				select {
				case v := <-result:
					if v.(string) != value {
						t.Errorf("Should be eq. '%v'", value)
						stop <- true
					}
				case <-time.After(5 * workTime):
					t.Errorf("Get timeout")
					stop <- true
				}
				<-a
			}(i)
			if i%10000 == 0 {
				log.Printf("%d requests sent", i)
			}
		}
	}()

	select {
	case <-time.After(5 * 60 * time.Second):
	case <-stop:
	}
}
