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

func (service *UserService) GetUserByID(ctx context.Context, modelID uint64) (*models.User, error) {
	res := models.User{}
	err := service.db.Get(&res, `
        SELECT id_user, u_fk_integration, u_name, u_short, u_description, u_avatar, u_phone, u_email, u_phone_approved, u_email_approved, u_updated_at, u_created_at
        FROM motify_users WHERE id_user = ?
    `, modelID)
	return &res, err
}

func (service *UserService) SetUser(ctx context.Context, model *models.User) (uint64, error) {
	if model.ID > 0 {
		return service.updateUser(ctx, model)
	}
	return service.createUser(ctx, model)
}

func (service *UserService) createUser(ctx context.Context, model *models.User) (uint64, error) {
	args := model.ToArgs()
	fkField := ""
	fkValue := ""
	if _, exists := args["u_fk_integration"]; exists {
		fkField = "u_fk_integration, "
		fkValue = ":u_fk_integration, "
	}
	insertRes, err := service.db.Exec(fmt.Sprintf(`
            INSERT INTO motify_users (u_name, %s u_short, u_description, u_avatar, u_phone, u_email, u_phone_approved, u_email_approved)
            VALUES (:u_name, %s :u_short, :u_description, :u_avatar, :u_phone, :u_email, :u_phone_approved, :u_email_approved)
        `, fkField, fkValue), model.ToArgs())
	if err != nil {
		return 0, fmt.Errorf("Insert DB exec error: %v", err)
	}

	id, err := insertRes.LastInsertId()
	if id == 0 {
		return 0, fmt.Errorf("Insert DB users error: id = 0")
	}
	return uint64(id), err
}

func (service *UserService) updateUser(ctx context.Context, model *models.User) (uint64, error) {
	args := model.ToArgs()
	fkField := ""
	if _, exists := args["e_fk_user"]; exists {
		fkField = "e_fk_user = :e_fk_user,"
	}
	updateRes, err := service.db.Exec(fmt.Sprintf(`
            UPDATE motify_users SET
                u_name = :u_name,
                %s
                u_short = :u_short,
                u_description = :u_description,
                u_avatar = :u_avatar,
                u_phone = :u_phone,
                u_email = :u_email
                u_phone_approved = :u_phone_approved
                u_email_approved = :u_email_approved
            WHERE id_user = :id_user
        `, fkField), args)
	if err != nil {
		return 0, fmt.Errorf("Update DB exec error: %v", err)
	}
	rowsCount, err := updateRes.RowsAffected()
	if rowsCount == 0 {
		return 0, fmt.Errorf("Update DB exec error: nothing changed")
	}
	return model.ID, nil
}

func (service *UserService) DeleteUser(ctx context.Context, modelID uint64) error {
	deleteRes, err := service.db.Exec(`
            DELETE FROM motify_users
            WHERE id_user = :id_user
        `, map[string]interface{}{
		"id_user": modelID,
	})
	if err != nil {
		return fmt.Errorf("Delete DB exec error: %v", err)
	}
	rowsCount, err := deleteRes.RowsAffected()
	if rowsCount == 0 {
		return fmt.Errorf("Delete DB exec error: nothing changed")
	}
	return nil
}

func (service *UserService) SetUserAccess(ctx context.Context, model *models.UserAccess) (uint64, error) {
	if model.ID > 0 {
		return service.updateUserAccess(ctx, model)
	}
	return service.createUserAccess(ctx, model)
}

func (service *UserService) createUserAccess(ctx context.Context, model *models.UserAccess) (uint64, error) {
	if model.UserFK == 0 {
		return 0, fmt.Errorf("Insert DB exec error: no fk_user")
	}
	insertRes, err := service.db.Exec(`
            INSERT INTO motify_user_access (ua_fk_user, ua_type, ua_login, ua_password)
            VALUES (:ua_fk_user, :ua_type, :ua_login, :ua_password)
        `, model.ToArgs())
	if err != nil {
		return 0, fmt.Errorf("Insert DB exec error: %v", err)
	}

	id, err := insertRes.LastInsertId()
	if id == 0 {
		return 0, fmt.Errorf("Insert DB user_access error: id = 0")
	}
	return uint64(id), err
}

func (service *UserService) updateUserAccess(ctx context.Context, model *models.UserAccess) (uint64, error) {
	updateRes, err := service.db.Exec(`
            UPDATE motify_user_access SET
                ua_type = :ua_type,
                ua_login = :ua_login,
                ua_password = :ua_password
            WHERE id_user_access = :id_user_access
        `, model.ToArgs())
	if err != nil {
		return 0, fmt.Errorf("Update DB exec error: %v", err)
	}
	rowsCount, err := updateRes.RowsAffected()
	if rowsCount == 0 {
		return 0, fmt.Errorf("Update DB exec error: nothing changed")
	}
	return model.ID, nil
}

func (service *UserService) IsLoginBusy(ctx context.Context, login string) (bool, error) {
	accessList, err := service.getUserAssessListByLogin(login)
	if err != nil {
		return false, err
	}
	if len(accessList) > 0 {
		return true, nil
	}
	return false, nil
}

func (service *UserService) GetUserIDByLogin(ctx context.Context, login string) (uint64, error) {
	accessList, err := service.getUserAssessListByLogin(login)
	if err != nil {
		return 0, err
	}
	if len(accessList) == 0 || len(accessList) > 1 {
		return 0, nil
	}
	return accessList[0].UserFK, nil
}

func (service *UserService) Authentificate(ctx context.Context, login, password string) (uint64, error) {
	accessList, err := service.getUserAssessListByLoginAndPass(login, password)
	if err != nil {
		return 0, err
	}
	if len(accessList) == 0 {
		return 0, nil //fmt.Errorf("Could not authentificate user by login '%s' and password '%s', not found", login, password)
	}
	userIDs := make(map[uint64]bool, len(accessList))
	for i := range accessList {
		userIDs[accessList[i].UserFK] = true
	}
	if len(userIDs) > 1 {
		return 0, fmt.Errorf("Could not authentificate user by login '%s' and password '%s', too many users: %v", login, password, userIDs)
	} else if len(userIDs) == 0 {
		return accessList[0].UserFK, nil
	}
	return 0, nil
}

func (service *UserService) getUserAssessListByLogin(login string) ([]*models.UserAccess, error) {
	res := []*models.UserAccess{}
	loginHash := utils.Hash(login)
	err := service.db.Select(&res, `
        SELECT id_user_access, ua_fk_user, ua_type, ua_login, ua_password, ua_updated_at, ua_created_at
        FROM motify_user_access WHERE ua_login = ?
    `, loginHash)
	for i, access := range res {
		access.MarkAllHashed()
		res[i] = access
	}
	return res, err
}

func (service *UserService) getUserAssessListByLoginAndPass(login, password string) ([]*models.UserAccess, error) {
	res := []*models.UserAccess{}
	loginHash := utils.Hash(login)
	passwordHash := utils.Hash(password)
	err := service.db.Select(&res, `
        SELECT id_user_access, ua_fk_user, ua_type, ua_login, ua_password, ua_updated_at, ua_created_at
        FROM motify_user_access WHERE ua_login = ? AND ua_password = ?
    `, loginHash, passwordHash)
	for i, access := range res {
		access.MarkAllHashed()
		res[i] = access
	}
	return res, err
}

func (service *UserService) GetUserAssessByUserIDAndType(ctx context.Context, userID uint64, t string) (*models.UserAccess, error) {
	list, err := service.GetUserAssessListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, access := range list {
		if access.Type == t {
			return access, nil
		}
	}
	return nil, nil
}

func (service *UserService) GetUserAssessListByUserID(ctx context.Context, userID uint64) (map[string]*models.UserAccess, error) {
	list := []*models.UserAccess{}
	err := service.db.Select(&list, `
        SELECT id_user_access, ua_fk_user, ua_type, ua_login, ua_password, ua_updated_at, ua_created_at
        FROM motify_user_access WHERE ua_fk_user = ?
    `, userID)

	res := make(map[string]*models.UserAccess, len(list))
	for _, access := range list {
		access.MarkAllHashed()
		res[access.Type] = access
	}
	return res, err
}

func (service *UserService) DeleteUserAccessByUserID(ctx context.Context, userID uint64) error {
	deleteRes, err := service.db.Exec(`
            DELETE FROM motify_user_access
            WHERE ua_fk_user = :ua_fk_user
        `, map[string]interface{}{
		"ua_fk_user": userID,
	})
	if err != nil {
		return fmt.Errorf("Delete DB exec error: %v", err)
	}
	rowsCount, err := deleteRes.RowsAffected()
	if rowsCount == 0 {
		return fmt.Errorf("Delete DB exec error: nothing changed")
	}
	return nil
}
