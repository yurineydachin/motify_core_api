package registrator

// dummyRegistrator is a registrator which does nothing
type dummyRegistrator struct{}

var _ IRegistrator = &dummyRegistrator{}

// Register does nothing
func (d *dummyRegistrator) Register() error { return nil }

// Unregister does nothing
func (d *dummyRegistrator) Unregister() {}

// EnableDiscovery does nothing
func (d *dummyRegistrator) EnableDiscovery() error { return nil }

// DisableDiscovery does nothing
func (d *dummyRegistrator) DisableDiscovery() {}
