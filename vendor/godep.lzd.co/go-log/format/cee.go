package format

import (
	"encoding/json"
	"io"
	"strconv"
)

type CEE struct {
	Data interface{}

	level Severity
}

func (f *CEE) SetLevel(s Severity) {
	f.level = s
}

func (f *CEE) WriteTo(w io.Writer) (int64, error) {
	var lvl int
	var n int64
	if f.level != nil {
		lvl = f.level.Code()
	}
	header := "<" + strconv.Itoa(facility*8+lvl) + "> @cee:"
	if m, err := io.WriteString(w, header); err != nil {
		return n + int64(m), err
	} else {
		n += int64(m)
	}
	if m, ok := f.Data.(json.Marshaler); ok {
		data, err := m.MarshalJSON()
		if err != nil {
			return n, err
		}
		if m, err := w.Write(data); err != nil {
			return n + int64(m), err
		} else {
			n += int64(m)
		}
	} else {
		data, err := json.Marshal(f.Data)
		if err != nil {
			return n, err
		}
		if m, err := w.Write(data); err != nil {
			return n + int64(m), err
		} else {
			n += int64(m)
		}
	}
	if m, err := io.WriteString(w, "\n"); err != nil {
		return n + int64(m), err
	} else {
		n += int64(m)
	}
	return n, nil
}
