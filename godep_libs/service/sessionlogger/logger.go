package sessionlogger

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"motify_core_api/godep_libs/service/closer"
	"motify_core_api/godep_libs/service/config"
	"motify_core_api/godep_libs/service/dconfig"
	"motify_core_api/godep_libs/service/logger"
	"motify_core_api/godep_libs/service/watcher"
)

const (
	day = 24 * time.Hour

	SessionLoggerMode           = "SessionLoggerMode"
	SessionLoggingModeOff       = "OFF"
	SessionLoggingModeIndexOnly = "INDEX_ONLY"
	SessionLoggingModeFull      = "FULL"

	LogFileSuffix = ".log"
	DumpDirSuffix = ".dumps"
)

var (
	ErrSessionLoggerClosed = errors.New("Session logger is closed")

	errNoFileLogWriter = errors.New("No session file log writer")
)

type Logger struct {
	object       *object
	dconfManager *dconfig.Manager
}

type object struct {
	mutex sync.RWMutex
	//	f                *flake.Flake
	logsPath         string
	storeDays        uint16
	curFileLogWriter *FileLogWriter
	loggingMode      string
	stop             chan struct{}
	stopped          bool
}

var CleanPeriod = 12 * time.Hour

func init() {
	config.RegisterString("sessions_logs_path", "Path to sessions logs", "")
	config.RegisterUint("sessions_logs_store_days", "How long store logs in days", 1)
	config.RegisterBool("sessions_logs_auto_stop_enabled", "Disable session logging automatically", true)

	rand.Seed(time.Now().UnixNano())
	timeShift = time.Duration(rand.Intn(20)) * time.Second
}

func NewSessionLoggerFromFlags(dconfManager *dconfig.Manager) (*Logger, error) {
	logsPath, _ := config.GetString("sessions_logs_path")
	storeDays, _ := config.GetUint("sessions_logs_store_days")
	sessionLogger, err := NewLogger(logsPath, uint16(storeDays), dconfManager)
	if err != nil {
		return nil, err
	}
	if sessionLogger.object.loggingMode == SessionLoggingModeOff {
		logger.Notice(nil, "session logger is disabled")
		return sessionLogger, nil
	}

	closer.Add(func() {
		if err := sessionLogger.Close(); err != nil {
			logger.Critical(nil, "failed to close session logger: %v", err)
		}
	})

	if sessionLoggerAutoStopEnabled, _ := config.GetBool("sessions_logs_auto_stop_enabled"); sessionLoggerAutoStopEnabled && sessionLogger.dconfManager != nil {
		sessionLoggerAutoStopTimeoutDefault := 20 * time.Minute
		timeout := sessionLoggerAutoStopTimeoutDefault
		sessionLoggerAutoStopTimeout := &timeout
		dconfig.RegisterDuration("SessionLoggerAutoStopTimeout", "Stop session logger automatically timeout", sessionLoggerAutoStopTimeoutDefault, func(timeout time.Duration) {
			*sessionLoggerAutoStopTimeout = timeout
			if timeout == 0 {
				*sessionLoggerAutoStopTimeout = sessionLoggerAutoStopTimeoutDefault
			}
		})
		dconfig.RegisterString(SessionLoggerMode, "Session logging mode. Possible values: "+SessionLoggingModeOff+", "+SessionLoggingModeIndexOnly+", "+SessionLoggingModeFull+".", sessionLogger.object.loggingMode, func(mode string) {
			sessionLogger.SetLoggingMode(mode)
			sessionLogger.autoStop(mode, sessionLoggerAutoStopTimeout)
		})
	} else {
		dconfig.RegisterString(SessionLoggerMode, "Session logging mode. Possible values: "+SessionLoggingModeOff+", "+SessionLoggingModeIndexOnly+", "+SessionLoggingModeFull+".", sessionLogger.object.loggingMode, func(mode string) {
			sessionLogger.SetLoggingMode(mode)
		})
	}
	return sessionLogger, nil
}

func NewLogger(logsPath string, storeDays uint16, dconfManager *dconfig.Manager) (*Logger, error) {
	o := &object{
		logsPath:    logsPath,
		storeDays:   storeDays,
		loggingMode: SessionLoggingModeOff,
		stop:        make(chan struct{}),
	}

	watcher.Watch(func() { o.cleanLogs() }, o.stop, CleanPeriod)

	_, err := o.getOrCreateFileLogWriter()
	if err != nil {
		return nil, err
	}

	l := &Logger{object: o, dconfManager: dconfManager}

	runtime.SetFinalizer(l, func(logger *Logger) { logger.Close() })

	return l, nil
}

func (logger *Logger) NewSession(traceID string, caption string, request interface{}) (*Session, error) {
	return logger.object.NewSession(traceID, caption, request)
}

func (logger *Logger) NewSessionWithFullLoggingMode(traceID string, caption string, request interface{}) (*Session, error) {
	return logger.object.NewSessionWithFullLoggingMode(traceID, caption, request)
}

func (logger *Logger) Close() error {
	return logger.object.Close()
}

func (logger *Logger) SetLoggingMode(mode string) {
	logger.object.SetLoggingMode(mode)
}

func (o *object) SetLoggingMode(mode string) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if mode != SessionLoggingModeOff && mode != SessionLoggingModeIndexOnly && mode != SessionLoggingModeFull {
		return
	}

	o.loggingMode = mode
}

func (o *object) Close() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.stopped || o.curFileLogWriter == nil {
		return nil
	}

	err := o.curFileLogWriter.Close()
	o.curFileLogWriter = nil
	close(o.stop)
	o.stopped = true

	return err
}

func (o *object) newSession(traceID string, caption string, request interface{}, mode string) (*Session, error) {
	logWriter, err := o.rotateFileLogWriter()
	if err != nil {
		return nil, err
	}

	root := &rootSession{
		logWriter:   logWriter,
		traceID:     traceID,
		idGen:       0,
		loggingMode: mode,
	}

	return startSession(root, 0, caption, request), nil
}

func (o *object) NewSession(traceID string, caption string, request interface{}) (*Session, error) {
	return o.newSession(traceID, caption, request, o.loggingMode)
}

func (o *object) NewSessionWithFullLoggingMode(traceID string, caption string, request interface{}) (*Session, error) {
	return o.newSession(traceID, caption, request, SessionLoggingModeFull)
}

func currentDay() time.Time {
	return time.Now().UTC().Truncate(day)
}

func (o *object) rotateFileLogWriter() (*FileLogWriter, error) {
	fileLogWriter, err := o.getFileLogWriter()

	if err == nil {
		return fileLogWriter, nil
	} else if err == errNoFileLogWriter {
	} else {
		return fileLogWriter, err
	}

	return o.getOrCreateFileLogWriter()
}

func (o *object) getFileLogWriter() (*FileLogWriter, error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	if o.stopped {
		return nil, ErrSessionLoggerClosed
	}
	if o.logsPath == "" || o.storeDays == 0 {
		// TODO: for now I return nil log writer to not log.
		// In the future I'll make Session interface, PersistentSession and DiscardSession.
		// I cant do it now because of huge master branch conflicts.
		return nil, nil
	}
	date := currentDay()
	if o.curFileLogWriter != nil && date.Equal(o.curFileLogWriter.Date()) {
		return o.curFileLogWriter, nil
	}

	return nil, errNoFileLogWriter
}

func (o *object) getOrCreateFileLogWriter() (*FileLogWriter, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.stopped {
		return nil, ErrSessionLoggerClosed
	}
	if o.logsPath == "" || o.storeDays == 0 {
		return nil, nil
	}
	date := currentDay()
	if o.curFileLogWriter != nil && date.Equal(o.curFileLogWriter.Date()) {
		return o.curFileLogWriter, nil
	}

	writer, err := o.createFileLogWriter(date)
	if err != nil {
		return nil, err
	}

	o.curFileLogWriter = writer
	return writer, nil
}

func (o *object) createFileLogWriter(date time.Time) (*FileLogWriter, error) {
	fmtDate := date.Format("2006-01-02")

	err := os.MkdirAll(o.logsPath, 0750)
	if err != nil {
		return nil, err
	}

	logFileName := path.Join(o.logsPath, fmtDate+LogFileSuffix)
	logDirName := path.Join(o.logsPath, fmtDate+DumpDirSuffix)

	return NewFileLogWriter(date, logFileName, logDirName)
}

func (o *object) cleanLogs() {
	now := time.Now().UTC()
	days := day * time.Duration(o.storeDays)
	minTime := now.Add(-days).Truncate(day)
	clean(o.logsPath+"/*"+LogFileSuffix, minTime)
	clean(o.logsPath+"/*"+DumpDirSuffix, minTime)
}

func clean(pattern string, minTime time.Time) {
	fileNames, err := filepath.Glob(pattern)
	if err != nil {
		return
	}
	for _, fileName := range fileNames {
		fileStat, err := os.Stat(fileName)
		if err != nil {
			logger.Error(nil, "failed to stat session logs: %s", err)
			continue
		}
		if fileStat.ModTime().Before(minTime) {
			err := os.RemoveAll(fileName)
			if err != nil {
				logger.Error(nil, "failed to clean session logs: %s", err)
			}
		}
	}
}

var autoStopMutex sync.Mutex
var lastUpdateTime = time.Now()
var lastMode = ""
var timeShift time.Duration

func (logger *Logger) autoStop(mode string, sessionLoggerAutoStopTimeout *time.Duration) {
	autoStopMutex.Lock()
	defer autoStopMutex.Unlock()

	lastUpdateTime = time.Now()
	lastMode = mode

	if mode != SessionLoggingModeOff {
		mode := lastMode
		thisUpdateTime := lastUpdateTime
		go time.AfterFunc(*sessionLoggerAutoStopTimeout+timeShift, func() {
			if lastUpdateTime == thisUpdateTime && mode == lastMode {
				logger.dconfManager.EditSetting(context.Background(), "", SessionLoggerMode, SessionLoggingModeOff)
			}
		})
	}
}
