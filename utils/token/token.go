package token

import (
	"fmt"

	mtoken "godep.lzd.co/mobapi_lib/token"
)

const (
	ModelMobileUser    = uint64(1)
	ModelAgentUser     = uint64(2)
	ModelEmployee      = uint64(3)
	ModelAgent         = uint64(4)
	ModelAgentSettings = uint64(5)
	ModelPayslip       = uint64(6)
	ModelSetting       = uint64(7)
	ModelRemindUser    = uint64(8)
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

func NewEmployee(id uint64, integrationID uint64) *Token {
	return newToken(id, ModelEmployee, integrationID)
}

func NewAgent(id uint64, integrationID uint64) *Token {
	return newToken(id, ModelAgent, integrationID)
}

func NewPayslip(id uint64, integrationID uint64) *Token {
	return newToken(id, ModelPayslip, integrationID)
}

func NewSetting(id uint64, integrationID uint64) *Token {
	return newToken(id, ModelSetting, integrationID)
}

func NewRemindUser(id, integrationID uint64) *Token {
	return newToken(id, ModelRemindUser, integrationID)
}

func (token *Token) Fixed() *Token {
	if token != nil && token.t != nil {
		token.t.Fixed()
		return token
	}
	return nil
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

func ParseEmployee(value string) (mtoken.IToken, error) {
	return parseToken(value, ModelEmployee)
}

func ParseAgent(value string) (mtoken.IToken, error) {
	return parseToken(value, ModelAgent)
}

func ParsePayslip(value string) (mtoken.IToken, error) {
	return parseToken(value, ModelPayslip)
}

func ParseRemindUser(value string) (mtoken.IToken, error) {
	return parseToken(value, ModelRemindUser)
}
