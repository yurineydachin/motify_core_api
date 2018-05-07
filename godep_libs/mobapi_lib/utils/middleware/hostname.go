package middleware

import (
	"motify_core_api/godep_libs/go-config"
	"os"
)

func GetHostname() (string, error) {
	hostname, _ := config.GetString("advertised-hostname")

	if hostname != "" {
		return hostname, nil
	}

	return os.Hostname()
}
