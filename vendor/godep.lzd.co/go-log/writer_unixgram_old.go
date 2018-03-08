package log

import (
	"syscall"
	"sync/atomic"
	"time"
)

type unixgramWriter_old struct {
	fd     int
	path   string
	errors uint64
}

func newUnixgramWriter_old(path string) (*unixgramWriter_old, error) {
	fd, err := syscall.Socket(syscall.AF_UNIX, syscall.SOCK_DGRAM, 0)
	if err != nil {
		return nil, err
	}
	return &unixgramWriter_old{fd: fd, path: path}, nil
}

const errorsCountBeforeReconnect = 1e6
const maxAttempts uint64 = 2000

func (w *unixgramWriter_old) write(msg []byte, attempts uint64) (int, error) {
	for {
		if err := syscall.Sendto(w.fd, msg, syscall.MSG_DONTWAIT, &syscall.SockaddrUnix{Name: w.path}); err != nil {
			if err == syscall.EAGAIN {
				atomic.AddUint64(&w.errors, 1)
				if attempts > 1 {
					attempts--
					time.Sleep(1)
					continue
				}
			}
			if c := atomic.LoadUint64(&w.errors); c > errorsCountBeforeReconnect {
				atomic.StoreUint64(&w.errors, 0)
				return 0, err
			}
			return 0, doNotRetryError{err}
		}
		atomic.StoreUint64(&w.errors, 0)
		return len(msg), nil
	}
}

func (w *unixgramWriter_old) Close() error {
	return syscall.Close(w.fd)
}

func (w *unixgramWriter_old) Write(msg []byte) (int, error) {
	return w.write(msg, w.backoff())
}

func (w *unixgramWriter_old) backoff() uint64 {
	e := atomic.LoadUint64(&w.errors)
	if e >= maxAttempts {
		return 1
	}
	return maxAttempts - e
}
