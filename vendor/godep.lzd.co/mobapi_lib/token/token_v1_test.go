package token

import (
	"errors"
	"testing"

	"godep.lzd.co/awg"
	"godep.lzd.co/go-config"
)

func init() {
	config.RegisterString("env", "Environment", "live")
	config.RegisterString("venture", "Venture", "my")

	err := InitTokenV1([]byte("m2ZdwUFT3w7TkggfuTm7M4r8"), []byte("744Ve4KjjA9LMDEw"))
	if err != nil {
		panic(err)
	}
}

func TestObsoleteToken(t *testing.T) {
	customerID := uint64(3004272)
	hash := "AdjIgPtFgm1WdAdr0n376eWCKMcunYOl9S7krcNkQj+IqE8N2bi0D8DtGWHt01sUVw=="

	decodedToken, err := ParseToken(hash)
	if err != nil {
		t.Errorf("GetTokenV1ByHash error: %s", err)
	}

	// compare all fields
	if decodedToken.GetCustomerID() != customerID {
		t.Errorf("Token customerID does not match. expected: %d, decoded: %d", customerID, decodedToken.GetCustomerID())
	}
}

// TestTokenV1EncodingAndDecoding tries to encode API Token and decode new one from hash same as source one
func TestTokenV1EncodingAndDecoding(t *testing.T) {
	customerID := uint64(8763)
	originToken := NewTokenV1(customerID)

	// encode new hash by token struct
	hash := originToken.String()

	// get decoded token struct by given hash
	decodedToken, err := ParseToken(hash)
	if err != nil {
		t.Errorf("GetTokenV1ByHash error: %s", err)
	}

	// compare all fields
	if originToken.GetCustomerID() != uint64(customerID) {
		t.Errorf("Token customerID does not match. Original: %d, encoded: %d", customerID, originToken.GetCustomerID())
	}
	if originToken.GetCustomerID() != decodedToken.GetCustomerID() {
		t.Errorf("Token customerID does not match. Encoded: %d, decoded: %d", originToken.GetCustomerID(), decodedToken.GetCustomerID())
	}
	if originToken.GetDate() != decodedToken.GetDate() {
		t.Errorf("Token datetime does not match. Encoded: %s, decoded: %s", originToken.GetDate(), decodedToken.GetDate())
	}

	// and test decoding same hash again to escape wrong second decryption
	decodedToken, err = ParseToken(hash)
	if err != nil {
		t.Errorf("GetTokenV1ByHash error: %s", err)
	}

	// check cart hash decoding for non-empty result
	if GetCartHashByTokenHash(hash) == "" {
		t.Error("GetCartHashByTokenHash error: empty result")
	}
}

// TestGuestTokenV1ForUniqueness tries to create a lot of tokens in short time and checks it for uniqueness
func TestGuestTokenV1ForUniqueness(t *testing.T) {
	tries := 500
	result1 := make(map[string]bool, tries)
	result2 := make(map[string]bool, tries)

	var wg awg.AdvancedWaitGroup

	wg.Add(func() error {
		for i := 0; i < tries; i++ {
			hash := NewGuestTokenV1().String()
			if _, exist := result1[hash]; exist {
				return errors.New("Guest token duplicate was found")
			}
			result1[hash] = true
		}
		return nil
	})
	wg.Add(func() error {
		for i := 0; i < tries; i++ {
			hash := NewGuestTokenV1().String()
			if _, exist := result2[hash]; exist {
				return errors.New("Guest token duplicate was found")
			}
			result2[hash] = true
		}
		return nil
	})

	wg.SetStopOnError(true).Start()

	if wg.GetLastError() != nil {
		t.Error(wg.GetLastError().Error())
	}

	for hash1 := range result1 {
		for hash2 := range result2 {
			if hash1 == hash2 {
				t.Errorf("Token duplicate! Hash is '%s'", hash1)
			}
		}
	}
}

// TestGuestToken checks guest token behaviour
func TestGuestToken(t *testing.T) {
	token := NewGuestTokenV1()
	if !token.IsGuest() {
		t.Error("Token should be guest")
	}
	if token.GetCustomerID() > 0 {
		t.Error("Token should have zero customer ID")
	}
}
