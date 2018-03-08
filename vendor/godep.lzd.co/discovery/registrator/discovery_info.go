package registrator

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	httpScheme = "http"
)

// discoveryInfo is a struct describing service discovery value
type discoveryInfo struct {
	PresetValue *DiscoveryValue

	// host is a common host all endpoints listen on
	Host     string
	HTTPPort int
	GRPCPort int
}

// DiscoveryValue is a common value struct for registering in provider
type DiscoveryValue struct {
	EndpointMain       string `json:"endpoint_main"`
	EndpointAdditional string `json:"endpoint_additional,omitempty"`
	Login              string `json:"login,omitempty"`
	Password           string `json:"pass,omitempty"`
}

// NewDiscoveryValueFromString returns new DiscoveryValue from string value
func NewDiscoveryValueFromString(s string) (DiscoveryValue, error) {
	v := DiscoveryValue{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return DiscoveryValue{}, err
	}

	return v, nil
}

func (d *DiscoveryValue) validate() error {
	if d.EndpointMain == "" {
		return fmt.Errorf("invalid discovery value: EndpointMain is empty")
	}
	return nil
}

// hostValue returns modified host value for registering in the discovery
// dirty hack, GO-4096
// TODO: remove when Infra is ready
// https://confluence.lazada.com/display/INFRA/Standards+and+agreements
func (i *discoveryInfo) hostValue() string {
	if strings.Contains(i.Host, "-pub") {
		// we should be ready for Infra change - don't duplicate "-pub" if found.
		return i.Host
	}

	value := i.Host
	if idx := strings.LastIndex(value, ".iddc"); idx != -1 {
		value = value[:idx] + "-pub" + value[idx:]
	} else if idx := strings.LastIndex(value, ".sgdc"); idx != -1 {
		value = value[:idx] + "-pub" + value[idx:]
	} else if idx := strings.LastIndex(value, ".hkdc"); idx != -1 {
		value = value[:idx] + "-pub" + value[idx:]
	}
	return value
}

// httpValue returns HTTP discovery value
func (i *discoveryInfo) httpValue() string {
	if i.HTTPPort != 0 {
		return fmt.Sprintf("%s://%s:%d", httpScheme, i.hostValue(), i.HTTPPort)
	}
	return ""
}

// grpcValue returns gRPC discovery value
func (i *discoveryInfo) grpcValue() string {
	if i.GRPCPort != 0 {
		return fmt.Sprintf("%s:%d", i.hostValue(), i.GRPCPort)
	}
	return ""
}

// mainPort returns main port, choosing between HTTP and gRPC
func (i *discoveryInfo) mainPort() int {
	if i.HTTPPort != 0 {
		return i.HTTPPort
	}
	return i.GRPCPort
}

// value returns discovery value for service
func (i *discoveryInfo) value() string {
	b, err := i.marshal()
	if err != nil {
		return ""
	}

	return string(b)
}

// marshal returns bytes representation of discoveryInfo or error
func (i *discoveryInfo) marshal() ([]byte, error) {
	return json.Marshal(i.discoveryValue())
}

// discoveryValue returns the discovery value to be marshaled
func (i *discoveryInfo) discoveryValue() DiscoveryValue {
	if i.PresetValue != nil {
		return *i.PresetValue
	}

	// Main endpoint is HTTP, however there can be apps, serving only GRPC connections
	// We should be able such situations
	value := DiscoveryValue{}
	if i.HTTPPort != 0 {
		value.EndpointMain = i.httpValue()
		value.EndpointAdditional = i.grpcValue()
	} else {
		value.EndpointMain = i.grpcValue()
	}
	return value
}

// validate validates discoveryInfo and returns an error
func (i *discoveryInfo) validate() error {
	errorTmpl := "invalid discovery Info: %s"

	if i.PresetValue != nil {
		// if PresetValue presents we should just validate it and that's it
		return i.PresetValue.validate()
	}

	if i.Host == "" {
		return fmt.Errorf(errorTmpl, "empty Host")
	}
	if i.HTTPPort == 0 && i.GRPCPort == 0 {
		return fmt.Errorf(errorTmpl, "both HTTPPort and GRPCPort are empty")
	}
	if i.HTTPPort < 0 || i.HTTPPort > 65535 {
		return fmt.Errorf(errorTmpl, "invalid HTTPPort")
	}
	if i.GRPCPort < 0 || i.GRPCPort > 65535 {
		return fmt.Errorf(errorTmpl, "invalid GRPCPort")
	}
	if _, err := i.marshal(); err != nil {
		return fmt.Errorf("error marshaling discovery value: %s", err)
	}

	return nil
}
