package etcdV3

import (
	"godep.lzd.co/discovery/provider"
	. "gopkg.in/check.v1"
)

type TestValidatorsSuit struct{}

var _ = Suite(&TestValidatorsSuit{})

func newValidService() provider.Service {
	return provider.Service{
		Name:         "test_service",
		Type:         provider.TypeApp,
		Owner:        provider.DefaultOwner,
		RolloutType:  provider.RolloutTypeStable,
		ClusterType:  provider.DefaultClusterType,
		InstanceName: provider.NewInstanceNameFromString("host:port"),
	}
}

func (s *TestValidatorsSuit) TestValidateKV_RawKey(c *C) {
	c.Logf("Nothing should be checked if RawKey is given")
	kv := provider.KV{
		RawKey: "blabla",
	}

	c.Assert(validateKV(kv), IsNil)
}

func (s *TestValidatorsSuit) TestValidateKV_EmptyValue(c *C) {
	kv := provider.KV{
		Service:   newValidService(),
		Namespace: "foo",
	}

	err := validateKV(kv)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "empty Value")
}

func (s *TestValidatorsSuit) TestValidateKV_EmptyNamespace(c *C) {
	kv := provider.KV{
		Service: newValidService(),
		Value:   "foo",
	}

	err := validateKV(kv)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "empty Namespace")
}

func (s *TestValidatorsSuit) TestValidateKV_InvalidService(c *C) {
	kv := provider.KV{
		Service:   newValidService(),
		Value:     "foo",
		Namespace: "foo",
	}
	kv.Service.Name = ""

	err := validateKV(kv)
	c.Assert(err, NotNil)
}

func (s *TestValidatorsSuit) TestValidateKV_OK(c *C) {
	kv := provider.KV{
		Service:   newValidService(),
		Namespace: "foo",
		Value:     "value",
	}

	err := validateKV(kv)
	c.Assert(err, IsNil)
}

func (s *TestValidatorsSuit) TestValidateRegistrationData_OK(c *C) {
	kv := provider.KV{
		Service:   newValidService(),
		Namespace: "foo",
		Value:     "value",
	}

	err := validateRegistrationData(kv)
	c.Assert(err, IsNil)
}

func (s *TestValidatorsSuit) TestValidateRegistrationData_NoData(c *C) {
	err := validateRegistrationData([]provider.KV{}...)
	c.Assert(err, NotNil)
}

func (s *TestValidatorsSuit) TestValidateRegistrationData_EmptyKV(c *C) {
	err := validateRegistrationData([]provider.KV{provider.KV{}}...)
	c.Assert(err, NotNil)
}
