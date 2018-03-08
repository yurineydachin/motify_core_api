package etcdV3

import (
	"godep.lzd.co/discovery/provider"
	. "gopkg.in/check.v1"
)

type TestStorageSuit struct{}

var _ = Suite(&TestStorageSuit{})

func (s *TestStorageSuit) TestKeyPrefix_ByNamespace(c *C) {
	filter := provider.KeyFilter{
		Namespace: "discovery",
	}
	prefix := keyPrefix(filter)
	c.Assert(prefix, Equals, "/discovery/")
}

func (s *TestStorageSuit) TestKeyPrefix_IncompleteFields(c *C) {
	c.Logf("We can't omit some parts of the key - just prefix-based search")
	filter := provider.KeyFilter{
		Namespace:   "discovery",
		Type:        provider.TypeApp,
		RolloutType: "stable",
	}
	prefix := keyPrefix(filter)
	c.Assert(prefix, Equals, "/discovery/app/")
}

func (s *TestStorageSuit) TestKeyPrefix_WithPrefix(c *C) {
	filter := provider.KeyFilter{
		Prefix: "/fooo",
	}
	prefix := keyPrefix(filter)
	c.Assert(prefix, Equals, filter.Prefix)
}

func (s *TestStorageSuit) TestStorageKey(c *C) {
	service := provider.Service{
		Type:         provider.TypeApp,
		Name:         "search_api",
		RolloutType:  "stable",
		Owner:        "shared",
		ClusterType:  "common",
		InstanceName: provider.NewInstanceName("go1.dc", 8080),
	}
	key := newServiceKey("discovery", service)
	storageKey := key.storageKey()
	c.Assert(storageKey, Equals, "/discovery/app/search_api/stable/shared/common/go1.dc:8080")
}

func (s *TestStorageSuit) TestParseServiceKey_OK(c *C) {
	storageKey := "/discovery/app/search_api/stable/shared/common/go1.dc:8080"
	expected := serviceKey{
		Namespace: "discovery",
		Service: provider.Service{
			Type:         provider.TypeApp,
			Name:         "search_api",
			RolloutType:  "stable",
			Owner:        "shared",
			ClusterType:  "common",
			InstanceName: provider.NewInstanceName("go1.dc", 8080),
		},
	}

	key, err := parseServiceKey(storageKey)
	c.Assert(err, IsNil)
	c.Assert(key, DeepEquals, expected)
}

func (s *TestStorageSuit) TestParseServiceKey_ErrorIncorrectFormat(c *C) {
	storageKey := "/some/incorrect/key/"

	key, err := parseServiceKey(storageKey)
	c.Assert(err, NotNil)
	c.Assert(key, Equals, serviceKey{})
}

func (s *TestStorageSuit) TestParseServiceKey_ErrorMissingParts(c *C) {
	storageKey := "/discovery/app/search_api///common/go1.dc:8080"

	key, err := parseServiceKey(storageKey)
	c.Assert(err, NotNil)
	c.Assert(key, Equals, serviceKey{})
}

func (s *TestStorageSuit) TestParseServiceKey_ErrorIncorrectType(c *C) {
	storageKey := "/discovery/fooo/search_api/stable/shared/common/go1.dc:8080"

	key, err := parseServiceKey(storageKey)
	c.Assert(err, NotNil)
	c.Assert(key, Equals, serviceKey{})
}
