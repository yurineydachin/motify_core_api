package provider

import (
	"context"
)

// Mock is mock provider
type Mock struct {
	RegisterValuesCallback func(ctx context.Context, kvs ...KV) error
	GetCallback            func(ctx context.Context, filter KeyFilter) ([]KV, error)
	WatchCallback          func(ctx context.Context, filter KeyFilter) <-chan *Event
}

var _ IProvider = &Mock{}

// RegisterValues calls RegisterValuesCallback
func (p *Mock) RegisterValues(ctx context.Context, kvs ...KV) error {
	return p.RegisterValuesCallback(ctx, kvs...)
}

// Get calls GetCallback
func (p *Mock) Get(ctx context.Context, filter KeyFilter) ([]KV, error) {
	return p.GetCallback(ctx, filter)
}

// Watch calls WatchCallback
func (p *Mock) Watch(ctx context.Context, filter KeyFilter) <-chan *Event {
	return p.WatchCallback(ctx, filter)
}
