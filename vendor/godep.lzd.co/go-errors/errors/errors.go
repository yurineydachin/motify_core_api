// Package errors module implements functions which manipulate errors and provide stack
// trace information.
//
// NOTE: This package intentionally mirrors the standard "errors" module.
// All dropbox code should use this.
package errors

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// IDropboxError interface exposes additional information about the error.
type IDropboxError interface {
	// GetMessage returns the error message without the stack trace.
	GetMessage() string

	// GetStack returns the stack trace without the error message.
	GetStack() string

	// GetContext returns the stack trace's context.
	GetContext() string

	// GetInner returns the wrapped error.  This returns nil if this does not wrap
	// another error.
	GetInner() error

	// Error implements the built-in error interface.
	Error() string
}

// DropboxBaseError is a standard struct for general types of errors.
//
// For an example of custom error type, look at databaseError/newDatabaseError
// in errors_test.go.
type DropboxBaseError struct {
	Msg     string
	Stack   string
	Context string
	inner   error
}

// GetMessage returns the error string without stack trace information.
func GetMessage(err interface{}) string {
	switch e := err.(type) {
	case IDropboxError:
		dberr := IDropboxError(e)
		ret := []string{}
		for dberr != nil {
			ret = append(ret, dberr.GetMessage())
			d := dberr.GetInner()
			if d == nil {
				break
			}
			var ok bool
			dberr, ok = d.(IDropboxError)
			if !ok {
				ret = append(ret, d.Error())
				break
			}
		}
		return strings.Join(ret, " ")
	case runtime.Error:
		return runtime.Error(e).Error()
	default:
		return "Passed a non-error to GetMessage"
	}
}

// Error returns a string with all available error information, including inner
// errors that are wrapped by this errors.
func (e *DropboxBaseError) Error() string {
	return DefaultError(e)
}

// GetMessage returns the error message without the stack trace.
func (e *DropboxBaseError) GetMessage() string {
	return e.Msg
}

// GetStack returns the stack trace without the error message.
func (e *DropboxBaseError) GetStack() string {
	return e.Stack
}

// GetContext returns the stack trace's context.
func (e *DropboxBaseError) GetContext() string {
	return e.Context
}

// GetInner returns the wrapped error, if there is one.
func (e *DropboxBaseError) GetInner() error {
	return e.inner
}

// New returns a new DropboxBaseError initialized with the given message and
// the current stack trace.
func New(msg string) IDropboxError {
	stack, context := StackTrace()
	return &DropboxBaseError{
		Msg:     msg,
		Stack:   stack,
		Context: context,
	}
}

// Newf same as New, but with fmt.Printf-style parameters.
func Newf(format string, args ...interface{}) IDropboxError {
	stack, context := StackTrace()
	return &DropboxBaseError{
		Msg:     fmt.Sprintf(format, args...),
		Stack:   stack,
		Context: context,
	}
}

// Wrap wraps another error in a new DropboxBaseError.
func Wrap(err error, msg string) IDropboxError {
	stack, context := StackTrace()
	return &DropboxBaseError{
		Msg:     msg,
		Stack:   stack,
		Context: context,
		inner:   err,
	}
}

// Wrapf same as Wrap, but with fmt.Printf-style parameters.
func Wrapf(err error, format string, args ...interface{}) IDropboxError {
	stack, context := StackTrace()
	return &DropboxBaseError{
		Msg:     fmt.Sprintf(format, args...),
		Stack:   stack,
		Context: context,
		inner:   err,
	}
}

// DefaultError default implementation of the Error method of the error interface.
func DefaultError(e IDropboxError) string {
	// Find the "original" stack trace, which is probably the most helpful for
	// debugging.
	errLines := make([]string, 1)
	var origStack string
	errLines[0] = "ERROR:"
	fillErrorInfo(e, &errLines, &origStack)
	errLines = append(errLines, "")
	errLines = append(errLines, "ORIGINAL STACK TRACE:")
	errLines = append(errLines, origStack)
	return strings.Join(errLines, "\n")
}

// Fills errLines with all error messages, and origStack with the inner-most
// stack.
func fillErrorInfo(err error, errLines *[]string, origStack *string) {
	if err == nil {
		return
	}

	derr, ok := err.(IDropboxError)
	if ok {
		*errLines = append(*errLines, derr.GetMessage())
		*origStack = derr.GetStack()
		fillErrorInfo(derr.GetInner(), errLines, origStack)
	} else {
		*errLines = append(*errLines, err.Error())
	}
}

// Returns a copy of the error with the stack trace field populated and any
// other shared initialization; skips 'skip' levels of the stack trace.
//
// NOTE: This panics on any error.
func stackTrace(skip int) (current, context string) {
	// grow buf until it's large enough to store entire stack trace
	buf := make([]byte, 128)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, len(buf)*2)
	}

	// Returns the index of the first occurrence of '\n' in the buffer 'b'
	// starting with index 'start'.
	//
	// In case no occurrence of '\n' is found, it returns len(b). This
	// simplifies the logic on the calling sites.
	indexNewline := func(b []byte, start int) int {
		if start >= len(b) {
			return len(b)
		}
		searchBuf := b[start:]
		index := bytes.IndexByte(searchBuf, '\n')
		if index == -1 {
			return len(b)
		}
		return (start + index)
	}

	// Strip initial levels of stack trace, but keep header line that
	// identifies the current goroutine.
	var strippedBuf bytes.Buffer
	index := indexNewline(buf, 0)
	if index != -1 {
		strippedBuf.Write(buf[:index])
	}

	// Skip lines.
	for i := 0; i < skip; i++ {
		index = indexNewline(buf, index+1)
		index = indexNewline(buf, index+1)
	}

	isDone := false
	startIndex := index
	lastIndex := index
	for !isDone {
		index = indexNewline(buf, index+1)
		if (index - lastIndex) <= 1 {
			isDone = true
		} else {
			lastIndex = index
		}
	}
	strippedBuf.Write(buf[startIndex:index])
	return strippedBuf.String(), string(buf[index:])
}

// StackTrace returns the current stack trace string.  NOTE: the stack creation code
// is excluded from the stack trace.
func StackTrace() (current, context string) {
	return stackTrace(3)
}

// Return a wrapped error or nil if there is none.
func unwrapError(ierr error) (nerr error) {
	// Internal errors have a well defined bit of context.
	if dbxErr, ok := ierr.(IDropboxError); ok {
		return dbxErr.GetInner()
	}

	// At this point, if anything goes wrong, just return nil.
	defer func() {
		if x := recover(); x != nil {
			nerr = nil
		}
	}()

	// Go system errors have a convention but paradoxically no
	// interface.  All of these panic on error.
	errV := reflect.ValueOf(ierr).Elem()
	errV = errV.FieldByName("Err")
	return errV.Interface().(error)
}

// RootError keep peeling away layers or context until a primitive error is revealed.
func RootError(ierr error) (nerr error) {
	nerr = ierr
	for i := 0; i < 20; i++ {
		terr := unwrapError(nerr)
		if terr == nil {
			return nerr
		}
		nerr = terr
	}
	return fmt.Errorf("too many iterations: %T", nerr)
}

// IsError perform a deep check, unwrapping errors as much as possilbe and
// comparing the string version of the error.
func IsError(err, errConst error) bool {
	if err == errConst {
		return true
	}
	// Must rely on string equivalence, otherwise a value is not equal
	// to its pointer value.
	rootErrStr := ""
	rootErr := RootError(err)
	if rootErr != nil {
		rootErrStr = rootErr.Error()
	}
	errConstStr := ""
	if errConst != nil {
		errConstStr = errConst.Error()
	}
	return rootErrStr == errConstStr
}
