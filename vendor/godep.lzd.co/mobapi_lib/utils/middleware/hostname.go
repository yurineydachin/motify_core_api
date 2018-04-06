package middleware

import (
	"godep.lzd.co/go-config"
	"os"
)

func GetHostname() (string, error) {
	hostname, _ := config.GetString("advertised-hostname")

	if hostname != "" {
		return hostname, nil
	}

	return os.Hostname()
}
