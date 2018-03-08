package registrator

import (
	"encoding/json"
	"fmt"
	"time"
)

// VersionInfo contain service version data.
// Some fields are optional, but they help to structurize the data anyway
type VersionInfo struct {
	AppVersion  string     `json:"app_version"`
	BuildDate   *time.Time `json:"build_date,omitempty"`
	GitDescribe string     `json:"git_describe,omitempty"`
	GoVersion   string     `json:"go_version,omitempty"`
}

// validate checks if the data is valid for registration
func (i *VersionInfo) validate() error {
	switch {
	case i.AppVersion == "":
		return fmt.Errorf("invalid version info: AppVersion is empty")
	}
	return nil
}

// adminInfo contains registration data for "admin" namespace
type adminInfo struct {
	AdminInterfaceURI string      `json:"admin_interface"`
	Version           VersionInfo `json:"version"`
}

// NewAdminInfoFromString returns new adminInfo from string value
func NewAdminInfoFromString(s string) (adminInfo, error) {
	info := adminInfo{}
	if err := json.Unmarshal([]byte(s), &info); err != nil {
		return adminInfo{}, err
	}

	return info, nil
}

// value returns string value of struct
func (i *adminInfo) value() string {
	b, err := json.Marshal(i)
	if err != nil {
		return ""
	}

	return string(b)
}

// validate checks if the data is valid for registration
func (i *adminInfo) validate() error {
	switch {
	case i.AdminInterfaceURI == "":
		return fmt.Errorf("invalid admin info: AdminInterfaceURI is empty")
	}

	return i.Version.validate()
}

func (i *adminInfo) isEmpty() bool {
	return *i == adminInfo{}
}
