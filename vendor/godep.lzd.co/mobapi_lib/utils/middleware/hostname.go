package middleware

import (
	"os"
	"godep.lzd.co/go-config"
)

func GetHostname() (string, error) {
	hostname, _ := config.GetString("advertised-hostname")

	if hostname != "" {
		return hostname, nil
	}

	return os.Hostname()
}
