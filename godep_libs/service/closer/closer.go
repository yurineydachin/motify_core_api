package closer

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var globalCloser = newCloser()

func Add(f func()) {
	globalCloser.Add(f)
}

func Wait() {
	globalCloser.Wait()
}

func CloseAll() {
	globalCloser.CloseAll()
}

type closer struct {
	sync.Mutex
	once  sync.Once
	done  chan struct{}
	funcs []func()
}

func newCloser() *closer {
	c := &closer{done: make(chan struct{})}
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM)
		<-ch
		signal.Stop(ch)
		c.CloseAll()
	}()
	return c
}

func (c *closer) Add(f func()) {
	c.Lock()
	c.funcs = append(c.funcs, f)
	c.Unlock()
}

func (c *closer) Wait() {
	select {
	case <-c.done:
	}
}

func (c *closer) CloseAll() {
	c.once.Do(func() {
		defer close(c.done)

		c.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.Unlock()

		for _, f := range funcs {
			f()
		}
	})
}
