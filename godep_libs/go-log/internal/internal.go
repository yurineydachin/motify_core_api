package internal

import (
	"fmt"
	"path/filepath"
	"runtime"
)

func CurrentLineMinusOne() string {
	_, file, line, _ := runtime.Caller(1)
	_, file = filepath.Split(file)
	return fmt.Sprintf("%s:%d", file, line-1)
}
