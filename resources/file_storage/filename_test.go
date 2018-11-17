package file_storage_service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateFileName(t *testing.T) {
	assert.Equal(t, "users/fe98f58d2f/62906d4bb4.jpg", GenerateFileName("users", "pic.jpg", 123))
	assert.Equal(t, "users/fe98f58d2f/3687aae10c.png", GenerateFileName("users", "pic.png", 123))
	assert.Equal(t, "agents/949811423f/a591353d5f.jpg", GenerateFileName("agents", "pic.jpg", 123))
	assert.Equal(t, "agents/949811423f/5b7d865ce5.png", GenerateFileName("agents", "pic.png", 123))
}
