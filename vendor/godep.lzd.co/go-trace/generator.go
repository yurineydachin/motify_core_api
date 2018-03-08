package gotrace

import (
	"godep.lzd.co/go-trace/flake"
)

var (
	f         *flake.Flake
	generator IGenerator = &SimpleGenerator{}
)

func init() {
	var err error
	f, err = flake.WithRandomID()
	if err != nil {
		panic(err)
	}
}

// IGenerator generates unique string values for using as TraceID and SpanID
type IGenerator interface {
	Generate() string
}

// SimpleGenerator very simple generator implementation
type SimpleGenerator struct{}

// Generate generates random int64 value
func (s *SimpleGenerator) Generate() string {
	return f.NextID().String()
}

// SetGenerator sets current generator
func SetGenerator(g IGenerator) {
	generator = g
}
