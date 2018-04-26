package registrator

import (
	"fmt"

	"motify_core_api/godep_libs/discovery/provider"
)

// ResourceRegistrationParams contains system and external resources discovery info
type ResourceRegistrationParams struct {
	// Service is used for complete definition of all service identification fields
	Service provider.Service
	// DiscoveryValue is used for any custom discovery value to be set
	DiscoveryValue DiscoveryValue

	// Host is used as common data for monitoring and admin info
	Host string

	// Admin info
	AdminPort int
	Version   VersionInfo

	// Monitoring info
	MonitoringPort int
	Environment    string
	Venture        string
}

type resourceRegistrationInfo struct {
	service    provider.Service
	discovery  discoveryInfo
	admin      adminInfo
	monitoring monitoringInfo
}

// NewResourceRegistrationInfo validates the registration params returns resourceRegistrationInfo instance
func NewResourceRegistrationInfo(p ResourceRegistrationParams) (*resourceRegistrationInfo, error) {
	info := &resourceRegistrationInfo{
		service: p.Service,
		discovery: discoveryInfo{
			PresetValue: &p.DiscoveryValue,
		},
		admin: adminInfo{
			Version: p.Version,
		},
	}

	// admin info and monitoring info are not mandatory for resources
	if p.AdminPort != 0 {
		info.admin.AdminInterfaceURI = fmt.Sprintf("http://%s:%d", p.Host, p.AdminPort)
	}
	if p.MonitoringPort != 0 {
		info.monitoring = monitoringInfo{
			Host:        p.Host,
			Port:        p.MonitoringPort,
			Environment: p.Environment,
			Venture:     p.Venture,
			ServiceName: p.Service.Name,
			RolloutType: p.Service.RolloutType,
		}
		// validate monitoring info, if it presents
		if err := info.monitoring.validate(); err != nil {
			return nil, err
		}
	}

	if err := info.service.Validate(); err != nil {
		return nil, err
	}
	if err := info.discovery.validate(); err != nil {
		return nil, err
	}

	return info, nil
}

// DiscoveryKV returns data to be set when discovery is enabled
func (i *resourceRegistrationInfo) DiscoveryData() []provider.KV {
	return []provider.KV{
		provider.KV{
			Namespace: provider.NamespaceDiscovery,
			Service:   i.service,
			Value:     i.discovery.value(),
		},
	}
}

// RegisterKVs returns nothing, because resource registration is not used now.
func (i *resourceRegistrationInfo) RegistrationData() []provider.KV {
	kvs := []provider.KV{}
	if !i.admin.isEmpty() {
		kvs = append(kvs, provider.KV{
			Namespace: provider.NamespaceAdmin,
			Service:   i.service,
			Value:     i.admin.value(),
		})
	}
	if !i.monitoring.isEmpty() {
		kvs = append(kvs, provider.KV{
			Namespace: provider.NamespaceMetrics,
			Service:   i.service,
			Value:     i.monitoring.value(),
		})
	}
	return kvs
}
