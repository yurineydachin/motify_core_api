package format

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCEE_WriteTo(t *testing.T) {
	buf := new(bytes.Buffer)
	f := &CEE{
		Data: map[string]interface{}{"attribute": "value with\nnew line"},
	}
	f.SetLevel(debug)
	f.WriteTo(buf)
	header := fmt.Sprintf("<%d> @cee:", facility*8+debug.Code())
	if buf.String()[:len(header)] != header {
		t.Errorf("expected start of string: \"%s\", got: \"%s\"\n", header, buf.String()[:len(header)])
	}
	if buf.String()[len(header):buf.Len()-1] != "{\"attribute\":\"value with\\nnew line\"}" {
		t.Errorf("expected message {\"attribute\":\"value with\\nnew line\"}, got: %s\n", buf.String()[len(header):])
	}
	if buf.String()[buf.Len()-1:] != "\n" {
		t.Errorf("expected \\n as end of string, got: '%s'\n", buf.String()[buf.Len():])
	}
}
