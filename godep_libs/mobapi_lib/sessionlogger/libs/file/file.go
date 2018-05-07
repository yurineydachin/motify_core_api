package file

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"motify_core_api/godep_libs/mobapi_lib/logger"
	"motify_core_api/godep_libs/mobapi_lib/watcher"
)

type SelfRescuingFile struct {
	object *object
}

type object struct {
	file    *os.File
	name    string
	flag    int
	perm    os.FileMode
	err     error
	mutex   sync.RWMutex
	stop    chan struct{}
	stopped bool
}

var RescuingPeriod = 5 * time.Minute

func Open(name string) (*SelfRescuingFile, error) {
	return OpenFile(name, os.O_RDONLY, 0)
}

func OpenFile(name string, flag int, perm os.FileMode) (file *SelfRescuingFile, err error) {
	o := &object{
		name: name,
		flag: flag,
		perm: perm,
		stop: make(chan struct{}),
	}

	o.file, err = o.open()
	if err != nil {
		return nil, err
	}

	file = &SelfRescuingFile{object: o}

	if !o.readOnly() {
		runtime.SetFinalizer(file, func(file *SelfRescuingFile) { file.object.Close() })
		watcher.Watch(func() { o.maintain() }, o.stop, RescuingPeriod)
	}

	return file, nil
}

func (file *SelfRescuingFile) Name() string {
	return file.object.Name()
}
func (file *SelfRescuingFile) Write(p []byte) (n int, err error) {
	return file.object.Write(p)
}

func (file *SelfRescuingFile) Close() error {
	return file.object.Close()
}

func (o *object) open() (*os.File, error) {
	return os.OpenFile(o.name, o.flag, o.perm)
}

func (o *object) Close() error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.stopped {
		return nil
	}
	close(o.stop)
	o.stopped = true
	o.err = fmt.Errorf("File '%s' is closed", o.Name())
	return o.close()
}

func (o *object) close() error {
	return o.file.Close()
}

func (o *object) Name() string {
	return o.name
}

func (o *object) Write(p []byte) (n int, err error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	if o.err != nil {
		return 0, o.err
	}

	return o.file.Write(p)
}

func (o *object) readOnly() bool {
	return o.flag == os.O_RDONLY
}

func (o *object) maintain() error {
	if _, err := os.Stat(o.name); err != nil {
		return o.reopen(err)
	}

	return nil
}

func (o *object) reopen(err error) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.stopped {
		return nil
	}

	logger.Error(nil, "File '%s' was stolen: %s", o.name, err)

	file, err := o.open()
	if err != nil {
		logger.Error(nil, "%v", err)
		o.err = err
		return err
	}

	err = o.close()
	if err != nil {
		logger.Error(nil, "Can't close file %s: %s", o.name, err)
	}

	o.file = file
	o.err = nil

	return nil
}
