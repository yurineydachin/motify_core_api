package testutils

import (
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type TestingLogger interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Log(args ...interface{})
	Logf(format string, args ...interface{})
}

func PrepareTemporaryDir(l TestingLogger) (temporaryDir string) {
	return prepareTemporaryDirNamedByCallerLevel(l, 2)
}

func prepareTemporaryDirNamedByCallerLevel(l TestingLogger, callerLevel int) (temporaryDir string) {
	callerFuncPtr, _, _, _ := runtime.Caller(callerLevel)
	callerFunc := runtime.FuncForPC(callerFuncPtr)

	dirPrefix := strings.Split(callerFunc.Name(), ".")[1]
	dirPrefix = strings.Replace(dirPrefix, "/", "__", -1)

	temporaryDir, err := ioutil.TempDir("/tmp", dirPrefix)
	if err != nil {
		l.Fatalf("Can't create temporary directory: %s", err)
	}

	l.Logf("Temporary directory: %s", temporaryDir)

	return temporaryDir
}

func Lsof(l TestingLogger, filePath string) []string {
	out, err := exec.Command("lsof", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		l.Fatal(err)
	}

	lines := strings.Split(string(out), "\n")
	openedHandlers := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, filePath) {
			openedHandlers = append(openedHandlers, line)
		}
	}

	return openedHandlers
}
