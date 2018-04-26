package registrator

import (
	"fmt"

	"godep.lzd.co/discovery/provider"
)

// AppRegistrationParams contains application discovery and registration data
type AppRegistrationParams struct {
	// App service discovery info
	ServiceName string
	RolloutType string
	Host        string
	HTTPPort    int
	GRPCPort    int

	// Admin info
	AdminPort int
	Version   VersionInfo

	// Monitoring info
	MonitoringPort int
	Environment    string
	Venture        string

	// DataSyncAPI info - see https://confluence.lzd.co/display/GO/DataSync+API
	ExportedEntities []string
}

// appRegistrationInfo implements IRegistrationInfo interface for application params
type appRegistrationInfo struct {
	service          provider.Service
	discovery        discoveryInfo
	admin            adminInfo
	monitoring       monitoringInfo
	exportedEntities exportedNames
}

var _ IRegistrationInfo = &appRegistrationInfo{}

// NewAppRegistrationInfo validates the registration params returns appRegistrationInfo instance
func NewAppRegistrationInfo(p AppRegistrationParams) (*appRegistrationInfo, error) {
	info := &appRegistrationInfo{
		discovery: discoveryInfo{
			Host:     p.Host,
			HTTPPort: p.HTTPPort,
			GRPCPort: p.GRPCPort,
		},
		admin: adminInfo{
			AdminInterfaceURI: fmt.Sprintf("http://%s:%d", p.Host, p.AdminPort),
			Version:           p.Version,
		},
		monitoring: monitoringInfo{
			Host:        p.Host,
			Port:        p.MonitoringPort,
			Environment: p.Environment,
			Venture:     p.Venture,
			ServiceName: p.ServiceName,
			RolloutType: p.RolloutType,
		},
		exportedEntities: exportedNames(p.ExportedEntities),
	}
	info.service = provider.Service{
		Type:         provider.TypeApp,
		Name:         p.ServiceName,
		RolloutType:  p.RolloutType,
		Owner:        provider.DefaultOwner,
		ClusterType:  provider.DefaultClusterType,
		InstanceName: provider.NewInstanceName(p.Host, info.discovery.mainPort()),
	}
	if err := info.service.Validate(); err != nil {
		return nil, err
	}
	if err := info.discovery.validate(); err != nil {
		return nil, err
	}
	if err := info.admin.validate(); err != nil {
		return nil, err
	}
	if err := info.monitoring.validate(); err != nil {
		return nil, err
	}
	if err := info.exportedEntities.validate(); err != nil {
		return nil, err
	}

	return info, nil
}

// DiscoveryData returns data to be set when discovery is enabled
func (i *appRegistrationInfo) DiscoveryData() []provider.KV {
	kvs := i.exportedEntities.newKVs(i.discovery.grpcValue())
	if kvs == nil {
		kvs = make([]provider.KV, 0, 1)
	}

	kvs = append(kvs, provider.KV{
		Namespace: provider.NamespaceDiscovery,
		Service:   i.service,
		Value:     i.discovery.value(),
	})
	return kvs
}

// RegistrationData returns data to be set on application registration
func (i *appRegistrationInfo) RegistrationData() []provider.KV {
	return []provider.KV{
		provider.KV{
			Namespace: provider.NamespaceAdmin,
			Service:   i.service,
			Value:     i.admin.value(),
		},
		provider.KV{
			Namespace: provider.NamespaceMetrics,
			Service:   i.service,
			Value:     i.monitoring.value(),
		},
	}
}
