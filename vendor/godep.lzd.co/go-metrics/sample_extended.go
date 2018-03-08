package metrics

import (
	"container/heap"
	"math"
	"math/rand"
	"time"
)

//ExpDecaySampleExtended is extended version of ExpDecaySample
type ExpDecaySampleExtended struct {
	*ExpDecaySample
	rescaleThreshold time.Duration
}

// NewExpDecaySampleExtended constructs a new exponentially-decaying sample with the
// given reservoir size, alpha, and recalculation threshold
func NewExpDecaySampleExtended(reservoirSize int, alpha float64, rescaleThreshold time.Duration) Sample {
	if UseNilMetrics {
		return NilSample{}
	}
	s := &ExpDecaySampleExtended{
		ExpDecaySample: &ExpDecaySample{
			alpha:         alpha,
			reservoirSize: reservoirSize,
			t0:            time.Now(),
			values:        make(expDecaySampleHeap, 0, reservoirSize),
		},
	}
	s.rescaleThreshold = rescaleThreshold

	s.t1 = time.Now().Add(rescaleThreshold)
	return s
}

// update samples a new value at a particular timestamp.  This is a method all
// its own to facilitate testing.
func (s *ExpDecaySampleExtended) update(t time.Time, v int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.count++
	if len(s.values) == s.reservoirSize {
		heap.Pop(&s.values)
	}
	heap.Push(&s.values, expDecaySample{
		k: math.Exp(t.Sub(s.t0).Seconds()*s.alpha) / rand.Float64(),
		v: v,
	})
	if t.After(s.t1) {
		values := s.values
		t0 := s.t0
		s.values = make(expDecaySampleHeap, 0, s.reservoirSize)
		s.t0 = t
		s.t1 = s.t0.Add(s.rescaleThreshold)
		for _, v := range values {
			v.k = v.k * math.Exp(-s.alpha*float64(s.t0.Sub(t0).Seconds()))
			heap.Push(&s.values, v)
		}
	}
}
