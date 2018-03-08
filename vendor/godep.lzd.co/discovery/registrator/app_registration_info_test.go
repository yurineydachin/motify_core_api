package registrator

import (
	. "gopkg.in/check.v1"
)

type TestAppRegistrationInfoSuit struct{}

var _ = Suite(&TestAppRegistrationInfoSuit{})

func newAppRegistrationParams() AppRegistrationParams {
	return AppRegistrationParams{
		ServiceName: "goblin",
		RolloutType: "stable",
		Host:        "go1.dc",
		HTTPPort:    8080,

		MonitoringPort: 8081,
		Venture:        "vn",
		Environment:    "dev",
		AdminPort:      8081,
		Version: VersionInfo{
			AppVersion: "N/A",
		},
	}
}

func (s *TestAppRegistrationInfoSuit) TestNew_AppTypical(c *C) {
	c.Logf("Typical app service data should init well")

	info, err := NewAppRegistrationInfo(newAppRegistrationParams())
	c.Assert(err, IsNil)
	c.Assert(info, NotNil)
}

func (s *TestAppRegistrationInfoSuit) TestNew_MonitoringInfoLack(c *C) {
	c.Logf("Absence of monitoring data should lead to an error by default")

	params := AppRegistrationParams{
		ServiceName: "goblin",
		RolloutType: "stable",
		Host:        "go1.dc",
		HTTPPort:    8080,

		AdminPort: 8081,
		Version: VersionInfo{
			AppVersion: "N/A",
		},
	}
	info, err := NewAppRegistrationInfo(params)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "invalid monitoring info.*")
	c.Assert(info, IsNil)
}

func (s *TestAppRegistrationInfoSuit) TestNew_VersionInfoLack(c *C) {
	c.Logf("Absence of version data should lead to an error by default")

	params := AppRegistrationParams{
		ServiceName: "goblin",
		RolloutType: "stable",
		Host:        "go1.dc",
		HTTPPort:    8080,
	}
	info, err := NewAppRegistrationInfo(params)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "invalid version info.*")
	c.Assert(info, IsNil)
}

func (s *TestAppRegistrationInfoSuit) TestNew_ExportedEntities_Error(c *C) {
	c.Logf("Names containing key separator '/' should be invalid")

	params := newAppRegistrationParams()
	params.ExportedEntities = []string{"name/with/slash"}
	info, err := NewAppRegistrationInfo(params)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "invalid ExportedEntity.*")
	c.Assert(info, IsNil)
}

func (s *TestAppRegistrationInfoSuit) TestDiscoveryData(c *C) {
	info, err := NewAppRegistrationInfo(newAppRegistrationParams())
	c.Assert(err, IsNil)
	c.Assert(info, NotNil)

	kvs := info.DiscoveryData()

	c.Assert(len(kvs), Equals, 1)
	c.Assert(kvs[0].Namespace, Equals, "discovery")
	c.Assert(kvs[0].Service, Equals, info.service)
	c.Assert(kvs[0].Value, Equals, info.discovery.value())
}

func (s *TestAppRegistrationInfoSuit) TestDiscoveryData_ExportedEntities_NoGRPC(c *C) {
	params := newAppRegistrationParams()
	params.ExportedEntities = []string{"foo"}
	info, err := NewAppRegistrationInfo(params)
	c.Assert(err, IsNil)
	c.Assert(info, NotNil)

	kvs := info.DiscoveryData()

	c.Assert(len(kvs), Equals, 1)
	c.Assert(kvs[0].Namespace, Equals, "discovery")
	c.Assert(kvs[0].Service, Equals, info.service)
	c.Assert(kvs[0].Value, Equals, info.discovery.value())
}

func (s *TestAppRegistrationInfoSuit) TestDiscoveryData_ExportedEntities_GRPC(c *C) {
	params := newAppRegistrationParams()
	params.ExportedEntities = []string{"foo"}
	params.GRPCPort = 666
	info, err := NewAppRegistrationInfo(params)
	c.Assert(err, IsNil)
	c.Assert(info, NotNil)

	kvs := info.DiscoveryData()

	c.Assert(len(kvs), Equals, 2)
	expected := info.exportedEntities.newKVs(info.discovery.grpcValue())
	c.Assert(kvs[0], Equals, expected[0])
}

func (s *TestAppRegistrationInfoSuit) TestRegistrationData(c *C) {
	info, err := NewAppRegistrationInfo(newAppRegistrationParams())
	c.Assert(err, IsNil)
	c.Assert(info, NotNil)

	kvs := info.RegistrationData()

	c.Assert(len(kvs), Equals, 2)
	c.Assert(kvs[0].Service, Equals, info.service)
	c.Assert(kvs[0].Value, Equals, info.admin.value())
	c.Assert(kvs[1].Service, Equals, info.service)
	c.Assert(kvs[1].Value, Equals, info.monitoring.value())
}
