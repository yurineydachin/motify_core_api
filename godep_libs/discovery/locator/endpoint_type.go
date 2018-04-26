package locator

import (
	"motify_core_api/godep_libs/discovery/provider"
)

// EndpointType is the type of requested endpoint.
// Is used to determine ServiceType and resource's endpoint field, containing
// network address info
type EndpointType uint8

// Available endpoint types
const (
	TypeUnknown EndpointType = iota
	TypeAppMain
	TypeAppAdditional
	TypeSystem // deprecated
	TypeSystemMain
	TypeSystemAdditional
	TypeExternal
)

// ServiceType converts EndpointType to provider.ServiceType
func (e EndpointType) ServiceType() provider.ServiceType {
	switch e {
	case TypeAppMain, TypeAppAdditional:
		return provider.TypeApp
	case TypeSystem, TypeSystemMain, TypeSystemAdditional:
		return provider.TypeSystem
	case TypeExternal:
		return provider.TypeExternal
	}
	return provider.TypeUnknown
}

// String returns EndpointType string representation for pretty debug print
func (e EndpointType) String() string {
	switch e {
	case TypeAppMain:
		return "app-main"
	case TypeAppAdditional:
		return "app-additional"
	case TypeSystem:
		return "system"
	case TypeSystemMain:
		return "system-main"
	case TypeSystemAdditional:
		return "system-additional"
	case TypeExternal:
		return "external"
	}
	return "unknown"
}
