package token

import (
	"encoding/base64"
	"fmt"
	"time"
)

// IToken itoken
type IToken interface {
	IsValid() bool
	IsGuest() bool
	GetID() uint64
	GetModel() uint64
	GetExtraID() uint64
	GetDate() time.Time
	IsFixed() bool
	Fixed()
	String() string // return base64 string (1st byte - version, all other bytes - encoded data)
	GetCartHash() string

	calcCheckSum() ([32]byte, error)
}

// IGuestToken iguesttoken
type IGuestToken interface {
	IToken
}

// INullToken inulltoken
type INullToken interface {
	IToken
}

// ParseToken parsetoken
func ParseToken(base64String string) (IToken, error) { // 1st byte - version, later encoded data
	byteHash, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return nil, err
	}

	version := getVersion(byteHash)
	switch version {
	case 1:
		return GetTokenV1ByHash(byteHash)
	default:
		return nil, fmt.Errorf("failed to parse token. Unknown token version number %d", version)
	}
}

// GetCartHashByTokenHash getcarthashbytokenhash
func GetCartHashByTokenHash(base64String string) string {
	byteHash, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return ""
	}

	switch getVersion(byteHash) {
	case 1:
		return GetCartHashByTokenV1Hash(base64String)
	default:
		return ""
	}
}

// getVersion getversion
func getVersion(byteHash []byte) int {
	if len(byteHash) == 0 {
		return 0
	}
	// first byte is token version number
	return int(byteHash[0])
}
