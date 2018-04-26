package registrator

import (
	"motify_core_api/godep_libs/discovery/provider"
)

// IRegistrationInfo is an interface to store service registration data
type IRegistrationInfo interface {
	// DiscoveryData returns data to be set when discovery is enabled
	DiscoveryData() []provider.KV
	// RegistrationData returns data to be set on service registration
	RegistrationData() []provider.KV
}
