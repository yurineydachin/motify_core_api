package format

import (
	"io"
	"os"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"bytes"

	"github.com/mailru/easyjson/jwriter"
	"motify_core_api/godep_libs/go-log/internal"
)

const NO_STACK_TRACE_INFO = -1

const facility = 18 // local use 2 (local2) from rfc5424
const delimiter = " | "
const none = "-"
const rolloutStable = "stable"

var hostname, _ = os.Hostname()

var max_buffer_size int = 128 * 1024

type Severity interface {
	Code() int
	String() string
}

type Std struct {
	CallStackSkip int
	TraceId       string
	SpanId        string
	ParentSpanId  string
	RolloutType   string
	Message       string
	Data          map[string]interface{}

	service          string
	syslogHeader     bool
	level            Severity
	extCallStackSkip int
	packagesSkip     []string
}

func (f *Std) SetLevel(s Severity) {
	f.level = s
}

func (f *Std) EnableSyslogHeader(s bool) {
	f.syslogHeader = s
}

func (f *Std) SetService(s string) {
	f.service = s
}

func (f *Std) SetBacktraceSkips(packages []string) {
	f.packagesSkip = packages
}

func (f *Std) IncExtCallStackSkip(d int) {
	f.extCallStackSkip += d
}

func (f *Std) WriteTo(out io.Writer) (int64, error) {
	parts := 12
	capacity := max_buffer_size -
		len(" | .\n") - //  end of string
		(parts-1)*len(delimiter) - // parts separator " | "
		parts // min 1 char of content for each part of log string

	f.writeSyslogHeader(out, &capacity)
	f.writeHostname(out, &capacity)
	f.writeTime(out, &capacity)
	f.writeTraceID(out, &capacity)
	f.writeParentSpanID(out, &capacity)
	f.writeSpanID(out, &capacity)
	f.writeRolloutType(out, &capacity)
	f.writeService(out, &capacity)
	f.writeLevelString(out, &capacity)
	f.writeStackTrace(out, &capacity)
	f.writeMessage(out, &capacity)
	f.writeData(out, &capacity)
	f.writeEnd(out, &capacity)

	return 0, nil
}

// PRI part of syslog message: "<%d> "
// Always writes but needs to decrement capacity
func (f *Std) writeSyslogHeader(out io.Writer, cap *int) {
	if !f.syslogHeader {
		return
	}

	levelCode := 0
	if f.level != nil {
		levelCode = f.level.Code()
	}

	var stack [25]byte
	stack[0] = '<'
	n := internal.IntToString(facility*8+levelCode, stack[1:])
	stack[n+1] = '>'
	stack[n+2] = ' '
	*cap -= n + 3
	writeBytes(out, nil, internal.MakeSliceWithData(unsafe.Pointer(&stack), n+3, 25))
}

func (f *Std) writeHostname(out io.Writer, cap *int) {
	writeString(out, cap, hostname, "")
	writeString(out, nil, delimiter, "")
}

func (f *Std) writeTime(out io.Writer, cap *int) {
	var stack [64]byte
	stackSlice := internal.MakeSliceWithData(unsafe.Pointer(&stack), 0, 64)
	stackSlice = time.Now().UTC().AppendFormat(stackSlice, time.RFC3339Nano)
	writeBytes(out, cap, stackSlice)
	writeString(out, nil, delimiter, "")
}

func (f *Std) writeTraceID(out io.Writer, cap *int) {
	writeString(out, cap, f.TraceId, none)
	writeString(out, nil, delimiter, "")
}

func (f *Std) writeParentSpanID(out io.Writer, cap *int) {
	writeString(out, cap, f.ParentSpanId, none)
	writeString(out, nil, delimiter, "")
}

func (f *Std) writeSpanID(out io.Writer, cap *int) {
	writeString(out, cap, f.SpanId, none)
	writeString(out, nil, delimiter, "")
}

func (f *Std) writeRolloutType(out io.Writer, cap *int) {
	writeString(out, cap, f.RolloutType, rolloutStable)
	writeString(out, nil, delimiter, "")
}

func (f *Std) writeService(out io.Writer, cap *int) {
	writeString(out, cap, f.service, none)
	writeString(out, nil, delimiter, "")
}

func (f *Std) writeLevelString(out io.Writer, cap *int) {
	levelString := ""
	if f.level != nil {
		levelString = f.level.String()
	}
	writeString(out, cap, levelString, none)
	writeString(out, nil, delimiter, "")
}

func (f *Std) writeStackTrace(out io.Writer, cap *int) {
	var stack [64]byte
	stackSlice := internal.MakeSliceWithData(unsafe.Pointer(&stack), 0, 64)
	funcName, file, line := f.getCaller(3)

	var componentName string
	if f.CallStackSkip != NO_STACK_TRACE_INFO {
		stackSlice = stackSlice[:0]
		i1 := strings.LastIndex(funcName, ".")
		i2 := strings.LastIndex(funcName, ".(")
		stackSlice = append(stackSlice, funcName[:minPositive(i1, i2)]...)
		componentName = internal.BytesToString(&stackSlice)
	}
	writeString(out, cap, componentName, none)
	writeString(out, nil, delimiter, "")

	var fileNameLine string
	if f.CallStackSkip != NO_STACK_TRACE_INFO {
		stackSlice = stackSlice[:0]
		stackSlice = append(stackSlice, file[strings.LastIndex(file, "/")+1:]...)
		stackSlice = append(stackSlice, ":"...)
		stackSlice = internal.IntToStringAppend(line, stackSlice)
		fileNameLine = internal.BytesToString(&stackSlice)
	}

	writeString(out, cap, fileNameLine, none)
	writeString(out, nil, delimiter, "")

	f.extCallStackSkip = 0

}

func (f *Std) getCaller(initSkip int) (name, file string, line int) {
	if len(f.packagesSkip) == 0 {
		pc, file, line, _ := runtime.Caller(initSkip + f.CallStackSkip + f.extCallStackSkip)
		name := runtime.FuncForPC(pc).Name()
		return name, file, line
	}

	// Caller skip is not the same as in Callers
	initSkip = initSkip + 1
	pc := make([]uintptr, 30)
	cnt := runtime.Callers(initSkip+f.CallStackSkip+f.extCallStackSkip, pc)
	if cnt == 0 {
		return "", "", 0
	}

	for d := 0; d < cnt; d++ {
		ptr := pc[d]
		fu := runtime.FuncForPC(ptr)

		funcName := fu.Name()
		funcPackage := extractPackageName(funcName)
		if isSkipped(funcPackage, f.packagesSkip) {
			continue
		}

		file, line = fu.FileLine(ptr)
		return funcName, file, line
	}

	ptr := pc[0]
	fu := runtime.FuncForPC(ptr)
	file, line = fu.FileLine(ptr)
	return fu.Name(), file, line
}

func extractPackageName(name string) string {
	i1 := strings.LastIndex(name, ".")
	i2 := strings.LastIndex(name, ".(")
	i3 := strings.Index(name, "/vendor/")
	if i3 > 0 {
		i3 = i3 + 8
	}
	return name[maxPositive(0, i3):minPositive(i1, i2)]
}

func isSkipped(packageName string, skipped []string) bool {
	for _, name := range skipped {
		if packageName == name {
			return true
		}
	}
	return false
}

func minPositive(a, b int) int {
	if b < 0 {
		if a < 0 {
			return 0
		}
		return a
	}
	if a <= b {
		return a
	}
	return b
}

func maxPositive(a, b int) int {
	if b < 0 {
		if a < 0 {
			return 0
		}
		return a
	}

	if a >= b {
		return a
	}
	return b
}

func (f *Std) writeMessage(out io.Writer, cap *int) {
	if f.Message == "" {
		writeString(out, cap, f.Message, none)
		writeString(out, nil, delimiter, "")
		return
	}
	b := internal.StringToBytes(f.Message)
	w := 0
	for i := 0; i < len(b) && *cap >= 0; i++ {
		switch b[i] {
		case '\\':
			writeBytes(out, cap, b[w:i])
			writeBytes(out, cap, internal.StringToBytes("\\"))
			w = i + 1
		case '|':
			writeBytes(out, cap, b[w:i])
			writeBytes(out, cap, internal.StringToBytes(`\|`))
			w = i + 1
		case '\r':
			writeBytes(out, cap, b[w:i])
			writeBytes(out, cap, internal.StringToBytes("\\r"))
			w = i + 1
		case '\n':
			writeBytes(out, cap, b[w:i])
			writeBytes(out, cap, internal.StringToBytes("\\n"))
			w = i + 1
		}
	}
	writeBytes(out, cap, b[w:])
	writeString(out, nil, delimiter, "")
}

func (f *Std) writeData(out io.Writer, cap *int) {
	if f.Data != nil {
		var data structuredData = f.Data
		w := jwriter.Writer{}
		w.Buffer.EnsureSpace(1000)
		data.MarshalEasyJSON(&w)
		b, _ := w.BuildBytes()
		b = bytes.Replace(b, []byte(delimiter), []byte(` \| `), -1)
		writeBytes(out, cap, b)
	} else {
		writeString(out, cap, "{}", "")
	}
}

func (f *Std) writeEnd(out io.Writer, cap *int) {
	writeString(out, nil, delimiter, "")
	if *cap <= 0 {
		writeString(out, nil, "X\n", "")
	} else {
		writeString(out, nil, ".\n", "")
	}
}

//easyjson:json
type structuredData map[string]interface{}

func writeBytes(out io.Writer, cap *int, data []byte) (int, error) {
	if cap != nil && *cap >= 0 && len(data) > *cap+1 {
		data = data[:*cap+1]
	}

	n, err := out.Write(data)
	if err != nil {
		return 0, err
	}

	if cap != nil {
		*cap -= n - 1
	}

	return n, nil
}

func writeString(out io.Writer, cap *int, str, replacement string) (int, error) {
	if str == "" {
		str = replacement
	}
	return writeBytes(out, cap, internal.StringToBytes(str))
}
