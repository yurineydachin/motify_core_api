package user_service

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"motify_core_api/models"
	"motify_core_api/resources/database"
	"motify_core_api/utils"
)

var (
	testUserID       = uint64(0)
	testUserIDStr    = "0"
	testUserAccessID = uint64(0)
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

func TestSetUser_Create(t *testing.T) {
	service := getService(t)
	user := &models.User{
		Name:        "user test",
		Short:       "short test",
		Description: "desc test",
		Avatar:      "avatar test",
		Phone:       "phone test",
		Email:       "email_test@text.com",
	}
	var err error
	testUserID, err = service.SetUser(context.Background(), user)
	testUserIDStr = fmt.Sprintf("%d", testUserID)
	if assert.Nil(t, err) {
		assert.Equal(t, testUserID > 0, true)
	}
}

func TestGetUserByID(t *testing.T) {
	service := getService(t)
	user, err := service.GetUserByID(context.Background(), testUserID)
	if assert.Nil(t, err) &&
		assert.NotNil(t, user) {
		assert.Equal(t, user.ID, testUserID)
		assert.Equal(t, user.Name, "user test")
	}
}

func TestSetUser_Update(t *testing.T) {
	service := getService(t)
	user := &models.User{
		ID:          testUserID,
		Name:        "user " + testUserIDStr,
		Short:       "short " + testUserIDStr,
		Description: "desc " + testUserIDStr,
		Avatar:      "avatar " + testUserIDStr,
		Phone:       "phone " + testUserIDStr,
		Email:       "email_" + testUserIDStr + "@text.com",
	}
	userID, err := service.SetUser(context.Background(), user)
	if assert.Nil(t, err) {
		assert.Equal(t, userID, user.ID)
	}

	userNew, err := service.GetUserByID(context.Background(), userID)
	if assert.Nil(t, err) &&
		assert.NotNil(t, userNew) {
		assert.Equal(t, userNew.ID, testUserID)
		assert.Equal(t, userNew.Name, "user "+testUserIDStr)
	}
}

func TestSetUserAccess_Create(t *testing.T) {
	login := "email_1@text.com"
	service := getService(t)
	access := &models.UserAccess{
		UserFK:   testUserID,
		Type:     models.UserAccessEmail,
		Login:    &login,
		Password: "password 1",
	}
	var err error
	testUserAccessID, err = service.SetUserAccess(context.Background(), access)
	if assert.Nil(t, err) {
		assert.Equal(t, testUserAccessID > 0, true, "user_access result from db is empty")
	}
}

func TestGetUserAssessListByUserID(t *testing.T) {
	service := getService(t)
	accessList, err := service.GetUserAssessListByUserID(context.Background(), testUserID)
	if assert.Nil(t, err) &&
		assert.Equal(t, len(accessList), 1) {
		assert.Equal(t, accessList[models.UserAccessEmail].UserFK, testUserID)
		assert.Equal(t, accessList[models.UserAccessEmail].Type, models.UserAccessEmail)
		assert.Equal(t, accessList[models.UserAccessEmail].Login, utils.Hash("email_1@text.com"))
		assert.Equal(t, accessList[models.UserAccessEmail].Password, utils.Hash("password 1"))
	}
}

func TestSetUserAccess_Update(t *testing.T) {
	login := "phone " + testUserIDStr
	service := getService(t)
	access := &models.UserAccess{
		ID:       testUserAccessID,
		UserFK:   testUserID,
		Type:     models.UserAccessPhone,
		Login:    &login,
		Password: "password " + testUserIDStr,
	}

	accessID, err := service.SetUserAccess(context.Background(), access)
	if assert.Nil(t, err) {
		assert.Equal(t, accessID, access.ID, "user result from db is empty")
	}

	accessNew, err := service.GetUserAssessListByUserID(context.Background(), access.UserFK)
	if assert.Nil(t, err) &&
		assert.Equal(t, len(accessNew) > 0, true, "user from DB is empty") {
		assert.Equal(t, accessNew[models.UserAccessEmail].ID, testUserAccessID)
		assert.Equal(t, accessNew[models.UserAccessEmail].UserFK, testUserID)
		assert.Equal(t, accessNew[models.UserAccessEmail].Type, models.UserAccessPhone)
		assert.Equal(t, accessNew[models.UserAccessEmail].Login, utils.Hash("phone "+testUserIDStr))
		assert.Equal(t, accessNew[models.UserAccessEmail].Password, utils.Hash("password "+testUserIDStr))
	}
}

func TestIsLoginBusy_OldEmail(t *testing.T) {
	isExist, err := getService(t).IsLoginBusy(context.Background(), "email_1@text.com")
	assert.Nil(t, err)
	assert.Equal(t, isExist, false)
}

func TestIsLoginBusy_OldPhone(t *testing.T) {
	isExist, err := getService(t).IsLoginBusy(context.Background(), "phone 1")
	assert.Nil(t, err)
	assert.Equal(t, isExist, false)
}

func TestCanNotAuthentificateByOldEmailAndPassword(t *testing.T) {
	service := getService(t)
	userID, err := service.Authentificate(context.Background(), "email_1@text.com", "password "+testUserIDStr)
	assert.NotNil(t, err)
	assert.Equal(t, userID, uint64(0))
}

func TestIsLoginBusy_NewEmail(t *testing.T) {
	isExist, err := getService(t).IsLoginBusy(context.Background(), "email_"+testUserIDStr+"@text.com")
	assert.Nil(t, err)
	assert.Equal(t, isExist, true)
}

func TestIsLoginBusy_NewPhone(t *testing.T) {
	isExist, err := getService(t).IsLoginBusy(context.Background(), "phone "+testUserIDStr)
	assert.Nil(t, err)
	assert.Equal(t, isExist, true)
}

func TestAuthentificateByEmailAndPassword(t *testing.T) {
	service := getService(t)
	userID, err := service.Authentificate(context.Background(), "email_"+testUserIDStr+"@text.com", "password "+testUserIDStr)
	assert.Nil(t, err)
	assert.Equal(t, userID, testUserID)
}

func TestAuthentificateByPhoneAndPassword(t *testing.T) {
	service := getService(t)
	userID, err := service.Authentificate(context.Background(), "phone "+testUserIDStr, "password "+testUserIDStr)
	assert.Nil(t, err)
	assert.Equal(t, userID, testUserID)
}

func TestDeleteUserAccessByUserID(t *testing.T) {
	service := getService(t)
	assert.Nil(t, service.DeleteUserAccessByUserID(context.Background(), testUserID))
}

func TestDeleteUser(t *testing.T) {
	service := getService(t)
	assert.Nil(t, service.DeleteUser(context.Background(), testUserID))
}

/*
// create test user
func TestSetUser_Create1(t *testing.T) {
	service := getService(t)
	user := &models.User{
		Name:        "yuri",
		Short:       "short no",
		Description: "desc no",
		Avatar:      "no",
		Phone:       "123456",
		Email:       "yuri@test.com",
	}
	var err error
	testUserID, err = service.SetUser(context.Background(), user)
	testUserIDStr = fmt.Sprintf("%d", testUserID)
	if assert.Nil(t, err) {
		assert.Equal(t, testUserID > 0, true)
	}
}

func TestSetUserAccess_Create1(t *testing.T) {
	login := "yuri@test.com"
	service := getService(t)
	access := &models.UserAccess{
		UserFK:   testUserID,
		Type:     models.UserAccessEmail,
		Login:    &login,
		Password: "123456",
	}
	var err error
	testUserAccessID, err = service.SetUserAccess(context.Background(), access)
	if assert.Nil(t, err) {
		assert.Equal(t, testUserAccessID > 0, true, "user_access result from db is empty")
	}
}
*/
