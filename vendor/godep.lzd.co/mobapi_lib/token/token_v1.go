package token

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"godep.lzd.co/go-config"
	cryptoRand "crypto/rand"
	mathRand "math/rand"
)

const (
	// cartHashPrefix is prefix for mobile api cart hash
	cartHashPrefix = "MOB_"
)

var (
	// tripleDESChiper used for encrypter/decrypter creation
	tripleDESChiper cipher.Block
	// ciphertext used for creating Initialization Vector
	ciphertext []byte
	hashSalt   []byte

	tripleDESChiperObsolete, _ = des.NewTripleDESCipher([]byte("NG2q2jNRELpyeZjJcDUd3vj8"))
	ciphertextObsolete         = []byte("KkLSFAJpX49N2XYS")[:des.BlockSize]
	hashSaltObsolete           = []byte("KkLSFAJpX49N2XYS")
	isObsoleteTokenSupported   = false
)

// V1 is base struct for token data storing
type V1 struct {
	CustomerID int64
	IssuedAt   int64
	Checksum   [32]byte
}

// InitTokenV1 performs encryption initialization for tokens
func InitTokenV1(tripleKey, salt []byte) error {
	l := len(tripleKey)
	if l != 24 {
		return fmt.Errorf("Invalid token des key has provided! It should contain 24 bit, but %d-bit key is provided", l)
	}

	l = len(salt)
	if l < des.BlockSize {
		return fmt.Errorf("Invalid salt length! It should be at least %d bytes but %d is provided", des.BlockSize, l)
	}

	ciphertext = salt[:des.BlockSize]
	hashSalt = salt

	venture, _ := config.GetString("venture")
	env, _ := config.GetString("env")
	isObsoleteTokenSupported = env == "live" && (venture == "id" || venture == "my" || venture == "sg" || venture == "vn")

	var err error
	// tripleDESChiper is chiper block based on tripleKey used for encryption/decryption
	tripleDESChiper, err = des.NewTripleDESCipher(tripleKey)
	return err
}

func newTokenV1(customerID int) *V1 {
	token := &V1{
		CustomerID: int64(customerID),
		IssuedAt:   time.Now().UnixNano(),
	}
	var err error
	token.Checksum, err = token.calcCheckSum()
	if err != nil {
		// very unlikely
		panic(err)
	}
	return token
}

// NewTokenV1 creates new valid token struct by provided client ID
func NewTokenV1(customerID uint64) *V1 {
	return newTokenV1(int(customerID))
}

// NewGuestTokenV1 creates new token for guest user
func NewGuestTokenV1() *V1 {
	// generate random numbers in []byte
	rnd := RandomCreateBytes(8, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}...)
	// convert []byte ti uint64
	number := binary.BigEndian.Uint64(rnd)
	// make it negative and create new token
	return newTokenV1(int(number) * -1)
}

// GetTokenV1ByHash returns new V1 decoded from hash
func GetTokenV1ByHash(desHash []byte) (*V1, error) {
	token, err := getTokenV1ByHash(desHash, tripleDESChiper, ciphertext)
	if err != nil && isObsoleteTokenSupported {
		return getTokenV1ByHash(desHash, tripleDESChiperObsolete, ciphertextObsolete)
	}

	return token, err
}

func getTokenV1ByHash(desHash []byte, tripleDESChiper cipher.Block, ciphertext []byte) (*V1, error) {
	// create decrypter
	decrypter := cipher.NewCBCDecrypter(tripleDESChiper, ciphertext)

	cutted := desHash[1:] // remove first byte because it's just version number
	decrypted := make([]byte, len(cutted))
	decrypter.CryptBlocks(decrypted, cutted)

	// create struct by decrypted data
	token := &V1{}
	if err := binary.Read(bytes.NewBuffer(decrypted), binary.LittleEndian, token); err != nil {
		return nil, err
	}
	if !token.IsValid() {
		return nil, errors.New("Decrypted token is invalid")
	}
	return token, nil
}

// GetCartHashByTokenV1Hash returns short hash for cart identification (`cookie_id` field in perpetual_cart table into DB)
func GetCartHashByTokenV1Hash(base64String string) string {
	hash32 := sha256.Sum256([]byte(base64String))
	hash := cartHashPrefix + base64.StdEncoding.EncodeToString(hash32[:])
	hash = strings.TrimSuffix(hash, `=`)
	return hash
}

// GetCartHash getcarthash
func (token *V1) GetCartHash() string {
	return GetCartHashByTokenV1Hash(token.String())
}

// GetCustomerID getcustomerid
func (token *V1) GetCustomerID() uint64 {
	if token != nil && token.CustomerID > 0 {
		return uint64(token.CustomerID)
	}
	return uint64(0)
}

// GetDate getdate
func (token *V1) GetDate() time.Time {
	return time.Unix(token.IssuedAt, 0)
}

// String performs token encoding and returns hash in string
func (token *V1) String() string {
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, token); err != nil {
		return ""
	}
	b := buf.Bytes()

	encrypted := make([]byte, len(b)+1)
	encrypted[0] = byte(1)
	encrypter := cipher.NewCBCEncrypter(tripleDESChiper, ciphertext)
	encrypter.CryptBlocks(encrypted[1:], b)

	return base64.StdEncoding.EncodeToString(encrypted)
}

// calcCheckSum calculates checksum by token data
func (token *V1) calcCheckSum() ([32]byte, error) {
	return token.calcCheckSumValue(hashSalt)
}

func (token *V1) calcCheckSumObsolete() ([32]byte, error) {
	return token.calcCheckSumValue(hashSaltObsolete)
}

func (token *V1) calcCheckSumValue(hashSalt []byte) ([32]byte, error) {
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.BigEndian, token.IssuedAt); err != nil {
		var empty [32]byte
		return empty, err
	}
	buf.Write(hashSalt)
	salt := sha256.Sum256(buf.Bytes())
	buf.Reset()
	buf.Write(salt[:])
	if err := binary.Write(&buf, binary.LittleEndian, token.CustomerID); err != nil {
		var empty [32]byte
		return empty, err
	}
	if err := binary.Write(&buf, binary.LittleEndian, token.IssuedAt); err != nil {
		var empty [32]byte
		return empty, err
	}
	return sha256.Sum256(buf.Bytes()), nil
}

// IsValid performs checksum testing
func (token *V1) IsValid() bool {
	checksum, err := token.calcCheckSum()
	if token.isValid(checksum, err) {
		return true
	}

	if !isObsoleteTokenSupported {
		return false
	}

	checksum, err = token.calcCheckSumObsolete()
	return token.isValid(checksum, err)
}

func (token *V1) isValid(checksum [32]byte, err error) bool {
	if err != nil {
		return false
	}

	return subtle.ConstantTimeCompare(token.Checksum[:], checksum[:]) == 1
}

// IsGuest returns true for guest users
func (token *V1) IsGuest() bool {
	return token.GetCustomerID() == 0
}

// RandomCreateBytes generate random []byte by specify chars.
func RandomCreateBytes(n int, alphabets ...byte) []byte {
	const alphaNum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	var rand bool
	if num, err := cryptoRand.Read(bytes); num != n || err != nil {
		mathRand.Seed(time.Now().UnixNano())
		rand = true
	}
	for i, b := range bytes {
		if len(alphabets) == 0 {
			if rand {
				bytes[i] = alphaNum[mathRand.Intn(len(alphaNum))]
			} else {
				bytes[i] = alphaNum[b%byte(len(alphaNum))]
			}
		} else {
			if rand {
				bytes[i] = alphabets[mathRand.Intn(len(alphabets))]
			} else {
				bytes[i] = alphabets[b%byte(len(alphabets))]
			}
		}
	}
	return bytes
}
