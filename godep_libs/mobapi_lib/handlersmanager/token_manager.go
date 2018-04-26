package handlersmanager

import (
	"errors"
	"reflect"

	"github.com/sergei-svistunov/gorpc"
	"godep.lzd.co/mobapi_lib/token"
)

const (
	errorCodeInvalidToken = "INVALID_TOKEN"

	errMsgAccessWithoutRequiredToken                   = "Access without token has not permitted"
	errMsgTokenParsingFailed                           = "Token parsing error"
	errMsgTokenDataParsedButInvalid                    = "Invalid token"
	errMsgGuestTokenProvidedButShouldBeAuthorizedToken = "Access with guest token has not permitted"

	// TokenTypeAuthorized is token of authorized user
	TokenTypeAuthorized TokenType = iota
	// TokenTypeGuest is token of guest user
	TokenTypeGuest
	// TokenTypeAny allows to use any type of token for handler or even use handler without any token
	TokenTypeAny
	// TokenTypeNoToken allows to use handler without any token
	TokenTypeNoToken
)

type TokenType int

func getTokenType(method reflect.Method) TokenType {
	tokenType := TokenTypeNoToken

	// handler requires access token
	if method.Type.NumIn() == 4 {
		switch method.Type.In(3).Name() {
		case "INullToken":
			tokenType = TokenTypeAny
		case "IGuestToken":
			tokenType = TokenTypeGuest
		case "IToken":
			tokenType = TokenTypeAuthorized
		}
	}

	return tokenType
}

func prepareToken(tokenHeaderHash string, extraData interface{}, tokenModel uint64) (apiToken token.IToken, shouldAppendToInputs bool, err error) {
	shouldAppendToInputs = false

	// fetch expected token type for handler from provided extra data interface
	expectedTokenType, ok := extraData.(TokenType)
	if !ok {
		return nil, shouldAppendToInputs, errors.New("Empty or invalid extra data. Couldn't fetch token type for handler.")
	}
	if expectedTokenType == TokenTypeNoToken {
		return nil, shouldAppendToInputs, nil
	}

	if tokenHeaderHash == "" {
		switch expectedTokenType {
		case TokenTypeNoToken:
			return nil, false, nil
		case TokenTypeAny:
			return nil, true, nil
		default:
			return nil, true, &gorpc.HandlerError{
				UserMessage: errMsgAccessWithoutRequiredToken,
				Err:         errors.New("Token header has not provided in request"),
				Code:        errorCodeInvalidToken,
			}
		}
	}

	shouldAppendToInputs = true

	// try to create token by provided hash
	apiToken, err = token.ParseToken(tokenHeaderHash)
	if err != nil {
		return nil, shouldAppendToInputs, &gorpc.HandlerError{
			UserMessage: errMsgTokenParsingFailed,
			Err:         err,
			Code:        errorCodeInvalidToken,
		}
	}

	// check is token valid
	if !apiToken.IsValid() {
		return nil, shouldAppendToInputs, &gorpc.HandlerError{
			UserMessage: errMsgTokenDataParsedButInvalid,
			Err:         err,
			Code:        errorCodeInvalidToken,
		}
	}

	// check is token type is valid for handler
	if expectedTokenType == TokenTypeAuthorized && apiToken.IsGuest() {
		return nil, shouldAppendToInputs, &gorpc.HandlerError{
			UserMessage: errMsgGuestTokenProvidedButShouldBeAuthorizedToken,
			Err:         errors.New("Guest token used in request instead of full customer token"),
			Code:        errorCodeInvalidToken,
		}
	}

	// check is dynamic token is valid for handler
	if apiToken.IsFixed() {
		return nil, shouldAppendToInputs, &gorpc.HandlerError{
			UserMessage: errMsgGuestTokenProvidedButShouldBeAuthorizedToken,
			Err:         errors.New("Wrong token used in request instead of dynamic token with time"),
			Code:        errorCodeInvalidToken,
		}
	}

	// check is token model is valid for handler
	if tokenModel != apiToken.GetModel() {
		return nil, shouldAppendToInputs, &gorpc.HandlerError{
			UserMessage: errMsgGuestTokenProvidedButShouldBeAuthorizedToken,
			Err:         errors.New("Wrong token used in request instead of full token with needed model"),
			Code:        errorCodeInvalidToken,
		}
	}

	return apiToken, shouldAppendToInputs, nil
}
