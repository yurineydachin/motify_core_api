package locator

import (
	"context"

	"motify_core_api/godep_libs/discovery"
	"motify_core_api/godep_libs/discovery/provider"
)

// ILocator is service locator interface
type ILocator interface {
	// Watch waits for service location updates and sends events to the channel
	// Watch is canceled via context cancellation.
	Watch(ctx context.Context, serviceName string, t EndpointType, filter *KeyFilter) <-chan *Event

	// Get returns service instance locations.
	Get(ctx context.Context, serviceName string, t EndpointType, filter *KeyFilter) ([]Location, error)
}

type locator struct {
	provider provider.IProvider
	logger   discovery.ILogger
}

// New returns new locator.
// If provider is nil (for example when it could not be initialized) - returns dummy locator.
func New(p provider.IProvider, logger discovery.ILogger) ILocator {
	if logger == nil {
		logger = discovery.NewNilLogger()
	}
	if p == nil {
		return &dummyLocator{}
	}

	return &locator{
		provider: p,
		logger:   logger,
	}
}

// Get returns service instances discovery data.
func (l *locator) Get(ctx context.Context, serviceName string, t EndpointType, filter *KeyFilter) ([]Location, error) {
	keyFilter := discoveryKeyFilter(serviceName, t, filter)
	kvs, err := l.provider.Get(ctx, keyFilter)
	if err != nil {
		return nil, err
	}

	return newLocationsFromKVs(kvs, t)
}

// Watch runs discovery watcher on discovery namespace for service and prepares locations results
func (l *locator) Watch(ctx context.Context, serviceName string, t EndpointType, filter *KeyFilter) <-chan *Event {
	out := make(chan *Event)
	go l.doWatch(ctx, out, serviceName, t, filter)

	return out
}

func (l *locator) doWatch(ctx context.Context, out chan<- *Event, serviceName string, t EndpointType, filter *KeyFilter) {
	l.logger.Debugf("watching for '%s' service discovery updates", serviceName)
	defer l.logger.Debugf("finish watching for '%s' service", serviceName)

	keyFilter := discoveryKeyFilter(serviceName, t, filter)

	events := l.provider.Watch(ctx, keyFilter)
	for event := range events {
		res, err := newEvent(event, t)
		if err != nil {
			l.logger.Warningf("'%s' service: %s", serviceName, err)
			continue
		}

		select {
		case <-ctx.Done():
			// don't want to block on channel write in case of done context
			l.logger.Debugf("'%s' context is Done, ", serviceName)
		case out <- res:
		}
	}
	// close output channel when done
	close(out)
}
