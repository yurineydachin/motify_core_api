package uniq_dumper_test

import (
	"godep.lzd.co/service/sessionlogger/libs/testutils"
	"godep.lzd.co/service/sessionlogger/libs/uniq_dumper"
	"io/ioutil"
	"os"
	"testing"
)

func TestDumper(t *testing.T) {
	temporaryDir := testutils.PrepareTemporaryDir(t)
	defer os.RemoveAll(temporaryDir)

	record := "test_record"
	md5 := "519b2c78b4c8932c23480ae1c800e55d"
	file := temporaryDir + "/" + "519b/2c78/b4c8/932c/2348/0ae1/c800/e55d"

	d := uniq_dumper.New(temporaryDir)

	if d.Exists(md5) {
		t.Errorf("Record mustn't exist")
	}
	{
		recordGot, err := d.Read(md5)
		if !os.IsNotExist(err) {
			t.Fatalf("Incorrect err: %s", err)
		}
		if recordGot != nil {
			t.Fatalf("Record must be nil")
		}
	}

	{
		md5Got, err := d.Write([]byte(record))
		if err != nil {
			t.Fatalf("Error: %s", err)
		}

		if md5Got != md5 {
			t.Errorf("MD5: got '%s', expected '%s'", md5Got, md5)
		}
	}
	{
		recordGot, err := ioutil.ReadFile(file)
		if err != nil {
			t.Fatalf("Error: %s", err)
		}

		if record != string(recordGot) {
			t.Errorf("Record: got '%s', expected: '%s'", recordGot, record)
		}
	}
	{
		recordGot, err := d.Read(md5)
		if err != nil {
			t.Fatalf("Error: %s", err)
		}

		if record != string(recordGot) {
			t.Errorf("Record: got '%s', expected: '%s'", recordGot, record)
		}
	}

	if !d.Exists(md5) {
		t.Errorf("Record must exist")
	}
}
