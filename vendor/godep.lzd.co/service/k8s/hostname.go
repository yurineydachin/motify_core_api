package k8s

import (
	"godep.lzd.co/service/config"
	"os"
)

const CustomHostnameParamName = "advertised-hostname"

func init() {
	config.RegisterString(CustomHostnameParamName, "os.Hostname replacement", "")
}

func GetHostname() (string, error) {
	hostname, _ := config.GetString(CustomHostnameParamName)

	if hostname != "" {
		return hostname, nil
	}

	return os.Hostname()
}
