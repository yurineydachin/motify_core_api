package log

import (
	"net"
)

type unixgramWriter struct {
	conn *net.UnixConn
	path string
}

func newUnixgramWriter(path string) (*unixgramWriter, error) {
	conn, err := net.DialUnix("unixgram", nil, &net.UnixAddr{path, "unixgram"})
	if err != nil {
		return nil, err
	}

	writer := &unixgramWriter{conn, path}

	return writer, nil
}

func (w *unixgramWriter) Close() error {
	return w.conn.Close()
}

func (w *unixgramWriter) Write(msg []byte) (int, error) {
	return w.conn.Write(msg)
}
