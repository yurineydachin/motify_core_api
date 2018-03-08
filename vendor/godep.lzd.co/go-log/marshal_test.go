package log

import (
	"strconv"
	"testing"
)

//easyjson:json
type additionalDataStringString map[string]string
//easyjson:json
type additionalDataStringInterface map[string]interface{}

var mapStringStringSmall = make(additionalDataStringString)
var mapStringInterfaceSmall = make(additionalDataStringInterface)

var mapStringStringMedium = make(additionalDataStringString)
var mapStringInterfaceMedium = make(additionalDataStringInterface)

var mapStringStringLarge = make(additionalDataStringString)
var mapStringInterfaceLarge = make(additionalDataStringInterface)

func init() {
	for i := 0; i <= 3; i++ {
		mapStringStringSmall["key_" + strconv.Itoa(i)] = "value_" + strconv.Itoa(i)
		mapStringInterfaceSmall["key_" + strconv.Itoa(i)] = "value_" + strconv.Itoa(i)
	}
	for i := 0; i <= 50; i++ {
		mapStringStringMedium["key_" + strconv.Itoa(i)] = "value_" + strconv.Itoa(i)
		mapStringInterfaceMedium["key_" + strconv.Itoa(i)] = "value_" + strconv.Itoa(i)
	}
	for i := 0; i <= 1000; i++ {
		mapStringStringLarge["key_" + strconv.Itoa(i)] = "value_" + strconv.Itoa(i)
		mapStringInterfaceLarge["key_" + strconv.Itoa(i)] = "value_" + strconv.Itoa(i)
	}
}

func BenchmarkMarshalMapStringStringSmall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mapStringStringSmall.MarshalJSON()
	}
}

func BenchmarkMarshalMapStringInterfaceSmall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mapStringInterfaceSmall.MarshalJSON()
	}
}

func BenchmarkMarshalMapStringStringMedium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mapStringStringMedium.MarshalJSON()
	}
}

func BenchmarkMarshalMapStringInterfaceMedium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mapStringInterfaceMedium.MarshalJSON()
	}
}

func BenchmarkMarshalMapStringStringLarge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mapStringStringLarge.MarshalJSON()
	}
}

func BenchmarkMarshalMapStringInterfaceLarge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mapStringInterfaceLarge.MarshalJSON()
	}
}