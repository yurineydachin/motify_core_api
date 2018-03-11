package user_service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"motify_core_api/resources/database"
	"motify_core_api/models"
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

func TestCreateUser(t *testing.T) {
    service := getService(t)
    user := models.User{
        Name: "user 1",
        Short: "short 1",
        Description: "desc 1",
        Awatar: "awatar 1",
        Phone:  "phone 1",
        Email: "email_1@text.com",
    }
    userID, err := service.SetUser(context.Background(), user)
    if assert.Nil(t, err, "err from DB") {
        assert.Equal(t, userID > 0, true, "user result from DB is empty")
    }
}
