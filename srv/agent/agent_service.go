package agent_service

import (
	"context"
	"fmt"

	"motify_core_api/models"
	"motify_core_api/resources/database"
)

const (
	defaultLimit = 30
)

type AgentService struct {
	db *database.DbAdapter
}

func NewAgentService(db *database.DbAdapter) *AgentService {
	return &AgentService{
		db: db,
	}
}

func (service *AgentService) GetSettingsListByUserID(ctx context.Context, userID, limit, offset uint64) ([]*models.AgentWithSetting, error) {
	if limit == 0 || limit > defaultLimit {
		limit = defaultLimit
	}
	res := []*models.AgentWithSetting{}

	err := service.db.Select(&res, `
        SELECT
            a.id_agent, a.name, a.company_id, a.description, a.logo, a.bg_image, a.address, a.phone, a.email, a.site, a.updated_at, a.created_at,
            s.role, s.notifications_enabled, s.is_main_agent
        INNER JOIN motify_agent_settings s ON s.fk_agent = a.id_agent
        FROM motify_agents a
        WHERE s.fk_user = ?
        LIMIT ?
        OFFSET ?
        ORDER BY s.created_at DESC
    `, userID, limit, offset)
	return res, err
}

func (service *AgentService) GetEmployeeListByUserID(ctx context.Context, userID, limit, offset uint64) ([]*models.AgentWithEmployee, error) {
	if limit == 0 || limit > defaultLimit {
		limit = defaultLimit
	}
	res := []*models.AgentWithEmployee{}

	err := service.db.Select(&res, `
        SELECT
            a.id_agent, a.name, a.company_id, a.description, a.logo, a.bg_image, a.address, a.phone, a.email, a.site, a.updated_at, a.created_at,
            e.fk_user, e.employee_code, e.hire_date, e.number_of_dependants, e.gross_base_salary, e.role
        INNER JOIN motify_agent_employee e ON e.fk_agent = a.id_agent
        FROM motify_agents a
        WHERE e.fk_user = ?
        LIMIT ?
        OFFSET ?
        ORDER BY e.created_at DESC
    `, userID, limit, offset)
	return res, err
}

func (service *AgentService) GetListByCompanyIDs(ctx context.Context, companyIDs []string) ([]*models.Agent, error) {
	res := []*models.Agent{}

	err := service.db.Select(&res, `
        SELECT
            id_agent, name, company_id, description, logo, bg_image, address, phone, email, site, updated_at, created_at
        FROM motify_agents a
        WHERE e.company_id IN (?)
        ORDER BY e.created_at DESC
    `, companyIDs)
	return res, err
}

func (service *AgentService) GetAgentByID(ctx context.Context, modelID uint64) (*models.Agent, error) {
	res := models.Agent{}
	err := service.db.Get(&res, `
        SELECT id_agent, name, company_id, description, logo, bg_image, address, phone, email, site, updated_at, created_at
        FROM motify_agents WHERE id_agent = ?
    `, modelID)
	return &res, err
}

func (service *AgentService) SetAgent(ctx context.Context, model *models.Agent) (uint64, error) {
	if model.ID > 0 {
		return service.updateAgent(ctx, model)
	}
	return service.createAgent(ctx, model)
}

func (service *AgentService) createAgent(ctx context.Context, model *models.Agent) (uint64, error) {
	insertRes, err := service.db.Exec(`
            INSERT INTO motify_agents (name, company_id, description, logo, bg_image, address, phone, email, site)
            VALUES (:name, :company_id, :description, :logo, :bg_image, :address, :phone, :email, :site)
        `, model.ToArgs())
	if err != nil {
		return 0, fmt.Errorf("Insert DB exec error: %v", err)
	}

	id, err := insertRes.LastInsertId()
	return uint64(id), err
}

func (service *AgentService) updateAgent(ctx context.Context, model *models.Agent) (uint64, error) {
	updateRes, err := service.db.Exec(`
            UPDATE motify_agents SET
                name = :name,
                company_id = :company_id,
                description = :description,
                logo = :logo,
                bg_image = :bg_image,
                address =  :address,
                phone = :phone,
                email = :email,
                site = :site
            WHERE
                id_agent = :id_agent
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

func (service *AgentService) DeleteAgent(ctx context.Context, modelID uint64) error {
	deleteRes, err := service.db.Exec(`
            DELETE FROM motify_agents
            WHERE
                id_agent = :id_agent
        `, map[string]interface{}{
		"id_agent": modelID,
	})
	if err != nil {
		return fmt.Errorf("Insert DB exec error: %v", err)
	}
	rowsCount, err := deleteRes.RowsAffected()
	if rowsCount == 0 {
		return fmt.Errorf("Delete DB exec error: nothing changed")
	}
	return nil
}

func (service *AgentService) GetEmployeeByAgentAndUser(ctx context.Context, agentFK, userFK uint64) (*models.Employee, error) {
	res := models.Employee{}
	err := service.db.Get(&res, `
        SELECT id_employee, fk_agent, fk_user, employee_code, name, email, hire_date, number_of_dependants, gross_base_salary, role, updated_at, created_at
        FROM motify_agent_employees WHERE fk_agent = ? AND fk_user = ?
    `, agentFK, userFK)
	return &res, err
}

func (service *AgentService) GetEmployeeByID(ctx context.Context, modelID uint64) (*models.Employee, error) {
	res := models.Employee{}
	err := service.db.Get(&res, `
        SELECT id_employee, fk_agent, fk_user, employee_code, name, email, hire_date, number_of_dependants, gross_base_salary, role, updated_at, created_at
        FROM motify_agent_employees WHERE id_employee = ?
    `, modelID)
	return &res, err
}

func (service *AgentService) SetEmployee(ctx context.Context, model *models.Employee) (uint64, error) {
	if model.ID > 0 {
		return service.updateEmployee(ctx, model)
	}
	return service.createEmployee(ctx, model)
}

func (service *AgentService) createEmployee(ctx context.Context, model *models.Employee) (uint64, error) {
	if model.AgentFK == 0 {
		return 0, fmt.Errorf("Insert DB exec error: no fk_agent")
	}
	args := model.ToArgs()
	fkField := ""
	fkValue := ""
	if _, exists := args["fk_user"]; exists {
		fkField = "fk_user, "
		fkValue = ":fk_user, "
	}
	sql := fmt.Sprintf(
		`INSERT INTO motify_agent_employees (fk_agent, %s employee_code, name, email, hire_date, number_of_dependants, gross_base_salary, role) 
        VALUES (:fk_agent, %s :employee_code, :name, :email, :hire_date, :number_of_dependants, :gross_base_salary, :role)`,
		fkField, fkValue)
	insertRes, err := service.db.Exec(sql, model.ToArgs())
	if err != nil {
		return 0, fmt.Errorf("Insert DB exec error: %v", err)
	}

	id, err := insertRes.LastInsertId()
	return uint64(id), err
}

func (service *AgentService) updateEmployee(ctx context.Context, model *models.Employee) (uint64, error) {
	if model.AgentFK == 0 {
		return 0, fmt.Errorf("Update DB exec error: no fk_agent")
	}
	args := model.ToArgs()
	fkField := ""
	if _, exists := args["fk_user"]; exists {
		fkField = "fk_user = :fk_user,"
	}
	sql := fmt.Sprintf(
		`UPDATE motify_agent_employees SET
                fk_agent = :fk_agent,
                %s
                employee_code = :employee_code,
                name = :name,
                email = :email,
                hire_date = :hire_date,
                number_of_dependants = :number_of_dependants,
                gross_base_salary = :gross_base_salary,
                role = :role
            WHERE
                id_employee = :id_employee`,
		fkField)
	updateRes, err := service.db.Exec(sql, args)
	if err != nil {
		return 0, fmt.Errorf("Update DB exec error: %v", err)
	}
	rowsCount, err := updateRes.RowsAffected()
	if rowsCount == 0 {
		return 0, fmt.Errorf("Update DB exec error: nothing changed")
	}
	return model.ID, nil
}

func (service *AgentService) DeleteEmployee(ctx context.Context, modelID uint64) error {
	deleteRes, err := service.db.Exec(`
            DELETE FROM motify_agent_employees
            WHERE
                id_employee = :id_employee
        `, map[string]interface{}{
		"id_employee": modelID,
	})
	if err != nil {
		return fmt.Errorf("Insert DB exec error: %v", err)
	}
	rowsCount, err := deleteRes.RowsAffected()
	if rowsCount == 0 {
		return fmt.Errorf("Delete DB exec error: nothing changed")
	}
	return nil
}

func (service *AgentService) GetSettingByID(ctx context.Context, modelID uint64) (*models.AgentSetting, error) {
	res := models.AgentSetting{}

	err := service.db.Get(&res, `
        SELECT id_setting, fk_agent, fk_agent_processed, fk_user, role, notifications_enabled, is_main_agent, updated_at, created_at
        FROM motify_agent_settings WHERE id_setting = ?
    `, modelID)
	return &res, err
}

func (service *AgentService) SetSetting(ctx context.Context, model *models.AgentSetting) (uint64, error) {
	if model.ID > 0 {
		return service.updateSetting(ctx, model)
	}
	return service.createSetting(ctx, model)
}

func (service *AgentService) createSetting(ctx context.Context, model *models.AgentSetting) (uint64, error) {
	if model.AgentFK == 0 {
		return 0, fmt.Errorf("Insert DB exec error: no fk_agent")
	}
	args := model.ToArgs()
	fkField := ""
	fkValue := ""
	if _, exists := args["fk_user"]; exists {
		fkField += "fk_user, "
		fkValue += ":fk_user, "
	}
	if _, exists := args["fk_agent_processed"]; exists {
		fkField += "fk_agent_processed, "
		fkValue += ":fk_agent_processed, "
	}
	sql := fmt.Sprintf(
		`INSERT INTO motify_agent_settings (fk_agent, %s role, notifications_enabled, is_main_agent) 
        VALUES (:fk_agent, %s :role, :notifications_enabled, :is_main_agent)`,
		fkField, fkValue)
	insertRes, err := service.db.Exec(sql, model.ToArgs())
	if err != nil {
		return 0, fmt.Errorf("Insert DB exec error: %v", err)
	}

	id, err := insertRes.LastInsertId()
	return uint64(id), err
}

func (service *AgentService) updateSetting(ctx context.Context, model *models.AgentSetting) (uint64, error) {
	if model.AgentFK == 0 {
		return 0, fmt.Errorf("Update DB exec error: no fk_agent")
	}
	args := model.ToArgs()
	fkField := ""
	if _, exists := args["fk_user"]; exists {
		fkField += "fk_user = :fk_user,"
	}
	if _, exists := args["fk_agent_processed"]; exists {
		fkField += "fk_agent_processed = :fk_agent_processed,"
	}
	sql := fmt.Sprintf(
		`UPDATE motify_agent_settings SET
                fk_agent = :fk_agent,
                %s
                role = :role,
                notifications_enabled = :notifications_enabled,
                is_main_agent = :is_main_agent
            WHERE
                id_setting = :id_setting`,
		fkField)
	updateRes, err := service.db.Exec(sql, args)
	if err != nil {
		return 0, fmt.Errorf("Update DB exec error: %v", err)
	}
	rowsCount, err := updateRes.RowsAffected()
	if rowsCount == 0 {
		return 0, fmt.Errorf("Update DB exec error: nothing changed")
	}
	return model.ID, nil
}

func (service *AgentService) DeleteSetting(ctx context.Context, modelID uint64) error {
	deleteRes, err := service.db.Exec(`
            DELETE FROM motify_agent_settings
            WHERE
                id_setting = :id_setting
        `, map[string]interface{}{
		"id_setting": modelID,
	})
	if err != nil {
		return fmt.Errorf("Insert DB exec error: %v", err)
	}
	rowsCount, err := deleteRes.RowsAffected()
	if rowsCount == 0 {
		return fmt.Errorf("Delete DB exec error: nothing changed")
	}
	return nil
}
