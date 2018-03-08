// THIS FILE WAS MODIFIED AFTER GENERATION, DO NOT REGENERATE IT

package format

import (
	"encoding/json"
	"github.com/mailru/easyjson/jwriter"
	"fmt"
	"reflect"
)

func easyjsonB6915918EncodeGoLog(out *jwriter.Writer, in structuredData) {
	if in == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
		out.RawString(`null`)
	} else {
		out.RawByte('{')
		v2First := true
		for v2Name, v2Value := range in {
			if !v2First {
				out.RawByte(',')
			}
			v2First = false
			out.String(string(v2Name))
			out.RawByte(':')
			if v, ok := v2Value.(string); ok {
				out.String(v)
			} else if v, ok := v2Value.(int); ok {
				out.Int(v)
			} else if v, ok := v2Value.(int8); ok {
				out.Int8(v)
			} else if v, ok := v2Value.(int16); ok {
				out.Int16(v)
			} else if v, ok := v2Value.(int32); ok {
				out.Int32(v)
			} else if v, ok := v2Value.(int64); ok {
				out.Int64(v)
			} else if v, ok := v2Value.(uint); ok {
				out.Uint(v)
			} else if v, ok := v2Value.(uint8); ok {
				out.Uint8(v)
			} else if v, ok := v2Value.(uint16); ok {
				out.Uint16(v)
			} else if v, ok := v2Value.(uint32); ok {
				out.Uint32(v)
			} else if v, ok := v2Value.(uint64); ok {
				out.Uint64(v)
			} else if v, ok := v2Value.(float32); ok {
				out.Float32(v)
			} else if v, ok := v2Value.(float64); ok {
				out.Float64(v)
			} else if v, ok := v2Value.(bool); ok {
				out.Bool(v)
			} else if v, ok := v2Value.([]byte); ok {
				out.String(string(v))
			} else if m, ok := v2Value.(json.Marshaler); ok {
				if s, err := m.MarshalJSON(); err != nil {
					out.Error = err
				} else {
					out.String(string(s))
				}
			} else if m, ok := v2Value.(fmt.Stringer); ok {
				out.String(m.String())
			} else if reflect.TypeOf(v2Value).Kind() == reflect.String {
				out.String(reflect.ValueOf(v2Value).String())
			} else {
				if s, err := json.Marshal(v2Value); err != nil {
					out.Error = err
				} else {
					out.String(string(s))
				}
			}
		}
		out.RawByte('}')
	}
}

// MarshalJSON supports json.Marshaler interface
func (v structuredData) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonB6915918EncodeGoLog(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v structuredData) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonB6915918EncodeGoLog(w, v)
}
