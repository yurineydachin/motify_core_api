package locker

import (
	"sync"
)

type Lock struct {
	lock map[string]*sync.Mutex
	mu   sync.RWMutex
}

func New() *Lock {
	return &Lock{
		lock: make(map[string]*sync.Mutex, 8),
	}
}

func (l *Lock) IsLocked(hash string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	_, ok := l.lock[hash]
	return ok
}

func (l *Lock) Lock(hash string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.lock[hash]; !ok {
		l.lock[hash] = &sync.Mutex{}
	}

	l.lock[hash].Lock()
}

func (l *Lock) Unlock(hash string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.lock[hash].Unlock()
	delete(l.lock, hash)
}
