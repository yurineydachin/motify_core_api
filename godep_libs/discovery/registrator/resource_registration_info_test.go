package registrator

import (
	"motify_core_api/godep_libs/discovery/provider"
	. "gopkg.in/check.v1"
)

type TestResourceRegistrationInfoSuit struct{}

var _ = Suite(&TestResourceRegistrationInfoSuit{})

func newSystemRegistrationParams() ResourceRegistrationParams {
	return ResourceRegistrationParams{
		Service: provider.Service{
			Name:         "mysql",
			Type:         provider.TypeSystem,
			RolloutType:  "stable",
			Owner:        "bundle_api",
			ClusterType:  "master",
			InstanceName: provider.NewInstanceName("go1.dc", 3306),
		},
		DiscoveryValue: DiscoveryValue{
			EndpointMain: "go1.dc:3306",
			Login:        "admin",
			Password:     "1234",
		},
	}
}

func (s *TestResourceRegistrationInfoSuit) TestNew(c *C) {
	c.Logf("Typical system resource data should init well")

	info, err := NewResourceRegistrationInfo(newSystemRegistrationParams())
	c.Assert(err, IsNil)
	c.Assert(info, NotNil)
}

func (s *TestResourceRegistrationInfoSuit) TestNew_InvalidService(c *C) {
	c.Logf("Service field should be validated")

	params := ResourceRegistrationParams{
		DiscoveryValue: DiscoveryValue{
			EndpointMain: "go1.dc:3306",
		},
	}
	info, err := NewResourceRegistrationInfo(params)
	c.Assert(info, IsNil)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "invalid Service info.*")
}

func (s *TestResourceRegistrationInfoSuit) TestNew_InvalidDiscoveryValue(c *C) {
	c.Logf("Service field should be validated")

	params := newSystemRegistrationParams()
	params.DiscoveryValue = DiscoveryValue{}

	info, err := NewResourceRegistrationInfo(params)
	c.Assert(info, IsNil)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "invalid discovery value.*")
}

func (s *TestResourceRegistrationInfoSuit) TestNew_WithAdminInfo(c *C) {
	params := newSystemRegistrationParams()
	params.Version = VersionInfo{AppVersion: "version"}
	params.Host = "host"
	params.AdminPort = 666

	info, err := NewResourceRegistrationInfo(params)
	c.Assert(info, NotNil)
	c.Assert(err, IsNil)
}

func (s *TestResourceRegistrationInfoSuit) TestNew_WithMonitoringInfo_Invalid(c *C) {
	c.Logf("Monitoring info should be validated if given")

	params := newSystemRegistrationParams()
	params.MonitoringPort = 666

	info, err := NewResourceRegistrationInfo(params)
	c.Assert(info, IsNil)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "invalid monitoring info.*")
}

func (s *TestResourceRegistrationInfoSuit) TestDiscoveryData(c *C) {
	info, err := NewResourceRegistrationInfo(newSystemRegistrationParams())
	c.Assert(err, IsNil)
	c.Assert(info, NotNil)

	kvs := info.DiscoveryData()

	c.Assert(len(kvs), Equals, 1)
	c.Assert(kvs[0].Namespace, Equals, "discovery")
	c.Assert(kvs[0].Service, Equals, info.service)
	c.Assert(kvs[0].Value, Equals, info.discovery.value())
}

func (s *TestResourceRegistrationInfoSuit) TestRegistrationData(c *C) {
	info, err := NewResourceRegistrationInfo(newSystemRegistrationParams())
	c.Assert(err, IsNil)
	c.Assert(info, NotNil)

	kvs := info.RegistrationData()

	c.Assert(len(kvs), Equals, 0)
}

func (s *TestResourceRegistrationInfoSuit) TestRegistrationData_WithAdminInfo(c *C) {
	params := newSystemRegistrationParams()
	params.Version = VersionInfo{AppVersion: "version"}
	params.Host = "host"
	params.AdminPort = 666

	info, err := NewResourceRegistrationInfo(params)
	c.Assert(err, IsNil)
	c.Assert(info, NotNil)
	admin := adminInfo{
		Version:           params.Version,
		AdminInterfaceURI: "http://host:666",
	}
	c.Assert(info.admin, Equals, admin)

	kvs := info.RegistrationData()
	c.Assert(len(kvs), Equals, 1)
	c.Assert(kvs[0], Equals, provider.KV{
		Namespace: provider.NamespaceAdmin,
		Service:   info.service,
		Value:     admin.value(),
	})
}

func (s *TestResourceRegistrationInfoSuit) TestRegistrationData_WithMonitoringInfo(c *C) {
	params := newSystemRegistrationParams()
	params.MonitoringPort = 666
	params.Host = "host"
	params.Venture = "vn"
	params.Environment = "test"

	info, err := NewResourceRegistrationInfo(params)
	c.Assert(err, IsNil)
	c.Assert(info, NotNil)
	monitoring := monitoringInfo{
		Port:        params.MonitoringPort,
		Host:        params.Host,
		Venture:     params.Venture,
		Environment: params.Environment,
		ServiceName: params.Service.Name,
		RolloutType: params.Service.RolloutType,
	}
	c.Assert(info.monitoring, Equals, monitoring)

	kvs := info.RegistrationData()
	c.Assert(len(kvs), Equals, 1)
	c.Assert(kvs[0], Equals, provider.KV{
		Namespace: provider.NamespaceMetrics,
		Service:   info.service,
		Value:     monitoring.value(),
	})
}
