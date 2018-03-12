package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkAllHashed(t *testing.T) {
	access := &UserAccess{
		IsHashedEmail:    false,
		IsHashedPhone:    false,
		IsHashedPassword: false,
	}
	access.MarkAllHashed()
	assert.Equal(t, access.IsHashedEmail, true, "user result from DB is empty")
	assert.Equal(t, access.IsHashedPhone, true, "user result from DB is empty")
	assert.Equal(t, access.IsHashedPassword, true, "user result from DB is empty")
}
