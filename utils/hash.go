package utils

import (
	"crypto/md5"
	"encoding/hex"
)

const salt = "mtf_some_rand_string_with_numbers31312342-8343"

func hash(value string) string {
	hasher := md5.New()
	hasher.Write([]byte(value + salt))
	return hex.EncodeToString(hasher.Sum(nil))
}

func HashLogin(value string) string {
	return hash("login_" + value)
}

func HashPass(value string) string {
	return hash("pass_" + value)
}
