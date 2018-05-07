package balancer

import (
	"errors"
	"net"
	"time"
)

// DialCheck tries to connect to host:port over tcp
func DialCheck(address string, timeout time.Duration) error {
	if address == "" {
		return errors.New("DialCheck: empty address")
	}
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return err
	}

	if err := conn.Close(); err != nil {
		return err
	}

	return nil
}
