package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkAllHashed(t *testing.T) {
	access := &UserAccess{
		IsHashedLogin:    false,
		IsHashedPassword: false,
	}
	access.MarkAllHashed()
	assert.Equal(t, access.IsHashedLogin, true, "user result from DB is empty")
	assert.Equal(t, access.IsHashedPassword, true, "user result from DB is empty")
}
