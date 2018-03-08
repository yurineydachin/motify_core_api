package registrator

import (
	. "gopkg.in/check.v1"
)

type TestAdminInfoSuit struct{}

var _ = Suite(&TestAdminInfoSuit{})

func (t *TestAdminInfoSuit) TestNewAdminInfoFromString(c *C) {
	s := `{
        "admin_interface": "foo",
        "version": {
            "app_version": "v"
        }
    }`
	info, err := NewAdminInfoFromString(s)
	c.Assert(err, IsNil)
	c.Assert(info, Equals, adminInfo{
		AdminInterfaceURI: "foo",
		Version: VersionInfo{
			AppVersion: "v",
		},
	})
}

func (t *TestAdminInfoSuit) TestNewAdminInfoFromString_Error(c *C) {
	s := "foo"
	info, err := NewAdminInfoFromString(s)
	c.Assert(err, NotNil)
	c.Assert(info, Equals, adminInfo{})
}
