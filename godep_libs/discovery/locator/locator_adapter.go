package locator

import (
	"context"

	etcd "github.com/coreos/etcd/client"
	"motify_core_api/godep_libs/discovery"
	"motify_core_api/godep_libs/discovery/provider"
)

// locatorAdapter is a temp locator adapter for discovery IServiceLocator2 to implement ILocator interface.
// TODO: remove totally after all services move to etcdV3
type locatorAdapter struct {
	info    LocationInfo
	locator IServiceLocator2

	logger discovery.ILogger
}

var _ ILocator = &locatorAdapter{}

// NewAdapter returns new locator for discovery IServiceLocator2
func NewAdapter(locator IServiceLocator2, info LocationInfo, logger discovery.ILogger) ILocator {
	if logger == nil {
		logger = discovery.NewNilLogger()
	}

	return &locatorAdapter{
		info:    info,
		locator: locator,
		logger:  logger,
	}
}

// Get returns service instances discovery data.
func (l *locatorAdapter) Get(ctx context.Context, serviceName string, t EndpointType, _ *KeyFilter) ([]Location, error) {
	if t != TypeAppMain {
		panic("locator adapter supports only good-old HTTP enpoints, stored in etcd2")
	}

	info := LocationInfo{
		Namespace:   l.info.Namespace,
		Venture:     l.info.Venture,
		Environment: l.info.Environment,
		Property:    discovery.NodesProperty,
		ServiceName: serviceName,
	}

	res, err := l.locator.Get(ctx, info)
	if err != nil {
		return nil, err
	}

	locations := make([]Location, 0, len(res))
	for _, info := range res {
		locations = append(locations, newLocationFromServiceInfo(info, serviceName))
	}

	return locations, nil
}

// Watch runs locator for service and prepares locations results
func (l *locatorAdapter) Watch(ctx context.Context, serviceName string, t EndpointType, _ *KeyFilter) <-chan *Event {
	if t != TypeAppMain {
		panic("locator adapter supports only good-old HTTP enpoints, stored in etcd2")
	}
	out := make(chan *Event)
	in := make(chan *etcd.Response, 1)

	info := LocationInfo{
		Namespace:   l.info.Namespace,
		Venture:     l.info.Venture,
		Environment: l.info.Environment,
		Property:    discovery.NodesProperty,
		ServiceName: serviceName,
	}

	go l.locator.Locate(ctx, info, in)
	go l.doWatch(ctx, serviceName, in, out)
	return out
}

func (l *locatorAdapter) doWatch(ctx context.Context, serviceName string, in <-chan *etcd.Response, out chan<- *Event) {
	for resp := range in {
		event := newEventFromResponse(resp, serviceName)
		if event == nil {
			continue
		}

		select {
		case <-ctx.Done():
			// don't want to block on channel write in case of done context
			l.logger.Debugf("'%s' context is Done, ", serviceName)
		case out <- event:
		}
	}
	close(out)
}

func newLocationFromServiceInfo(info ServiceInfo, serviceName string) Location {
	return Location{
		Service: provider.Service{
			Name:         serviceName,
			InstanceName: provider.InstanceName(info.Name),
		},
		Endpoint: info.Value,
	}
}

func newLocationFromEtcdNode(node *etcd.Node, serviceName string) Location {
	return Location{
		Service: provider.Service{
			Name:         serviceName,
			InstanceName: provider.InstanceName(node.Key),
		},
		Endpoint: node.Value,
	}
}

func newEventFromResponse(resp *etcd.Response, serviceName string) *Event {
	switch resp.Action {
	case "get":
		locations := make([]Location, 0, len(resp.Node.Nodes))
		for _, node := range resp.Node.Nodes {
			locations = append(locations, newLocationFromEtcdNode(node, serviceName))
		}
		return &Event{
			Type:      provider.EventPut,
			Locations: locations,
		}

	case "create", "update", "set":
		return &Event{
			Type:      provider.EventPut,
			Locations: []Location{newLocationFromEtcdNode(resp.Node, serviceName)},
		}

	case "delete", "expire":
		return &Event{
			Type:      provider.EventDelete,
			Locations: []Location{newLocationFromEtcdNode(resp.Node, serviceName)},
		}
	}

	return nil
}
