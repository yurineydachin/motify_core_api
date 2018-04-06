package qrcode

import (
	"bytes"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"image/png"
)

const (
	defaultSize = 250
)

func Generate(magicCode string, size int) ([]byte, error) {
	code, err := qr.Encode(magicCode, qr.L, qr.Auto)
	if err != nil {
		return nil, err
	}

	if size <= 0 {
		size = defaultSize
	}

	code, err = barcode.Scale(code, size, size)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, code); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
