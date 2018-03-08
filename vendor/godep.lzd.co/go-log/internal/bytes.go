package internal

import (
	"reflect"
	"unsafe"
)

func BytesToString(b *[]byte) string {
	return *(*string)(unsafe.Pointer(b))
}

func StringToBytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

// may panic if out is less then needed
func IntToString(in int, out []byte) int {
	var stack [25]byte

	i := 0
	if in < 0 {
		stack[i] = '-'
		in = -in
		i++
	}

	if in == 0 {
		stack[i] = '0'
		i++
		goto out
	}

	for ; in > 0; i++ {
		x := in % 10
		stack[i] = byte(int('0') + x)
		in /= 10
	}

out:
	j := 0
	for p := i - 1; p >= 0; p-- {
		out[j] = stack[p]
		j++
	}
	return i
}

func IntToStringAppend(in int, out []byte) []byte {
	var stack [25]byte
	sl := MakeSliceWithData(unsafe.Pointer(&stack), 25, 25)
	n := IntToString(in, sl)
	return append(out, MakeSliceWithData(unsafe.Pointer(&stack), n, n)...)
}

func MakeSliceWithData(data unsafe.Pointer, len, cap int) []byte {
	bh := reflect.SliceHeader{
		Data: uintptr(data),
		Len:  len,
		Cap:  cap,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}
