package registrator

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
)

// monitoringInfo contains registration data for monitoring
type monitoringInfo struct {
	Host        string
	Port        int
	Environment string
	Venture     string
	ServiceName string
	RolloutType string
}

// monitoringValue is monitoringInfo value struct for registering in provider
type monitoringValue struct {
	URL  string       `json:"url"`
	Tags []metricsTag `json:"tags"`
}

// metricsTag is a tag for host
type metricsTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// NewMonitoringInfoFromString returns new monitoringInfo from string value
func NewMonitoringInfoFromString(s string) (monitoringInfo, error) {
	info := monitoringInfo{}
	if err := json.Unmarshal([]byte(s), info); err != nil {
		return monitoringInfo{}, err
	}

	return info, nil
}

// value returns string value of struct
func (i *monitoringInfo) value() string {
	b, err := i.marshal()
	if err != nil {
		return ""
	}

	return string(b)
}

func (i *monitoringInfo) marshal() ([]byte, error) {
	return json.Marshal(monitoringValue{
		URL: fmt.Sprintf("http://%s/metrics", net.JoinHostPort(i.Host, strconv.Itoa(i.Port))),
		Tags: []metricsTag{
			{"venture", i.Venture},
			{"env", i.Environment},
			{"service", i.ServiceName},
			{"rollout_type", i.RolloutType},
		},
	})
}

func (i *monitoringInfo) validate() error {
	errorTmpl := "invalid monitoring info: %s"
	switch {
	case i.ServiceName == "":
		return fmt.Errorf(errorTmpl, "ServiceName is empty")
	case i.Host == "":
		return fmt.Errorf(errorTmpl, "Host is empty")
	case i.Port <= 0 || i.Port > 65535:
		return fmt.Errorf(errorTmpl, "invalid Port")
	case i.Environment == "":
		return fmt.Errorf(errorTmpl, "Environment is empty")
	case i.Venture == "":
		return fmt.Errorf(errorTmpl, "Venture is empty")
	case i.RolloutType == "":
		return fmt.Errorf(errorTmpl, "RolloutType is empty")
	}

	if _, err := i.marshal(); err != nil {
		return fmt.Errorf("error marshaling monitoringInfo: %s", err)
	}

	return nil
}

func (i *monitoringInfo) isEmpty() bool {
	return *i == monitoringInfo{}
}
