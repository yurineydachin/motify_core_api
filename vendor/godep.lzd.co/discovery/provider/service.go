package provider

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reRolloutType = regexp.MustCompile(`^(stable|unstable\d{1,2})$`)
)

const (
	// DefaultOwner is the default Owner field value
	DefaultOwner = "shared"
	// DefaultClusterType is the default ClusterType field value
	DefaultClusterType = "common"
	// RolloutTypeStable is a rolloutType string of stable services
	RolloutTypeStable = "stable"

	errorTmpl = "invalid Service info: %s"
	// service names are limited in length - GOLIBS-1073
	maxServiceNameLen = 30
	// keyPartSeparator should not be used in any part of the key
	keyPartSeparator  = "/"
	separatorErrorMsg = " contains forbidden '/' char"
)

// checker is a DSL struct for validating Service data
type checker struct {
	val            string
	checkSeparator bool
}

// Service contains data for service identification
type Service struct {
	Name         string
	Type         ServiceType  // "app" | "external" | "system"
	Owner        string       // app name or "shared"
	RolloutType  string       // "stable" | "unstableN"
	ClusterType  string       // "common", "master", "slave", etc
	InstanceName InstanceName // "host:port" for apps or other strings for other types
}

// InstanceKey return service instance key string
func (s *Service) InstanceKey() string {
	return s.InstanceName.String()
}

// Validate checks if the data is valid for registration
func (s *Service) Validate() error {
	nonEmptyFields := map[string]checker{
		"Name":         checker{s.Name, true},
		"Type":         checker{s.Type.String(), false}, // Type is a enum, no need to check
		"Owner":        checker{s.Owner, true},
		"ClusterType":  checker{s.ClusterType, true},
		"InstanceName": checker{s.InstanceName.String(), true},
	}

	for label, field := range nonEmptyFields {
		if field.val == "" {
			return fmt.Errorf(errorTmpl, label+" is empty")
		} else if field.checkSeparator && strings.Contains(field.val, keyPartSeparator) {
			return fmt.Errorf(errorTmpl, label+separatorErrorMsg)
		}
	}

	switch {
	case len([]rune(s.Name)) > maxServiceNameLen:
		return fmt.Errorf(errorTmpl, fmt.Sprintf("Name %q is too long, max len is %d symbols", s.Name, maxServiceNameLen))
	case !reRolloutType.MatchString(s.RolloutType):
		return fmt.Errorf(errorTmpl, "RolloutType is invalid")
	}
	return nil
}
