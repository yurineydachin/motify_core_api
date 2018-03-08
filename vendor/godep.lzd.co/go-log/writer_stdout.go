package log

import (
	"os"
)

type stdoutWriter struct {}

func newStdoutWriter() (*stdoutWriter, error) {
	return &stdoutWriter{}, nil
}

func (w *stdoutWriter) Write(msg []byte) (int, error) {
	return os.Stdout.Write(msg)
}

func (w *stdoutWriter) Close() error {
	return nil
}
