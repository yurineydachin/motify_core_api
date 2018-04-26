package format

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"godep.lzd.co/go-log/internal"
)

const service = "service"

func TestMarshalAdditionalData(t *testing.T) {
	type s1 struct {
		M string
	}
	type s2 struct {
		N s1
	}
	type SpanKindEnum string

	tests := []struct {
		In  map[string]interface{}
		Out string
	}{
		{
			In: map[string]interface{}{
				"key": "value",
			},
			Out: `{"key":"value"}`,
		},
		{
			In: map[string]interface{}{
				"key": 1,
			},
			Out: `{"key":1}`,
		},
		{
			In: map[string]interface{}{
				"key": true,
			},
			Out: `{"key":true}`,
		},
		{
			In: map[string]interface{}{
				"key": []byte("value"),
			},
			Out: `{"key":"value"}`,
		},
		{
			In: map[string]interface{}{
				"key": s1{"value"},
			},
			Out: `{"key":"{\"M\":\"value\"}"}`,
		},
		{
			In: map[string]interface{}{
				"key": s2{s1{"value"}},
			},
			Out: `{"key":"{\"N\":{\"M\":\"value\"}}"}`,
		},
		{
			In: map[string]interface{}{
				"span.kind": SpanKindEnum("server"), // testing custom (aliased) string types
			},
			Out: `{"span.kind":"server"}`, // custom string types should be handled exactly like plain strings
		},
	}

	var marshal = func(data structuredData) (string, error) {
		b, e := data.MarshalJSON()
		return string(b), e
	}
	for _, c := range tests {
		r, e := marshal(c.In)
		if e != nil {
			t.Errorf("Error during marshal: %s", e)
		}
		if r != c.Out {
			t.Errorf("expected result is: '%s', got: '%s'", c.Out, r)
		}
	}
}

type debugMock struct{}

func (l debugMock) Code() int {
	return 7
}

func (l debugMock) String() string {
	return "DEBUG"
}

var debug = debugMock{}

func TestStd_WriteToForSyslog(t *testing.T) {
	buf := new(bytes.Buffer)
	f := &Std{
		CallStackSkip: 0,
		TraceId:       "3ymrswshj4sg",
		SpanId:        "6yaoivssj1rt",
		Message:       "test message with | and \n",
		Data:          map[string]interface{}{"attribute": "value with | and \n"},
	}
	f.EnableSyslogHeader(true)
	f.SetService(service)
	f.SetLevel(debug)
	f.WriteTo(buf)
	source := internal.CurrentLineMinusOne()
	pri := fmt.Sprintf("<%d> ", facility*8+debug.Code())
	parts := strings.Split(buf.String(), delimiter)
	i := 0
	if parts[i][:len(pri)] != pri {
		t.Errorf("expected start of string: '%s', got: '%s'\n", pri, string(parts[i][:len(pri)]))
	}
	if len(parts) != 13 {
		t.Errorf("expected 13 parts of log string, got: %d\n", len(parts))
	}
	if parts[i][len(pri):] != hostname {
		t.Errorf("expected hostname \"%s\", got: \"%s\"\n", hostname, parts[i][len(pri):])
	}
	i++
	ts, err := time.Parse(time.RFC3339Nano, parts[i])
	if err != nil {
		t.Errorf("expected timestamp format: \"2006-01-02T15:04:05.999999Z07:00\", got: \"%s\"\n",
			parts[i])
	}
	since := time.Since(ts)
	if since > time.Millisecond {
		t.Errorf("expected duration between now and timestamp: < 1 ms, got: %s\n", since)
	}
	if i++; parts[i] != "3ymrswshj4sg" {
		t.Errorf("expected TraceId \"3ymrswshj4sg\", got: \"%s\"\n", parts[i])
	}
	if i++; parts[i] != none {
		t.Errorf("expected ParentSpanId \"%s\", got: \"%s\"\n", none, parts[i])
	}
	if i++; parts[i] != "6yaoivssj1rt" {
		t.Errorf("expected SpanId \"6yaoivssj1rt\", got: \"%s\"\n", parts[i])
	}
	if i++; parts[i] != rolloutStable {
		t.Errorf("expected RolloutType \"%s\", got: \"%s\"\n", rolloutStable, parts[i])
	}
	if i++; parts[i] != service {
		t.Errorf("expected service name \"%s\", got: \"%s\"\n", service, parts[i])
	}
	if i++; parts[i] != debug.String() {
		t.Errorf("expected log level \"%s\", got: \"%s\"\n", debug.String(), parts[i])
	}
	if i++; !strings.HasSuffix(parts[i], "format") {
		t.Errorf("expected component_name suffix \"format\", got: \"%s\"\n", parts[i])
	}
	if i++; parts[i] != source {
		t.Errorf("expected file name and line \"%s\", got: \"%s\"\n", source, parts[i])
	}
	if i++; parts[i] != "test message with \\| and \\n" {
		t.Errorf("expected message \"test message with \\| and \\n\", got: \"%s\"\n", parts[i])
	}
	if i++; parts[i] != "{\"attribute\":\"value with \\| and \\n\"}" {
		t.Errorf("expected additional data {\"attribute\":\"value with \\| and \\n\"}, got: %s\n", parts[i])
	}
	if i++; parts[i] != ".\n" {
		t.Errorf("expected '.\\n' as end of string, got: '%s'\n", parts[i])
	}
}

func TestStd_WriteToForStdout(t *testing.T) {
	buf := new(bytes.Buffer)
	f := &Std{}
	f.SetService(service)
	f.SetLevel(debug)
	f.WriteTo(buf)
	source := internal.CurrentLineMinusOne()
	parts := strings.Split(buf.String(), delimiter)
	i := 0
	if len(parts) != 13 {
		t.Errorf("expected 13 parts of log string, got: %d\n", len(parts))
	}
	if parts[i] != hostname {
		t.Errorf("expected hostname \"%s\", got: \"%s\"\n", hostname, parts[i])
	}
	i++
	ts, err := time.Parse(time.RFC3339Nano, parts[i])
	if err != nil {
		t.Errorf("expected timestamp format: \"2006-01-02T15:04:05.999999Z07:00\", got: \"%s\"\n",
			parts[i])
	}
	since := time.Since(ts)
	if since > time.Millisecond {
		t.Errorf("expected duration between now and timestamp: < 1 ms, got: %s\n", since)
	}
	if i++; parts[i] != none {
		t.Errorf("expected TraceId \"%s\", got: \"%s\"\n", none, parts[i])
	}
	if i++; parts[i] != none {
		t.Errorf("expected ParentSpanId \"%s\", got: \"%s\"\n", none, parts[i])
	}
	if i++; parts[i] != none {
		t.Errorf("expected SpanId \"%s\", got: \"%s\"\n", none, parts[i])
	}
	if i++; parts[i] != rolloutStable {
		t.Errorf("expected RolloutType \"%s\", got: \"%s\"\n", rolloutStable, parts[i])
	}
	if i++; parts[i] != service {
		t.Errorf("expected service name \"%s\", got: \"%s\"\n", service, parts[i])
	}
	if i++; parts[i] != debug.String() {
		t.Errorf("expected log level \"%s\", got: \"%s\"\n", debug.String(), parts[i])
	}
	if i++; !strings.HasSuffix(parts[i], "format") {
		t.Errorf("expected component_name suffix \"format\", got: \"%s\"\n", parts[i])
	}
	if i++; parts[i] != source {
		t.Errorf("expected file name and line \"%s\", got: \"%s\"\n", source, parts[i])
	}
	if i++; parts[i] != none {
		t.Errorf("expected message \"%s\", got: \"%s\"\n", none, parts[i])
	}
	if i++; parts[i] != "{}" {
		t.Errorf("expected additional data {}, got: %s\n", parts[i])
	}
	if i++; parts[i] != ".\n" {
		t.Errorf("expected '.\\n' as end of string, got: '%s'\n", parts[i])
	}
}

func TestTruncate(t *testing.T) {
	orig_max_buffer_size := max_buffer_size
	defer func() {
		max_buffer_size = orig_max_buffer_size
	}()
	buf := new(bytes.Buffer)
	f := &Std{
		TraceId: "3ymrswshj4sg",
		SpanId:  "6yaoivssj1rt",
		Message: "test message",
		Data:    map[string]interface{}{"attribute": "value"},
	}
	f.WriteTo(buf)
	parts := strings.Split(buf.String(), delimiter)
	if len(parts) != 13 {
		t.Errorf("expected 13 parts of log string, got: %d\n", len(parts))
	}
	if parts[12] != ".\n" {
		t.Errorf("expected '.\\n' as end of string, got: '%s'\n", parts[11])
	}
	for i := 50; i <= 143+len(hostname)+len(parts[8])+3; i++ {
		max_buffer_size = i
		buf := new(bytes.Buffer)
		f := &Std{
			TraceId: "3ymrswshj4sg",
			SpanId:  "6yaoivssj1rt",
			Message: "test message",
			Data:    map[string]interface{}{"attribute": "value"},
		}
		f.WriteTo(buf)
		if buf.Len() != i {
			t.Errorf("expected len of buf %d, got: %d\n",
				i, buf.Len())
		}
		parts := strings.Split(buf.String(), delimiter)
		if len(parts) != 13 {
			t.Errorf("expected 13 parts of log string, got: %d\n", len(parts))
		}
		if parts[12] != "X\n" {
			t.Errorf("expected 'X\\n' as end of string, got: '%s'\n", parts[11])
		}
	}
}

func TestStd_StackWithoutPackagesSkip(t *testing.T) {
	buf := new(bytes.Buffer)
	f := &Std{
		CallStackSkip: 0,
		TraceId:       "3ymrswshj4sg",
		SpanId:        "6yaoivssj1rt",
		Message:       "test message with | and \n",
		Data:          map[string]interface{}{"attribute": "value with | and \n"},
	}
	f.WriteTo(buf)

	name, filename, _ := f.getCaller(0)
	if !strings.Contains(name, "getCaller") {
		t.Errorf("Expected getCaller function name, got %s", name)
	}
	if !strings.Contains(filename, "std.go") {
		t.Errorf("Expected std.go filename, got %s", filename)
	}
}

func TestStd_StackWithPackagesSkip(t *testing.T) {
	buf := new(bytes.Buffer)
	f := &Std{
		CallStackSkip: 2,
		TraceId:       "3ymrswshj4sg",
		SpanId:        "6yaoivssj1rt",
		Message:       "test message with | and \n",
		Data:          map[string]interface{}{"attribute": "value with | and \n"},
	}
	f.SetBacktraceSkips([]string{"testing"})
	f.WriteTo(buf)

	name, filename, _ := f.getCaller(1)
	if name != "runtime.goexit" {
		t.Errorf("Expected runtime.goexit function name, got %s", name)
	}
	if !strings.Contains(filename, "runtime") {
		t.Errorf("Expected runtime package, got %s", filename)
	}
}

type noop struct {
}

func (noop) Write(b []byte) (int, error) {
	return len(b), nil
}

func BenchmarkStdFormatter(b *testing.B) {
	f := &Std{
		CallStackSkip: 0,
		TraceId:       "3ymrswshj4sg",
		SpanId:        "6yaoivssj1rt",
		Message:       "test message with | and \n",
		Data:          map[string]interface{}{"attribute": "value with | and \n"},
	}
	f.EnableSyslogHeader(true)
	f.SetService(service)
	f.SetLevel(debug)
	b.ResetTimer()
	n := &noop{}
	for i := 0; i < b.N; i++ {
		f.Message = "test message with | and \n"
		_, err := f.WriteTo(n)
		if err != nil {
			b.FailNow()
		}
	}
}

func BenchmarkStdFormatterWithIgnore(b *testing.B) {
	f := &Std{
		CallStackSkip: 0,
		TraceId:       "3ymrswshj4sg",
		SpanId:        "6yaoivssj1rt",
		Message:       "test message with | and \n",
		Data:          map[string]interface{}{"attribute": "value with | and \n"},
	}
	f.EnableSyslogHeader(true)
	f.SetService(service)
	f.SetLevel(debug)
	f.SetBacktraceSkips([]string{"testing"})
	b.ResetTimer()
	n := &noop{}
	for i := 0; i < b.N; i++ {
		f.Message = "test message with | and \n"
		_, err := f.WriteTo(n)
		if err != nil {
			b.FailNow()
		}
	}
}
