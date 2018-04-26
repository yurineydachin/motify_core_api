package registrator

import (
	"fmt"

	"godep.lzd.co/discovery/provider"
	. "gopkg.in/check.v1"
)

type TestDatasyncSuite struct{}

var _ = Suite(&TestDatasyncSuite{})

func (s *TestDatasyncSuite) BenchmarkStructAlloc(c *C) {
	c.ResetTimer()
	for n := 0; n < c.N; n++ {
		_ = ExportedEntity{Name: "foo", Endpoint: "host"}.kv()
	}
}

func (s *TestDatasyncSuite) BenchmarkRawCallAlloc(c *C) {
	rawCall := func(name, endpoint string) provider.KV {
		return provider.KV{
			RawKey: fmt.Sprintf("/%s/%s/%s", provider.NamespaceExportedEntities, name, endpoint),
		}
	}

	c.ResetTimer()
	for n := 0; n < c.N; n++ {
		_ = rawCall("foo", "host")
	}
}

func (s *TestDatasyncSuite) TestKV(c *C) {
	e := ExportedEntity{Name: "foo", Endpoint: "host:port"}
	kv := e.kv()

	c.Assert(kv, Equals, provider.KV{
		RawKey: fmt.Sprintf("/%s/%s/%s",
			provider.NamespaceExportedEntities,
			e.Name,
			e.Endpoint),
	})
}

func (s *TestDatasyncSuite) TestNewKVs(c *C) {
	endpoint := "host:port"
	names := exportedNames{"foo", "bar", "buz"}
	kvs := names.newKVs(endpoint)

	c.Assert(len(kvs), Equals, len(names))
	for i, kv := range kvs {
		expectedKey := fmt.Sprintf("/%s/%s/%s", provider.NamespaceExportedEntities, names[i], endpoint)
		c.Assert(kv.RawKey, Equals, expectedKey)
	}
}

func (s *TestDatasyncSuite) TestValidate(c *C) {
	names := exportedNames{"foo", "bar/slash"}
	err := names.validate()
	c.Assert(err, NotNil)
}

func (s *TestDatasyncSuite) TestNewExportedEntityFromKey(c *C) {
	rawKey := fmt.Sprintf("/%s/%s/%s", provider.NamespaceExportedEntities, "foo", "host:port")
	e, err := NewExportedEntityFromKey(rawKey)

	c.Assert(err, IsNil)
	c.Assert(e, Equals, ExportedEntity{Name: "foo", Endpoint: "host:port"})
}

func (s *TestDatasyncSuite) TestNewExportedEntityFromKey_Error(c *C) {
	invalidKeys := []string{
		fmt.Sprintf("/%s/some/wrong/name/host:port", provider.NamespaceExportedEntities), // entity name with "/"
		"/foo/bar/buz",          // tree parts, invalid namespace
		"/just_name/host:port",  // no namespace
		"/just_name/host:port/", // trailing slash
	}

	for _, key := range invalidKeys {
		_, err := NewExportedEntityFromKey(key)
		c.Assert(err, NotNil)
	}
}
