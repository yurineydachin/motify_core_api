package file

import (
	"fmt"
	"motify_core_api/godep_libs/service/sessionlogger/libs/testutils"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"testing"
)

func TestSelfRescuingFileGarbageCollection(t *testing.T) {
	// lson is not allowed on UT server
	t.SkipNow()

	temporaryDir := testutils.PrepareTemporaryDir(t)
	defer os.RemoveAll(temporaryDir)

	filePath := path.Join(temporaryDir, "file")
	_, err := OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0640)
	if err != nil {
		t.Fatalf("Can't open file '%s'", filePath)
	}

	runtime.GC()

	openedHandlers := testutils.Lsof(t, filePath)
	if len(openedHandlers) > 0 {
		t.Errorf("Handlers are still opened after garbage collection:\n%s\n", openedHandlers)
	}
}

func TestSelfRescuingFileWriting(t *testing.T) {
	temporaryDir := testutils.PrepareTemporaryDir(t)
	defer os.RemoveAll(temporaryDir)

	filePath := path.Join(temporaryDir, "file")
	record := "Test\n"

	file, err := OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0640)
	if err != nil {
		t.Fatalf("Can't open file '%s' for writing: %s", file.Name(), err)
	}

	_, err = fmt.Fprintf(file, record)
	if err != nil {
		t.Fatalf("Cant write to the file '%s': %s", file.Name(), err)
	}

	err = file.Close()
	if err != nil {
		t.Fatalf("Can't close file '%s': %s", file.Name(), err)
	}

	file, err = Open(filePath)
	if err != nil {
		t.Fatalf("Can't open file '%s' for reading: %s", file.Name(), err)
	}

	err = file.Close()
	if err != nil {
		t.Fatalf("Can't close file '%s': %s", file.Name(), err)
	}

	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Can't read from file '%s': %s", file.Name(), err)
	}

	if string(contents) != record {
		t.Errorf("File contents is different from expected:\n'%s' - Got\n'%s' - Expected", string(contents), record)
	}
}

func TestSelfRescuingFileStolenRescue(t *testing.T) {
	temporaryDir := testutils.PrepareTemporaryDir(t)
	defer os.RemoveAll(temporaryDir)

	filePath := path.Join(temporaryDir, "file")
	record1 := "Test1\n"
	record2 := "Test2\n"

	file, err := OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0640)
	if err != nil {
		t.Fatalf("Can't open file '%s' for writing: %s", file.Name(), err)
	}

	_, err = fmt.Fprintf(file, record1)
	if err != nil {
		t.Fatalf("Cant write to the file '%s': %s", file.Name(), err)
	}

	os.Remove(filePath)
	file.object.maintain()

	_, err = fmt.Fprintf(file, record2)
	if err != nil {
		t.Fatalf("Cant write to the file '%s': %s", file.Name(), err)
	}

	err = file.Close()
	if err != nil {
		t.Fatalf("Can't close file '%s': %s", file.Name(), err)
	}

	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Can't read from file '%s': %s", file.Name(), err)
	}

	if string(contents) != record2 {
		t.Errorf("File contents is different from expected:\n'%s' - Got\n'%s' - Expected", string(contents), record2)
	}
}
