package token

import (
	"fmt"

	mtoken "godep.lzd.co/mobapi_lib/token"
)

const (
    ModelMobileUser = uint64(1);
    ModelAgentUser = uint64(2);
    ModelEmployeeQR = uint64(3);
)

type Token struct {
    t mtoken.IToken
}

func newToken(id uint64, modelID uint64, extraID uint64) *Token {
    return &Token{
        t: mtoken.NewTokenV1(id, modelID, extraID),
    }
}

func NewMobileUser(id uint64) *Token {
    return newToken(id, ModelMobileUser, 0)
}

func NewAgentUser(id uint64, integrationID uint64) *Token {
    return newToken(id, ModelAgentUser, integrationID)
}

func NewEmployeeQR(id uint64, integrationID uint64) *Token {
    return newToken(id, ModelEmployeeQR, integrationID)
}

func (token *Token) String() string {
    if token != nil && token.t != nil {
        return token.t.String()
    }
    return ""
}

func parseToken(value string, modelID uint64) (mtoken.IToken, error) {
	t, err := mtoken.ParseToken(value)
	if err != nil {
		return nil, err
	}
	if t.GetID() == 0 || t.GetModel() != modelID {
		return nil, fmt.Errorf("Invalid token")
	}
    return t, nil
}

func ParseEmployeeQR(value string) (mtoken.IToken, error) {
    return parseToken(value, ModelEmployeeQR)
}
