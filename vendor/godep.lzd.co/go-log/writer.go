package log

import (
	"io"
	"net"
	"sync"
	"time"
)

const timeout = 50 * time.Millisecond

type writer struct {
	hostname string
	network  string
	address  string
	mu       sync.RWMutex
	conn     io.WriteCloser
}

func newWriter(network, address string) *writer {
	w := &writer{
		network: network,
		address: address,
	}

	w.connect()
	return w
}

func (w *writer) connect() error {
	w.cls()

	var err error
	switch {
	case w.network == "unixgram":
		w.conn, err = newUnixgramWriter(w.address)
	case w.network == "unixgram_old":
		w.conn, err = newUnixgramWriter_old(w.address)
	case w.network == "" || w.address == "":
		w.conn, err = newStdoutWriter()
	default:
		var c net.Conn
		c, err = net.Dial(w.network, w.address)
		w.conn = &netConnWriter{c}
	}

	if err != nil {
		w.conn = nil
	}

	return err
}

type doNotRetryError struct {
	error
}

func (w *writer) Write(msg []byte) (int, error) {
	// allows parallel write but guards w.conn
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.conn != nil {
		if n, err := w.conn.Write(msg); err == nil {
			return n, err
		} else if _, ok := err.(doNotRetryError); ok {
			return 0, err
		}
	}

	w.mu.RUnlock()
	w.mu.Lock()
	err := w.connect()
	w.mu.Unlock()
	w.mu.RLock()

	if err != nil {
		return 0, err
	}
	return w.conn.Write(msg)
}

func (w *writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.cls()
}

func (w *writer) cls() (err error) {
	if w.conn != nil {
		err = w.conn.Close()
		w.conn = nil
	}
	return
}
