package provider

import (
	. "gopkg.in/check.v1"
)

type TestServiceTypeSuit struct{}

var _ = Suite(&TestServiceTypeSuit{})

func (s *TestServiceTypeSuit) TestString(c *C) {
	c.Logf("Correct strings for every type should be returned")
	c.Assert(TypeApp.String(), Equals, "app")
	c.Assert(TypeSystem.String(), Equals, "system")
	c.Assert(TypeExternal.String(), Equals, "external")
}

func (s *TestServiceTypeSuit) TestString_InvalidType(c *C) {
	c.Logf("Empty strings should be returned for every unknown service type")
	c.Assert(TypeUnknown.String(), Equals, "")
	c.Assert(ServiceType(123).String(), Equals, "")
}

func (s *TestServiceTypeSuit) TestParseServiceType(c *C) {
	c.Logf("String values should be parsed into correct service types")
	v, err := ParseServiceType("app")
	c.Assert(v, Equals, TypeApp)
	c.Assert(err, IsNil)

	v, err = ParseServiceType("system")
	c.Assert(v, Equals, TypeSystem)
	c.Assert(err, IsNil)

	v, err = ParseServiceType("external")
	c.Assert(v, Equals, TypeExternal)
	c.Assert(err, IsNil)

	v, err = ParseServiceType("")
	c.Assert(v, Equals, TypeUnknown)
	c.Assert(err, NotNil)
}
