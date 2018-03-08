package locator

import (
	"godep.lzd.co/discovery/provider"
	"godep.lzd.co/discovery/registrator"
)

// Event contains data about service location change
type Event struct {
	// Type is a type of event
	Type provider.EventType
	// Locations contains service instances location updates
	Locations []Location
}

// Location contains service ID, served endpoint and optional credentials
type Location struct {
	Service  provider.Service
	Endpoint string
	Login    string
	Password string
}

func newLocationFromKV(kv provider.KV, t EndpointType) (Location, error) {
	l := Location{
		Service: kv.Service,
	}
	if kv.Value == "" {
		// Value can be empty in case of delete event
		return l, nil
	}

	v, err := registrator.NewDiscoveryValueFromString(kv.Value)
	if err != nil {
		return Location{}, err
	}
	l.Login = v.Login
	l.Password = v.Password

	switch t {
	case TypeAppMain, TypeSystem, TypeSystemMain, TypeExternal:
		l.Endpoint = v.EndpointMain
	case TypeAppAdditional, TypeSystemAdditional:
		l.Endpoint = v.EndpointAdditional
	}

	return l, nil
}

func newLocationsFromKVs(kvs []provider.KV, t EndpointType) ([]Location, error) {
	locations := make([]Location, 0, len(kvs))
	for _, kv := range kvs {
		l, err := newLocationFromKV(kv, t)
		if err != nil {
			return nil, err
		}
		locations = append(locations, l)
	}
	return locations, nil
}

func newEvent(event *provider.Event, t EndpointType) (*Event, error) {
	locations, err := newLocationsFromKVs(event.KVs, t)
	if err != nil {
		return nil, err
	}

	res := &Event{
		Type:      event.Type,
		Locations: locations,
	}
	return res, nil
}
