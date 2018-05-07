package locator

import (
	"context"
)

// dummyLocator is a locator which does nothing
type dummyLocator struct{}

// Watch returns channel with no data, which is closed via context cancellation.
func (d *dummyLocator) Watch(ctx context.Context, serviceName string, t EndpointType, filter *KeyFilter) <-chan *Event {
	c := make(chan *Event)
	go func() {
		<-ctx.Done()
		close(c)
	}()
	return c
}

// Get does nothing
func (d *dummyLocator) Get(ctx context.Context, serviceName string, t EndpointType, filter *KeyFilter) ([]Location, error) {
	return []Location{}, nil
}
