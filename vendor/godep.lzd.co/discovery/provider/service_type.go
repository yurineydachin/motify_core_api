package provider

import (
	"fmt"
)

// ServiceType is a type, identifying service type
type ServiceType uint8

// Service types enum
const (
	TypeUnknown ServiceType = iota
	TypeApp
	TypeSystem
	TypeExternal
)

// ParseServiceType returns ServiceType from string
func ParseServiceType(s string) (ServiceType, error) {
	switch s {
	case "app":
		return TypeApp, nil
	case "system":
		return TypeSystem, nil
	case "external":
		return TypeExternal, nil
	}
	return TypeUnknown, fmt.Errorf("invalid service type: '%s'", s)
}

// String returns ServiceType as string
func (t ServiceType) String() string {
	switch t {
	case TypeApp:
		return "app"
	case TypeSystem:
		return "system"
	case TypeExternal:
		return "external"
	}

	return ""
}
