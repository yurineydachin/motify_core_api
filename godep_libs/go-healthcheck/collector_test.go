package healthcheck

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetResourceStatus(t *testing.T) {
	var result string
	someErr := fmt.Errorf("Some error")

	result = getResourceStatus(false, nil)
	assert.Equal(t, result, ServiceResourceStatusOk)

	result = getResourceStatus(false, someErr)
	assert.Equal(t, result, ServiceResourceStatusUnstable)

	result = getResourceStatus(true, nil)
	assert.Equal(t, result, ServiceResourceStatusOk)

	result = getResourceStatus(true, someErr)
	assert.Equal(t, result, ServiceResourceStatusError)
}
