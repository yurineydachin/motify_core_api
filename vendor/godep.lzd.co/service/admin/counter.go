package admin

import (
	"sync/atomic"
	"time"
)

type Counter struct {
	counter  uint64
	rate     uint64
	interval time.Duration
}

func NewCounter(interval time.Duration) *Counter {
	c := &Counter{
		interval: interval,
	}
	go func() {
		for {
			time.Sleep(interval)
			atomic.StoreUint64(&c.rate, atomic.SwapUint64(&c.counter, 0))
		}
	}()
	return c
}

func (c *Counter) Inc(val uint64) {
	atomic.AddUint64(&c.counter, val)
}

func (c *Counter) Rate() uint64 {
	return atomic.LoadUint64(&c.rate)
}
