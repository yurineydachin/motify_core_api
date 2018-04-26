package sessionlogger

import (
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"runtime"
	"testing"
	"time"

	"godep.lzd.co/service/sessionlogger/libs/file"
	"godep.lzd.co/service/sessionlogger/libs/testutils"
)

var zeroDuration = time.Duration(0)
var logSessionName = "TestLogWriter"
var (
	logFileWasStollenRegexp = regexp.MustCompile(`ERROR\s.*: Log file was stollen`)
	logNewSessionRegexp     = regexp.MustCompile(`REQ\s+` + logSessionName)
	logFinishRegexp         = regexp.MustCompile(`RESP\s+` + logSessionName)
)

func TestFileLogWriterGarbageCollection(t *testing.T) {
	// lsof is not allowed on UT server
	t.SkipNow()

	temporaryDir := testutils.PrepareTemporaryDir(t)
	defer os.RemoveAll(temporaryDir)

	logName := path.Join(temporaryDir, "log")
	logDir := path.Join(temporaryDir, "dumps")

	_, err := NewFileLogWriter(time.Now(), logName, logDir)
	if err != nil {
		t.Fatal(err)
	}

	runtime.GC()

	{
		openedHandlers := testutils.Lsof(t, logName)
		if len(openedHandlers) > 0 {
			t.Errorf("File handlers are still opened after garbage collection:\n%s\n", openedHandlers)
		}
	}
}

func TestLoggerGarbageCollection(t *testing.T) {
	// lsof is not allowed on UT server
	t.SkipNow()

	temporaryDir := testutils.PrepareTemporaryDir(t)
	defer os.RemoveAll(temporaryDir)

	var logName string
	{
		appLogger := createLogger(t, temporaryDir)

		logFile := appLogger.object.curFileLogWriter.object.file

		logName = logFile.Name()

		t.Logf("log file: %s", logName)
	}

	runtime.GC()

	{
		openedHandlers := testutils.Lsof(t, logName)
		if len(openedHandlers) > 0 {
			t.Errorf("File handlers are still opened after garbage collection:\n%s\n", openedHandlers)
		}
	}
}

func TestLoggerNewSessionGarbageCollection(t *testing.T) {
	// lsof is not allowed on UT server
	t.SkipNow()

	temporaryDir := testutils.PrepareTemporaryDir(t)
	defer os.RemoveAll(temporaryDir)

	var logName string
	{
		appLogger := createLogger(t, temporaryDir)

		logFile := appLogger.object.curFileLogWriter.object.file
		logName = logFile.Name()

		t.Logf("log file: %s", logName)

		sessionLogger, err := appLogger.NewSession("id", logSessionName, struct{}{})
		if err != nil {
			t.Fatalf("Can't start new logger session: %s", err)
		}

		sessionLogger.Finish(struct{}{})
	}

	runtime.GC()

	{
		openedHandlers := testutils.Lsof(t, logName)
		if len(openedHandlers) > 0 {
			t.Errorf("File handlers are still opened after garbage collection:\n%s\n", openedHandlers)
		}
	}
}

func TestLoggerWriterChangeGarbageCollection(t *testing.T) {
	// lsof is not allowed on UT server
	t.SkipNow()

	temporaryDir := testutils.PrepareTemporaryDir(t)
	defer os.RemoveAll(temporaryDir)

	var logName string
	appLogger := createLogger(t, temporaryDir)

	{

		logFile := appLogger.object.curFileLogWriter.object.file
		logName = logFile.Name()
	}

	t.Logf("log file: %s", logName)

	{
		sessionLogger, err := appLogger.NewSession("id", logSessionName, struct{}{})
		if err != nil {
			t.Fatalf("Can't start new logger session: %s", err)
		}

		sessionLogger.Finish(struct{}{})
	}

	{
		yesterday := currentDay().Add(-24 * time.Hour)

		logWriter, err := appLogger.object.createFileLogWriter(yesterday)
		if err != nil {
			t.Fatalf("Can't create writer: %s", err)
		}

		appLogger.object.curFileLogWriter = logWriter
		logWriter.Write(&Log{})
	}

	{
		sessionLogger, err := appLogger.NewSession("id", logSessionName, struct{}{})
		if err != nil {
			t.Fatalf("Can't start new logger session: %s", err)
		}

		sessionLogger.Finish(struct{}{})
	}

	runtime.GC()

	{
		openedHandlers := testutils.Lsof(t, logName)
		if len(openedHandlers) > 1 {
			t.Errorf("File handlers are still opened after garbage collection:\n%s\n", openedHandlers)
		}
	}
}

func TestLogWriter(t *testing.T) {
	temporaryDir := testutils.PrepareTemporaryDir(t)
	defer os.RemoveAll(temporaryDir)

	appLogger := createLogger(t, temporaryDir)
	appLogger.Close()

	appLogger = createLogger(t, temporaryDir)
	appLogger.SetLoggingMode(SessionLoggingModeIndexOnly)

	logFile := appLogger.object.curFileLogWriter.object.file.Name()

	t.Logf("log file: %s", logFile)

	sessionLogger, err := appLogger.NewSession("id", logSessionName, struct{}{})
	if err != nil {
		t.Fatalf("Can't start new logger session: %s", err)
	}

	sessionLogger.Finish(struct{}{})

	appLogger.Close()

	logWritten := readFile(t, logFile)
	if !logNewSessionRegexp.MatchString(logWritten) {
		t.Errorf("Written log doesn't contain message about new session\nRegexp: '%s'\nContent:\n%s", logNewSessionRegexp.String(), logWritten)
	}
	if !logFinishRegexp.MatchString(logWritten) {
		t.Errorf("Written log doesn't contain message about finish\nRegexp: '%s'\nContent:\n%s", logFinishRegexp.String(), logWritten)
	}

}

func readFile(t *testing.T, filepath string) string {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Can't read log capture file content: %s", string(content))
	}

	return string(content)
}

func BenchmarkCreateFileLogWriter(b *testing.B) {
	temporaryDir := testutils.PrepareTemporaryDir(b)
	defer os.RemoveAll(temporaryDir)

	appLogger := createLogger(b, temporaryDir)

	date := time.Unix(0, 0)
	for i := 0; i < b.N; i++ {
		_, err := appLogger.object.createFileLogWriter(date)
		if err != nil {
			b.Fatalf("Can't create file log writer: %s", err)
		}

		date = date.AddDate(0, 0, 1)
	}
}

func BenchmarkCreateAndCloseFileLogWriter(b *testing.B) {
	temporaryDir := testutils.PrepareTemporaryDir(b)
	defer os.RemoveAll(temporaryDir)

	appLogger := createLogger(b, temporaryDir)
	defer appLogger.Close()

	date := time.Unix(0, 0)
	for i := 0; i < b.N; i++ {
		_, err := appLogger.object.createFileLogWriter(date)
		if err != nil {
			b.Fatalf("Can't create file log writer: %s", err)
		}

		err = appLogger.Close()
		if err != nil {
			b.Fatalf("Can't close file log writer: %s", err)
		}

		date = date.AddDate(0, 0, 1)
	}
}

func BenchmarkGetFileLogWriter(b *testing.B) {
	temporaryDir := testutils.PrepareTemporaryDir(b)
	defer os.RemoveAll(temporaryDir)

	appLogger := createLogger(b, temporaryDir)
	defer appLogger.Close()

	_, err := appLogger.object.createFileLogWriter(time.Now().UTC())
	if err != nil {
		b.Fatalf("Can't create file log writer: %s", err)
	}

	for i := 0; i < b.N; i++ {
		_, err := appLogger.object.getFileLogWriter()
		if err != nil {
			b.Fatalf("Can't get file log writer: %s", err)
		}
	}
}

func BenchmarkNewSession(b *testing.B) {
	temporaryDir := testutils.PrepareTemporaryDir(b)
	defer os.RemoveAll(temporaryDir)

	appLogger := createLogger(b, temporaryDir)
	defer appLogger.Close()

	_, err := appLogger.object.createFileLogWriter(currentDay())
	if err != nil {
		b.Fatalf("Can't create file log writer: %s", err)
	}

	for i := 0; i < b.N; i++ {
		_, err := appLogger.NewSession("id", logSessionName, struct{}{})
		if err != nil {
			b.Fatalf("Can't start new logger session: %s", err)
		}
	}
}

func BenchmarkSessionLog(b *testing.B) {
	temporaryDir := testutils.PrepareTemporaryDir(b)
	defer os.RemoveAll(temporaryDir)

	appLogger := createLogger(b, temporaryDir)
	defer appLogger.Close()

	_, err := appLogger.object.createFileLogWriter(currentDay())
	if err != nil {
		b.Fatalf("Can't create file log writer: %s", err)
	}

	sessionLogger, err := appLogger.NewSession("id", logSessionName, struct{}{})
	if err != nil {
		b.Fatalf("Can't start new logger session: %s", err)
	}

	for i := 0; i < b.N; i++ {
		sessionLogger.log("TEST", struct{}{})
	}
}

func createLogger(l testutils.TestingLogger, temporaryDir string) *Logger {
	CleanPeriod = zeroDuration
	file.RescuingPeriod = zeroDuration

	appLogger, err := NewLogger(temporaryDir, uint16(1), nil)
	if err != nil {
		l.Fatalf("Can't create logger: %s", err)
	}

	return appLogger
}
