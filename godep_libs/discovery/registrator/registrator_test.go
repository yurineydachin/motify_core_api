package registrator

import (
	"context"
	"testing"
	"time"

	"motify_core_api/godep_libs/discovery/provider"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type TestRegistratorSuit struct{}

var _ = Suite(&TestRegistratorSuit{})

func newAppRegistrationInfo() *appRegistrationInfo {
	info, err := NewAppRegistrationInfo(newAppRegistrationParams())
	if err != nil {
		panic(err)
	}
	return info
}

func (s *TestRegistratorSuit) TestRegister(c *C) {
	values := make(map[provider.KV]bool)
	valuesSaved := make(chan bool)
	p := &provider.Mock{
		RegisterValuesCallback: func(ctx context.Context, kvs ...provider.KV) error {
			for _, kv := range kvs {
				values[kv] = true
			}
			valuesSaved <- true
			<-ctx.Done()

			return nil
		},
	}

	info := newAppRegistrationInfo()
	r := New(p, info, nil)
	c.Assert(r, NotNil)

	c.Logf("Registration values should be set")
	err := r.Register()
	c.Assert(err, IsNil)

	select {
	case <-time.After(time.Second):
		c.Fatalf("Register timeout")
	case <-valuesSaved:
	}

	for _, kv := range info.RegistrationData() {
		c.Assert(values[kv], Equals, true)
	}
}

func (s *TestRegistratorSuit) TestUnregister(c *C) {
	canceled := make(chan bool)
	p := &provider.Mock{
		RegisterValuesCallback: func(ctx context.Context, kvs ...provider.KV) error {
			<-ctx.Done()
			canceled <- true
			return nil
		},
	}

	r := New(p, newAppRegistrationInfo(), nil)
	c.Assert(r, NotNil)

	err := r.Register()
	c.Assert(err, IsNil)

	c.Logf("Registration should be canceled")
	go r.Unregister()

	select {
	case <-time.After(time.Second):
		c.Fatalf("Unregister timeout")
	case <-canceled:
	}
}

func (s *TestRegistratorSuit) TestEnableDiscovery(c *C) {
	values := make(map[provider.KV]bool)
	valuesSaved := make(chan bool)
	p := &provider.Mock{
		RegisterValuesCallback: func(ctx context.Context, kvs ...provider.KV) error {
			for _, kv := range kvs {
				values[kv] = true
			}
			valuesSaved <- true
			<-ctx.Done()

			return nil
		},
	}

	info := newAppRegistrationInfo()
	r := New(p, info, nil)
	c.Assert(r, NotNil)

	c.Logf("Discovery values should be set")
	err := r.EnableDiscovery()
	c.Assert(err, IsNil)

	select {
	case <-time.After(time.Second):
		c.Fatalf("Register timeout")
	case <-valuesSaved:
	}

	for _, kv := range info.DiscoveryData() {
		c.Assert(values[kv], Equals, true)
	}
}

func (s *TestRegistratorSuit) TestDisableDiscovery(c *C) {
	canceled := make(chan bool)
	p := &provider.Mock{
		RegisterValuesCallback: func(ctx context.Context, kvs ...provider.KV) error {
			<-ctx.Done()
			canceled <- true
			return nil
		},
	}

	r := New(p, newAppRegistrationInfo(), nil)
	c.Assert(r, NotNil)

	err := r.EnableDiscovery()
	c.Assert(err, IsNil)

	c.Logf("Registration should be canceled")
	go r.DisableDiscovery()

	select {
	case <-time.After(time.Second):
		c.Fatalf("DisableDiscovery timeout")
	case <-canceled:
	}
}
