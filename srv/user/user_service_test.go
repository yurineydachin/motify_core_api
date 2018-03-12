package user_service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"motify_core_api/models"
	"motify_core_api/resources/database"
)

func getService(t *testing.T) *UserService {
	db, err := database.NewDbAdapter([]string{"root:123456@tcp(localhost:3306)/motify_core_api"}, "Europe/Moscow", false)
	if assert.Nil(t, err, "DB adapter init error") &&
		assert.NotNil(t, db, "DB adapter is empty") {
		return NewUserService(db)
	}
	return nil
}

func TestCreateDBAdapterAndService(t *testing.T) {
	assert.NotNil(t, getService(t), "service is nil")
}

func TestGetUserByID(t *testing.T) {
	service := getService(t)
	user, err := service.GetUserByID(context.Background(), 1)
	if assert.Nil(t, err, "err from DB") &&
		assert.NotNil(t, user, "user from DB is empty") {
		assert.Equal(t, user.ID, uint64(1))
		assert.Equal(t, user.Name, "user 1")
	}
}

func testSetUser_Create(t *testing.T) {
	service := getService(t)
	user := models.User{
		Name:        "user 1",
		Short:       "short 1",
		Description: "desc 1",
		Awatar:      "awatar 1",
		Phone:       "phone 1",
		Email:       "email_1@text.com",
	}
	userID, err := service.SetUser(context.Background(), user)
	if assert.Nil(t, err, "err from db") {
		assert.Equal(t, userID > 0, true, "user result from db is empty")
	}
}

func testSetUser_Update(t *testing.T) {
	service := getService(t)
	user := models.User{
		ID:          2,
		Name:        "user 2",
		Short:       "short 2",
		Description: "desc 2",
		Awatar:      "awatar 2",
		Phone:       "phone 2",
		Email:       "email_2@text.com",
	}
	userID, err := service.SetUser(context.Background(), user)
	if assert.Nil(t, err, "err from db") {
		assert.Equal(t, userID, user.ID, "user result from db is empty")
	}

	userNew, err := service.GetUserByID(context.Background(), userID)
	if assert.Nil(t, err, "err from DB") &&
		assert.NotNil(t, userNew, "user from DB is empty") {
		assert.Equal(t, userNew.ID, uint64(2))
		assert.Equal(t, userNew.Name, "user 2")
	}
}
