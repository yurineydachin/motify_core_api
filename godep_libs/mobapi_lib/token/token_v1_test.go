package token

import (
	"errors"
	"testing"

	"motify_core_api/godep_libs/awg"
	"motify_core_api/godep_libs/go-config"
)

func init() {
	config.RegisterString("env", "Environment", "live")
	config.RegisterString("venture", "Venture", "my")

	err := InitTokenV1([]byte("m2ZdwUFT3w7TkggfuTm7M4r8"), []byte("744Ve4KjjA9LMDEw"))
	if err != nil {
		panic(err)
	}
}

// TestTokenV1EncodingAndDecoding tries to encode API Token and decode new one from hash same as source one
func TestTokenV1EncodingAndDecoding(t *testing.T) {
	ID := uint64(3004272)
	model := uint64(2)
	extraID := uint64(2345)
	originToken := NewTokenV1(ID, model, extraID)

	// encode new hash by token struct
	hash := originToken.String()

	// get decoded token struct by given hash
	decodedToken, err := ParseToken(hash)
	if err != nil {
		t.Errorf("GetTokenV1ByHash error: %s", err)
	}

	// compare all fields
	if originToken.GetID() != uint64(ID) {
		t.Errorf("Token ID does not match. Original: %d, encoded: %d", ID, originToken.GetID())
	}
	if originToken.GetID() != decodedToken.GetID() {
		t.Errorf("Token ID does not match. Encoded: %d, decoded: %d", originToken.GetID(), decodedToken.GetID())
	}
	if originToken.GetModel() != model {
		t.Errorf("Token model does not match. Original: %d, encoded: %d", model, originToken.GetModel())
	}
	if originToken.GetModel() != decodedToken.GetModel() {
		t.Errorf("Token model does not match. Encoded: %d, decoded: %d", originToken.GetModel(), decodedToken.GetModel())
	}
	if originToken.GetExtraID() != extraID {
		t.Errorf("Token extraID does not match. Original: %d, encoded: %d", extraID, originToken.GetExtraID())
	}
	if originToken.GetExtraID() != decodedToken.GetExtraID() {
		t.Errorf("Token extraID does not match. Encoded: %d, decoded: %d", originToken.GetExtraID(), decodedToken.GetExtraID())
	}
	if originToken.GetDate() != decodedToken.GetDate() {
		t.Errorf("Token datetime does not match. Encoded: %s, decoded: %s", originToken.GetDate(), decodedToken.GetDate())
	}
	if originToken.IsFixed() {
		t.Errorf("Token datetime should not be fixed: %t", originToken.IsFixed())
	}
	if originToken.IsFixed() != decodedToken.IsFixed() {
		t.Errorf("Token datetime does not match. Encoded: %t, decoded: %t", originToken.IsFixed(), decodedToken.IsFixed())
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
	if token.GetID() > 0 {
		t.Error("Token should have zero customer ID")
	}
	if token.GetModel() != 0 {
		t.Error("Token should have zero model")
	}
	if token.GetExtraID() != 0 {
		t.Error("Token should have zero extraID")
	}
}
