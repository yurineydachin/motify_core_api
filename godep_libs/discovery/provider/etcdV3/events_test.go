package etcdV3

import (
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"motify_core_api/godep_libs/discovery/provider"
	. "gopkg.in/check.v1"
)

type TestEventsSuit struct{}

var _ = Suite(&TestEventsSuit{})

func (s *TestEventsSuit) TestExtractNamespace_Valid(c *C) {
	keys := []string{
		fmt.Sprintf("/%s/bla/bla", provider.NamespaceRollout),
		fmt.Sprintf("/%s/bla/bla", provider.NamespaceExportedEntities),
		fmt.Sprintf("/%s/", provider.NamespaceExportedEntities),
	}
	expectedNs := []string{
		provider.NamespaceRollout,
		provider.NamespaceExportedEntities,
		provider.NamespaceExportedEntities,
	}

	for i, key := range keys {
		ns, ok := extractNamespace(key)
		c.Assert(ns, Equals, expectedNs[i])
		c.Assert(ok, Equals, true)
	}
}

func (s *TestEventsSuit) TestExtractNamespace_Invalid(c *C) {
	keys := []string{
		"foo/bla/bla",             // no trailing slash
		"/some_namespace/bla/bla", // invalid namespace
		"/some_namespace",         // no other parts, no trailing slash
		"some_namespace",          // no slash
	}

	for _, key := range keys {
		ns, ok := extractNamespace(key)
		c.Assert(ns, Equals, "")
		c.Assert(ok, Equals, false)
	}
}

func (s *TestEventsSuit) TestDiscoveryKV_ServiceKey(c *C) {
	etcdKV := &mvccpb.KeyValue{
		Key:   []byte("/discovery/app/test_service/stable/shared/common/host:port"),
		Value: []byte("value"),
	}

	kv, err := discoveryKV(etcdKV)
	c.Assert(err, IsNil)
	c.Assert(kv, Equals, provider.KV{
		Namespace: provider.NamespaceDiscovery,
		Value:     "value",
		RawKey:    string(etcdKV.Key),
		Service: provider.Service{
			Name:         "test_service",
			Type:         provider.TypeApp,
			RolloutType:  provider.RolloutTypeStable,
			Owner:        provider.DefaultOwner,
			ClusterType:  provider.DefaultClusterType,
			InstanceName: provider.InstanceName("host:port"),
		},
	})
}

func (s *TestEventsSuit) TestDiscoveryKV_Rollout(c *C) {
	etcdKV := &mvccpb.KeyValue{
		Key:   []byte("/rollout/segregation/1"),
		Value: []byte("stable"),
	}

	kv, err := discoveryKV(etcdKV)
	c.Assert(err, IsNil)
	c.Assert(kv, Equals, provider.KV{
		Namespace: provider.NamespaceRollout,
		Value:     "stable",
		RawKey:    string(etcdKV.Key),
	})
}

func (s *TestEventsSuit) TestDiscoveryKV_UnknownNamespace(c *C) {
	etcdKV := &mvccpb.KeyValue{
		Key: []byte("/foo/bar"),
	}

	_, err := discoveryKV(etcdKV)
	c.Assert(err, NotNil)
}

func (s *TestEventsSuit) TestNewEventFromKVs_UnknownNamespace(c *C) {
	etcdKVs := []*mvccpb.KeyValue{&mvccpb.KeyValue{
		Key: []byte("/foo/bar"),
	}}

	_, err := newEventFromKVs(etcdKVs)
	c.Assert(err, NotNil)
}

func (s *TestEventsSuit) TestNewEventFromKVs_OK(c *C) {
	etcdKVs := []*mvccpb.KeyValue{
		&mvccpb.KeyValue{
			Key:   []byte("/discovery/app/test_service/stable/shared/common/host0"),
			Value: []byte("value0"),
		},
		&mvccpb.KeyValue{
			Key:   []byte("/discovery/app/test_service/stable/shared/common/host1"),
			Value: []byte("value1"),
		},
	}

	event, err := newEventFromKVs(etcdKVs)
	c.Assert(err, IsNil)
	c.Assert(event.Type, Equals, provider.EventPut)
	for i, kv := range event.KVs {
		c.Assert(kv, Equals, provider.KV{
			Namespace: provider.NamespaceDiscovery,
			Value:     string(etcdKVs[i].Value),
			RawKey:    string(etcdKVs[i].Key),
			Service: provider.Service{
				Name:         "test_service",
				Type:         provider.TypeApp,
				RolloutType:  provider.RolloutTypeStable,
				Owner:        provider.DefaultOwner,
				ClusterType:  provider.DefaultClusterType,
				InstanceName: provider.InstanceName("host" + fmt.Sprintf("%d", i)),
			},
		})
	}
}

func (s *TestEventsSuit) TestWatchResponse_UnknownNamespace(c *C) {
	etcdEvent := &clientv3.Event{
		Type: mvccpb.PUT,
		Kv: &mvccpb.KeyValue{
			Key: []byte("/foo/bar"),
		},
	}

	_, err := watchResponse(etcdEvent)
	c.Assert(err, NotNil)
}

func (s *TestEventsSuit) TestWatchResponse_NoKV(c *C) {
	etcdEvent := &clientv3.Event{}

	_, err := watchResponse(etcdEvent)
	c.Assert(err, NotNil)
}

func (s *TestEventsSuit) TestWatchResponse(c *C) {
	etcdEvent := &clientv3.Event{
		Type: mvccpb.PUT,
		Kv: &mvccpb.KeyValue{
			Key:   []byte("/discovery/app/test_service/stable/shared/common/host:port"),
			Value: []byte("value"),
		},
	}

	event, err := watchResponse(etcdEvent)
	c.Assert(err, IsNil)
	c.Assert(len(event.KVs), Equals, 1)
	c.Assert(event.KVs[0], Equals, provider.KV{
		Namespace: provider.NamespaceDiscovery,
		Value:     "value",
		RawKey:    string(etcdEvent.Kv.Key),
		Service: provider.Service{
			Name:         "test_service",
			Type:         provider.TypeApp,
			RolloutType:  provider.RolloutTypeStable,
			Owner:        provider.DefaultOwner,
			ClusterType:  provider.DefaultClusterType,
			InstanceName: provider.InstanceName("host:port"),
		},
	})
}
