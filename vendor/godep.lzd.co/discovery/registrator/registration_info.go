package registrator

import (
	"godep.lzd.co/discovery/provider"
)

// IRegistrationInfo is an interface to store service registration data
type IRegistrationInfo interface {
	// DiscoveryData returns data to be set when discovery is enabled
	DiscoveryData() []provider.KV
	// RegistrationData returns data to be set on service registration
	RegistrationData() []provider.KV
}
