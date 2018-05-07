package healthcheck

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewContext(t *testing.T) {
	var ctx context.Context
	var hc *HealthCheck

	ctx2 := NewContext(ctx, hc)
	assert.NotNil(t, ctx2)

	var key = "test"
	var val interface{} = 123

	ctx = context.WithValue(context.Background(), key, val)
	hc = &HealthCheck{version: "version"}
	ctx2 = NewContext(ctx, hc)
	assert.NotNil(t, ctx2)
	assert.Equal(t, val, ctx2.Value(key))
	assert.NotNil(t, ctx2.Value(contextKey))

	hcVal := ctx2.Value(contextKey)
	hcValConverted, ok := hcVal.(*HealthCheck)
	assert.True(t, ok)
	assert.NotNil(t, hcValConverted)
	assert.Equal(t, "version", hcValConverted.version)
}

func TestFromContext(t *testing.T) {
	hcVal, ok := FromContext(context.Background())
	assert.Nil(t, hcVal)
	assert.False(t, ok)

	hc := &HealthCheck{version: "version"}
	ctx := context.WithValue(context.Background(), contextKey, hc)

	hcVal, ok = FromContext(ctx)
	assert.True(t, ok)
	assert.NotNil(t, hcVal)
	assert.Equal(t, "version", hcVal.version)
}
