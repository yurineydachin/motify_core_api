package context

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCookiesInContextComplexTest(t *testing.T) {
	var cookie1 = http.Cookie{
		Name:     "test1",
		Value:    "true",
		HttpOnly: true,
	}
	var cookie2 = http.Cookie{
		Name:  "test2",
		Value: "hello",
	}

	data := Context{}
	var ctx = NewContext(context.Background(), &data)

	// try to add one cookie
	err := AddResponseCookies(ctx, &cookie1)
	assert.NoError(t, err, "Error")
	// check result
	cookies, ok := GetResponseCookies(ctx)
	assert.True(t, ok, "Empty cookies")
	assert.Equal(t, 1, len(cookies), "Cookies number is invalid")
	assert.Equal(t, &cookie1, cookies[0], "Wrong cookie data")

	// try to add another cookie
	err = AddResponseCookies(ctx, &cookie2)
	assert.NoError(t, err, "Error")
	// check result
	cookies, ok = GetResponseCookies(ctx)
	assert.True(t, ok, "Empty cookies")
	assert.Equal(t, 2, len(cookies), "Cookies number is invalid")
	assert.Equal(t, &cookie1, cookies[0], "Wrong cookie data")
	assert.Equal(t, &cookie2, cookies[1], "Wrong cookie data")
}
