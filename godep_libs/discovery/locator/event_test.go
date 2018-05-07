package locator

import (
	"motify_core_api/godep_libs/discovery/provider"
	. "gopkg.in/check.v1"
)

type TestEventSuit struct{}

var _ = Suite(&TestEventSuit{})

func (s *TestEventSuit) TestNewLocationFromKV_EmptyValue(c *C) {
	c.Logf("Empty value is a valid - delete events may produce such")

	kv := provider.KV{
		Service: provider.Service{
			Name: "test",
		},
	}
	expected := Location{
		Service: kv.Service,
	}

	location, err := newLocationFromKV(kv, TypeUnknown)
	c.Assert(err, IsNil)
	c.Assert(location, Equals, expected)
}

func (s *TestEventSuit) TestNewLocationFromKV_InvalidValue(c *C) {
	c.Logf("Empty value is a valid - delete events may produce such")

	kv := provider.KV{
		Value: "foo",
	}

	location, err := newLocationFromKV(kv, TypeUnknown)
	c.Assert(err, NotNil)
	c.Assert(location, Equals, Location{})
}

func (s *TestEventSuit) TestNewLocationFromKV_TypeMain(c *C) {
	c.Logf("Correct enpoint value should be taken")

	kv := provider.KV{
		Value: `{"endpoint_main": "go1.dc:8080", "login": "login", "pass": "123"}`,
	}
	expected := Location{
		Endpoint: "go1.dc:8080",
		Login:    "login",
		Password: "123",
	}

	location, err := newLocationFromKV(kv, TypeAppMain)
	c.Assert(err, IsNil)
	c.Assert(location, Equals, expected)
}

func (s *TestEventSuit) TestNewLocationFromKV_TypeAdditional(c *C) {
	c.Logf("Correct enpoint value should be taken")

	kv := provider.KV{
		Value: `{"endpoint_additional": "go1.dc:8080"}`,
	}
	expected := Location{
		Endpoint: "go1.dc:8080",
	}

	location, err := newLocationFromKV(kv, TypeAppAdditional)
	c.Assert(err, IsNil)
	c.Assert(location, Equals, expected)
}

func (s *TestEventSuit) TestNewEvent(c *C) {
	discoveryEvent := &provider.Event{
		Type: provider.EventPut,
		KVs: []provider.KV{
			provider.KV{
				Value: `{"endpoint_main": "go1.dc:8080"}`,
			},
			provider.KV{
				Value: `{"endpoint_main": "go2.dc:8080"}`,
			},
		},
	}
	expected := &Event{
		Type: provider.EventPut,
		Locations: []Location{
			Location{
				Endpoint: "go1.dc:8080",
			},
			Location{
				Endpoint: "go2.dc:8080",
			},
		},
	}

	event, err := newEvent(discoveryEvent, TypeAppMain)
	c.Assert(err, IsNil)
	c.Assert(event, DeepEquals, expected)
}
