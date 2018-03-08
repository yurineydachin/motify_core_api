package locator

import (
	"context"
)

// Mock is a struct for mocking locator interface
type Mock struct {
	WatchCallback func(ctx context.Context, serviceName string, t EndpointType, filter *KeyFilter) <-chan *Event
	GetCallback   func(ctx context.Context, serviceName string, t EndpointType, filter *KeyFilter) ([]Location, error)
}

var _ ILocator = &Mock{}

// Watch calls WatchCallback
func (l *Mock) Watch(ctx context.Context, serviceName string, t EndpointType, filter *KeyFilter) <-chan *Event {
	return l.WatchCallback(ctx, serviceName, t, filter)
}

// Get calls GetCallback
func (l *Mock) Get(ctx context.Context, serviceName string, t EndpointType, filter *KeyFilter) ([]Location, error) {
	return l.GetCallback(ctx, serviceName, t, filter)
}
