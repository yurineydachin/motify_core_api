package provider

import (
	. "gopkg.in/check.v1"
)

type TestInstanceNameSuit struct{}

var _ = Suite(&TestInstanceNameSuit{})

func (s *TestInstanceNameSuit) TestNewInstanceName(c *C) {
	c.Assert(NewInstanceName("go1.dc", 8080), Equals, InstanceName("go1.dc:8080"))
	c.Assert(NewInstanceName("", 8080), Equals, InstanceName(""))
	c.Assert(NewInstanceName("go1.dc", 0), Equals, InstanceName(""))
}
