package provider

import (
	"fmt"
)

// InstanceName is a string, identifying service instance
type InstanceName string

// String returns string representation of InstanceName
func (i InstanceName) String() string { return string(i) }

// NewInstanceName returns new InstanceName from host and port.
// Other compatible constructors could be implemented in future, if needed
func NewInstanceName(host string, port int) InstanceName {
	if host == "" || port <= 0 {
		// extra check, we don't want to permit invalid instance names
		return ""
	}
	return InstanceName(fmt.Sprintf("%s:%d", host, port))
}

// NewInstanceNameFromString returns new InstanceName from string.
func NewInstanceNameFromString(s string) InstanceName {
	return InstanceName(s)
}
