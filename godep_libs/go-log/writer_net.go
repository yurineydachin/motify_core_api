package log

import (
	"net"
	"time"
)

type netConnWriter struct {
	net.Conn
}

func (w *netConnWriter) Write(msg []byte) (int, error) {
	w.SetWriteDeadline(time.Now().Add(timeout))
	return w.Conn.Write(msg)
}

