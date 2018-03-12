package user_service

import (
	"context"
	"fmt"

	"motify_core_api/models"
	"motify_core_api/resources/database"
	"motify_core_api/utils"
)

type UserService struct {
	db *database.DbAdapter
}

func NewUserService(db *database.DbAdapter) *UserService {
	return &UserService{
		db: db,
	}
}

func (service *UserService) GetUserByID(ctx context.Context, id uint64) (*models.User, error) {
	res := models.User{}
	err := service.db.Get(&res, `
        SELECT id_user, name, p_description, description, awatar, phone, email, updated_at, created_at
        FROM motify_users WHERE id_user = ?
    `, id)
	return &res, err
}

func (service *UserService) SetUser(ctx context.Context, user models.User) (uint64, error) {
	if user.ID > 0 {
		return service.updateUser(ctx, user)
	}
	return service.createUser(ctx, user)
}

func (service *UserService) createUser(ctx context.Context, user models.User) (uint64, error) {
	insertRes, err := service.db.Exec(`
            INSERT INTO motify_users (name, p_description, description, awatar, phone, email)
            VALUES (:name, :p_description, :description, :awatar, :phone, :email)
        `, map[string]interface{}{
		"name":          user.Name,
		"p_description": user.Short,
		"description":   user.Description,
		"awatar":        user.Awatar,
		"phone":         user.Phone,
		"email":         user.Email,
	})
	if err != nil {
		return 0, fmt.Errorf("Insert DB exec error: %v", err)
	}

	id, err := insertRes.LastInsertId()
	return uint64(id), err
}

func (service *UserService) updateUser(ctx context.Context, user models.User) (uint64, error) {
	updateRes, err := service.db.Exec(`
            UPDATE motify_users SET
                name = :name,
                p_description = :p_description,
                description = :description,
                awatar = :awatar,
                phone = :phone,
                email = :email
            WHERE id_user = :id
        `, map[string]interface{}{
		"id":            user.ID,
		"name":          user.Name,
		"p_description": user.Short,
		"description":   user.Description,
		"awatar":        user.Awatar,
		"phone":         user.Phone,
		"email":         user.Email,
	})
	if err != nil {
		return 0, fmt.Errorf("Update DB exec error: %v", err)
	}
	rowsCount, err := updateRes.RowsAffected()
	if rowsCount == 0 {
		return 0, fmt.Errorf("Update DB exec error: nothing changed")
	}
	return user.ID, nil
}

func (service *UserService) SetUserAccess(ctx context.Context, access models.UserAccess) (uint64, error) {
	if access.ID > 0 {
		return service.updateUserAccess(ctx, access)
	}
	return service.createUserAccess(ctx, access)
}

func (service *UserService) createUserAccess(ctx context.Context, access models.UserAccess) (uint64, error) {
	if access.UserFK == 0 {
		return 0, fmt.Errorf("Insert DB exec error: no fk_user")
	}
	insertRes, err := service.db.Exec(`
            INSERT INTO motify_user_access (fk_user, type_access, phone, email, password)
            VALUES (:fk_user, :type_access, :phone, :email, :password)
        `, access.ToArgs())
	if err != nil {
		return 0, fmt.Errorf("Insert DB exec error: %v", err)
	}

	id, err := insertRes.LastInsertId()
	return uint64(id), err
}

func (service *UserService) updateUserAccess(ctx context.Context, access models.UserAccess) (uint64, error) {
	updateRes, err := service.db.Exec(`
            UPDATE motify_users SET
                type_access = :name,
                phone = :phone,
                email = :email
                password = :password
            WHERE id_user_access = :id
        `, access.ToArgs())
	if err != nil {
		return 0, fmt.Errorf("Update DB exec error: %v", err)
	}
	rowsCount, err := updateRes.RowsAffected()
	if rowsCount == 0 {
		return 0, fmt.Errorf("Update DB exec error: nothing changed")
	}
	return access.ID, nil
}

func (service *UserService) Authentificate(ctx context.Context, login, password string) (*models.User, error) {
	accessList, err := service.GetUserAssessListByLoginAndPass(login, password)
	if err != nil {
		return nil, err
	}
	if len(accessList) == 0 {
		return nil, fmt.Errorf("Could not authentificate user by login '%s' and password '%s', not found", login, password)
	}
	userIDs := make(map[uint64]bool, len(accessList))
	for i := range accessList {
		userIDs[accessList[i].UserFK] = true
	}
	if len(userIDs) > 1 {
		return nil, fmt.Errorf("Could not authentificate user by login '%s' and password '%s', too many users: %v", login, password, userIDs)
	}
	for i := range accessList {
		return service.GetUserByID(ctx, accessList[i].UserFK)
	}
	return nil, nil
}

func (service *UserService) GetUserAssessListByLoginAndPass(login, password string) ([]*models.UserAccess, error) {
	res := []*models.UserAccess{}
	loginHash := utils.Hash(login)
	passwordHash := utils.Hash(password)
	err := service.db.Select(&res, `
        SELECT id_user_access,fk_user,type_access,email,phone,password,updated_at,created_at
        FROM motify_user_access WHERE (email = ? OR phone = ?) AND password = ?
    `, loginHash, loginHash, passwordHash)
	for i, access := range res {
		access.MarkAllHashed()
		res[i] = access
	}
	return res, err
}

func (service *UserService) GetUserAssessList(ctx context.Context, id uint64) ([]*models.UserAccess, error) {
	res := []*models.UserAccess{}
	err := service.db.Select(&res, `
        SELECT id_user_access,fk_user,type_access,email,phone,password,updated_at,created_at
        FROM motify_user_access WHERE fk_user = ?
    `, id)
	for i, access := range res {
		access.MarkAllHashed()
		res[i] = access
	}
	return res, err
}
