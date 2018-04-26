package uniq_dumper

import (
	"fmt"

	"crypto/md5"
	"godep.lzd.co/mobapi_lib/sessionlogger/libs/locker"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var lock = locker.New()

type Dumper struct {
	dir string
}

func New(dir string) *Dumper {
	return &Dumper{
		dir: dir,
	}
}

func (d *Dumper) Write(record []byte) (string, error) {
	hash := fmt.Sprintf("%x", md5.Sum(record))

	if lock.IsLocked(hash) || d.Exists(hash) {
		return hash, nil
	}

	lock.Lock(hash)
	defer lock.Unlock(hash)

	if d.Exists(hash) {
		return hash, nil
	}

	return d.write(hash, record)
}

func (d *Dumper) write(hash string, record []byte) (string, error) {
	file := d.getFile(hash)
	dir := filepath.Dir(file)

	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return "", err
	}

	fileTmp := file + ".tmp"
	err = ioutil.WriteFile(fileTmp, record, 0500)
	if err != nil {
		return "", err
	}

	err = os.Rename(fileTmp, file)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func (d *Dumper) Read(hash string) ([]byte, error) {
	return ioutil.ReadFile(d.getFile(hash))
}

func (d *Dumper) Exists(hash string) bool {
	_, err := os.Stat(d.getFile(hash))
	return err == nil
}

func (d *Dumper) getFile(hash string) string {
	file, counter := d.dir, 0
	for _, symbol := range strings.Split(hash, "") {
		if counter%4 == 0 {
			file += "/"
		}
		file += symbol
		counter++
	}

	return file
}
