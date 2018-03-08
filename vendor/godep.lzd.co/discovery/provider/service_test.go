package provider

import (
	. "gopkg.in/check.v1"
)

type TestServiceSuit struct{}

var _ = Suite(&TestServiceSuit{})

func (s *TestServiceSuit) newValidService() Service {
	return Service{
		Name:         "test_service",
		Type:         TypeApp,
		Owner:        DefaultOwner,
		RolloutType:  RolloutTypeStable,
		ClusterType:  DefaultClusterType,
		InstanceName: NewInstanceNameFromString("host:port"),
	}
}

func (s *TestServiceSuit) generateName(r rune, count int) string {
	runes := make([]rune, 0, count)
	for i := 0; i < count; i++ {
		runes = append(runes, r)
	}
	return string(runes)
}

func (s *TestServiceSuit) TestValidate_Valid(c *C) {
	service := s.newValidService()
	err := service.Validate()
	c.Assert(err, IsNil)
}

func (s *TestServiceSuit) TestValidate_EmptyName(c *C) {
	service := s.newValidService()
	service.Name = ""

	err := service.Validate()
	c.Assert(err, NotNil)
}

func (s *TestServiceSuit) TestValidate_LongName(c *C) {
	service := s.newValidService()
	longNames := []string{
		s.generateName('a', maxServiceNameLen+1),
		s.generateName('å', maxServiceNameLen+1),
	}

	for _, v := range longNames {
		service.Name = v
		c.Logf("name %q (len=%d, char_count=%d) should be too long", v, len(v), len([]rune(v)))
		c.Check(service.Validate(), NotNil)
	}
}

func (s *TestServiceSuit) TestValidate_ForbiddenName(c *C) {
	service := s.newValidService()
	service.Name = "content/live"

	err := service.Validate()
	c.Assert(err, NotNil)
}

func (s *TestServiceSuit) TestValidate_ValidName(c *C) {
	service := s.newValidService()
	valid := []string{
		s.generateName('a', maxServiceNameLen/2),
		s.generateName('å', maxServiceNameLen/2),
	}

	for _, v := range valid {
		service.Name = v
		c.Logf("name %q (len=%d, char_count=%d) should be valid", v, len(v), len([]rune(v)))
		c.Check(service.Validate(), IsNil)
	}
}

func (s *TestServiceSuit) TestValidate_EmptyInstanceName(c *C) {
	service := s.newValidService()
	service.InstanceName = ""

	err := service.Validate()
	c.Assert(err, NotNil)
}

func (s *TestServiceSuit) TestValidate_EmptyOwner(c *C) {
	service := s.newValidService()
	service.Owner = ""

	err := service.Validate()
	c.Assert(err, NotNil)
}

func (s *TestServiceSuit) TestValidate_EmptyClusterType(c *C) {
	service := s.newValidService()
	service.ClusterType = ""

	err := service.Validate()
	c.Assert(err, NotNil)
}

func (s *TestServiceSuit) TestValidate_EmptyType(c *C) {
	service := s.newValidService()
	service.Type = TypeUnknown

	err := service.Validate()
	c.Assert(err, NotNil)
}

func (s *TestServiceSuit) TestValidate_InvalidRolloutType(c *C) {
	service := s.newValidService()
	invalidRolloutTypes := []string{"", "rollout_stable", "unstable_1", "unstable100", "un-stable1"}

	for _, v := range invalidRolloutTypes {
		service.RolloutType = v
		c.Logf("rollout type %s should be invalid", v)
		c.Check(service.Validate(), NotNil)
	}
}
