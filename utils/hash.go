package utils

import (
	"crypto/md5"
	"encoding/hex"
)

const salt = "mtf_some_rand_string_with_numbers31312342-8343"

func Hash(value string) string {
	hasher := md5.New()
	hasher.Write([]byte(value + salt))
	return hex.EncodeToString(hasher.Sum(nil))
}
