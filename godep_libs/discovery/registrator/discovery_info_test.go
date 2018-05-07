package registrator

import (
	. "gopkg.in/check.v1"
)

type TestDiscoverySuite struct{}

var _ = Suite(&TestDiscoverySuite{})

func (s *TestDiscoverySuite) TestHTTPValue(c *C) {
	c.Logf("http value host should contain '-pub' suffix")

	info := discoveryInfo{
		Host:     "go1.iddc",
		HTTPPort: 8080,
	}
	c.Assert(info.httpValue(), Equals, "http://go1-pub.iddc:8080")
}

func (s *TestDiscoverySuite) TestGRPCValue(c *C) {
	c.Logf("grpc value host should contain '-pub' suffix")

	info := discoveryInfo{
		Host:     "go1.iddc",
		GRPCPort: 8888,
	}
	c.Assert(info.grpcValue(), Equals, "go1-pub.iddc:8888")
}

func (s *TestDiscoverySuite) TestHostValue(c *C) {
	c.Logf("host value should be modified with '-pub' suffix for some DCs")
	DCs := []string{".iddc", ".sgdc", ".hkdc"}

	for _, dc := range DCs {
		info := discoveryInfo{
			Host: "go1" + dc,
		}
		expected := "go1-pub" + dc
		c.Assert(info.hostValue(), Equals, expected)
	}
}

func (s *TestDiscoverySuite) TestHostValue_Duplicate(c *C) {
	c.Logf("'-pub' suffix should not be duplicated")
	info := discoveryInfo{
		Host: "go1-pub.iddc",
	}
	c.Assert(info.hostValue(), Equals, "go1-pub.iddc")
}

func (s *TestDiscoverySuite) TestDiscoveryValue_Preset(c *C) {
	c.Logf("preset discovery value should be returned")
	v := DiscoveryValue{
		EndpointMain: "go1.dc:8080",
		Login:        "user",
		Password:     "password",
	}
	info := discoveryInfo{
		PresetValue: &v,
	}
	c.Assert(info.discoveryValue(), Equals, v)
}

func (s *TestDiscoverySuite) TestDiscoveryValue_HTTP(c *C) {
	info := discoveryInfo{
		Host:     "go1.dc",
		HTTPPort: 8080,
	}
	c.Assert(info.discoveryValue(), Equals, DiscoveryValue{EndpointMain: "http://go1.dc:8080"})
}

func (s *TestDiscoverySuite) TestDiscoveryValue_gRPC(c *C) {
	info := discoveryInfo{
		Host:     "go1.dc",
		GRPCPort: 8080,
	}
	c.Assert(info.discoveryValue(), Equals, DiscoveryValue{EndpointMain: "go1.dc:8080"})
}

func (s *TestDiscoverySuite) TestDiscoveryValue_gRPCWithHTTP(c *C) {
	info := discoveryInfo{
		Host:     "go1.dc",
		HTTPPort: 80,
		GRPCPort: 8888,
	}
	c.Assert(info.discoveryValue(), Equals, DiscoveryValue{EndpointMain: "http://go1.dc:80", EndpointAdditional: "go1.dc:8888"})
}

func (s *TestDiscoverySuite) TestValidate_InvalidPresetValue(c *C) {
	info := discoveryInfo{
		PresetValue: &DiscoveryValue{},
	}
	c.Assert(info.validate(), NotNil)
}

func (s *TestDiscoverySuite) TestValidate_NoPorts(c *C) {
	info := discoveryInfo{
		Host: "go1.dc",
	}
	c.Assert(info.validate(), NotNil)

	c.Logf("It should be possible to set gRPC port only")
	info.GRPCPort = 80
	c.Assert(info.validate(), IsNil)

	info.HTTPPort = 80
	c.Assert(info.validate(), IsNil)
}

func (s *TestDiscoverySuite) TestValidate_gRPC(c *C) {
	c.Logf("It should be possible to set gRPC port only")
	info := discoveryInfo{
		Host:     "go1.dc",
		GRPCPort: 8080,
	}
	c.Assert(info.validate(), IsNil)
}

func (s *TestDiscoverySuite) TestValidate_HTTP(c *C) {
	c.Logf("It should be possible to set HTTP port only")
	info := discoveryInfo{
		Host:     "go1.dc",
		HTTPPort: 80,
	}
	c.Assert(info.validate(), IsNil)
}

func (s *TestDiscoverySuite) TestValue(c *C) {
	c.Logf("The discovery value should be successfully marshaled and unmarshaled")
	info := discoveryInfo{
		Host:     "go1.dc",
		HTTPPort: 80,
		GRPCPort: 8888,
	}
	value := info.value()
	c.Assert(value, Not(Equals), "")

	expected := DiscoveryValue{
		EndpointMain:       "http://go1.dc:80",
		EndpointAdditional: "go1.dc:8888",
	}
	discoveryValue, err := NewDiscoveryValueFromString(value)
	c.Assert(err, IsNil)
	c.Assert(discoveryValue, DeepEquals, expected)
}

func (s *TestDiscoverySuite) TestValue_Preset(c *C) {
	c.Logf("The preset discovery value should be successfully marshaled and unmarshaled")
	expected := DiscoveryValue{
		EndpointMain:       "http://go1.dc:80",
		EndpointAdditional: "go1.dc:8888",
		Login:              "login",
		Password:           "password",
	}
	info := discoveryInfo{
		PresetValue: &expected,
	}
	value := info.value()
	c.Assert(value, Not(Equals), "")

	discoveryValue, err := NewDiscoveryValueFromString(value)
	c.Assert(err, IsNil)
	c.Assert(discoveryValue, DeepEquals, expected)
}

func (s *TestDiscoverySuite) TestMarshal(c *C) {
	c.Logf("The value should marshal in defined structure")
	v := DiscoveryValue{
		EndpointMain:       "http://go1.dc:80",
		EndpointAdditional: "go1.dc:8888",
		Login:              "login",
		Password:           "password",
	}
	info := discoveryInfo{
		PresetValue: &v,
	}
	value := info.value()
	expected := `{"endpoint_main":"http://go1.dc:80","endpoint_additional":"go1.dc:8888","login":"login","pass":"password"}`
	c.Assert(value, Equals, expected)
}
