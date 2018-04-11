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

type AgentWithSetting struct {
	*models.Agent
	*models.AgentSetting
}

type AgentWithLeftSetting struct {
	*models.Agent
	*models.AgentSettingLeft
}

type AgentWithEmployee struct {
	*models.Agent
	*models.Employee
}

func NewAgentService(db *database.DbAdapter) *AgentService {
	return &AgentService{
		db: db,
	}
}

func (service *AgentService) GetAgentWithSettingsListByIntegrationIDAndUserID(ctx context.Context, integrationID, userID uint64) ([]AgentWithSetting, error) {
	list := []AgentWithLeftSetting{}

	err := service.db.Select(&list, `
        SELECT
            id_agent, a_fk_integration, a_name, a_company_id, a_description, a_logo, a.a_bg_image, a_address, a_phone, a_email, a_site, a_updated_at, a_created_at,
            id_setting, s_fk_agent, s_fk_agent_processed, s_fk_user, s_role, s_notifications_enabled, s_is_main_agent, s_updated_at, s_created_at
        FROM motify_agents a
        LEFT JOIN motify_agent_settings s ON s.s_fk_agent = a.id_agent AND s.s_fk_user = ?
        WHERE a_fk_integration = ?
        ORDER BY s_created_at DESC
    `, userID, integrationID)
	if err != nil {
		return nil, err
	}

	res := make([]AgentWithSetting, len(list))
	for i := range list {
		res[i].Agent = list[i].Agent
		res[i].AgentSetting = list[i].AgentSettingLeft.ToAgentSetting()
	}
	return res, err
}

func (service *AgentService) GetSettingsListByUserID(ctx context.Context, userID, limit, offset uint64) ([]AgentWithSetting, error) {
	if limit == 0 || limit > defaultLimit {
		limit = defaultLimit
	}
	res := []AgentWithSetting{}

	err := service.db.Select(&res, `
        SELECT
            id_agent, a_fk_integration, a_name, a_company_id, a_description, a_logo, a.a_bg_image, a_address, a_phone, a_email, a_site, a_updated_at, a_created_at,
            id_setting, s_fk_agent, s_fk_agent_processed, s_fk_user, s_role, s_notifications_enabled, s_is_main_agent, s_updated_at, s_created_at
        FROM motify_agents a
        INNER JOIN motify_agent_settings s ON s.s_fk_agent = a.id_agent
        WHERE s.s_fk_user = ?
        ORDER BY s.s_created_at DESC
        LIMIT ?
        OFFSET ?
    `, userID, limit, offset)
	return res, err
}

func (service *AgentService) GetEmployeeListByUserID(ctx context.Context, userID, limit, offset uint64) ([]AgentWithEmployee, error) {
	if limit == 0 || limit > defaultLimit {
		limit = defaultLimit
	}
	res := []AgentWithEmployee{}

	err := service.db.Select(&res, `
        SELECT
            id_agent, a_fk_integration, a_name, a_company_id, a_description, a_logo, a.a_bg_image, a_address, a_phone, a_email, a_site, a_updated_at, a_created_at,
            id_employee, e_fk_agent, e_fk_user, e_code, e_name, e_email, e_hire_date, e_number_of_dependants, e_gross_base_salary, e_role, e_updated_at, e_created_at
        FROM motify_agents a
        INNER JOIN motify_agent_employees e ON e.e_fk_agent = a.id_agent
        WHERE e.e_fk_user = ?
        ORDER BY e.e_created_at DESC
        LIMIT ?
        OFFSET ?
    `, userID, limit, offset)
	return res, err
}

func (service *AgentService) GetAgentByID(ctx context.Context, modelID uint64) (*models.Agent, error) {
	res := models.Agent{}
	err := service.db.Get(&res, `
        SELECT id_agent, a_fk_integration, a_name, a_company_id, a_description, a_logo, a_bg_image, a_address, a_phone, a_email, a_site, a_updated_at, a_created_at
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
            INSERT INTO motify_agents (a_name, a_fk_integration, a_company_id, a_description, a_logo, a_bg_image, a_address, a_phone, a_email, a_site)
            VALUES (:a_name, :a_fk_integration, :a_company_id, :a_description, :a_logo, :a_bg_image, :a_address, :a_phone, :a_email, :a_site)
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
                a_fk_integration = :a_fk_integration,
                a_name = :a_name,
                a_company_id = :a_company_id,
                a_description = :a_description,
                a_logo = :a_logo,
                a_bg_image = :a_bg_image,
                a_address =  :a_address,
                a_phone = :a_phone,
                a_email = :a_email,
                a_site = :a_site
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

func (service *AgentService) GetEmployeeListByAgentID(ctx context.Context, agentID, limit, offset uint64) ([]*models.Employee, error) {
	if limit == 0 || limit > defaultLimit {
		limit = defaultLimit
	}
	res := []*models.Employee{}

	err := service.db.Select(&res, `
        SELECT
            id_employee, e_fk_agent, e_fk_user, e_code, e_name, e_email, e_hire_date, e_number_of_dependants, e_gross_base_salary, e_role, e_updated_at, e_created_at
        FROM motify_agent_employees
        WHERE e_fk_agent = ?
        ORDER BY e_name DESC
        LIMIT ?
        OFFSET ?
    `, agentID, limit, offset)
	return res, err
}

func (service *AgentService) GetEmployeeByAgentAndMobileUser(ctx context.Context, agentFK, userFK uint64) (*models.Employee, error) {
	res := models.Employee{}
	err := service.db.Get(&res, `
        SELECT id_employee, e_fk_agent, e_fk_user, e_code, e_name, e_email, e_hire_date, e_number_of_dependants, e_gross_base_salary, e_role, e_updated_at, e_created_at
        FROM motify_agent_employees WHERE e_fk_agent = ? AND e_fk_user = ?
    `, agentFK, userFK)
	return &res, err
}

func (service *AgentService) GetEmployeeByAgentAndEmploeeCode(ctx context.Context, agentFK uint64, code string) (*models.Employee, error) {
	res := models.Employee{}
	err := service.db.Get(&res, `
        SELECT id_employee, e_fk_agent, e_fk_user, e_code, e_name, e_email, e_hire_date, e_number_of_dependants, e_gross_base_salary, e_role, e_updated_at, e_created_at
        FROM motify_agent_employees WHERE e_fk_agent = ? AND e_code = ?
    `, agentFK, code)
	return &res, err
}

func (service *AgentService) GetEmployeeByCompanyIDAndEmploeeCode(ctx context.Context, integrationFK uint64, companyID, code string) (*models.Employee, error) {
	res := models.Employee{}
	err := service.db.Get(&res, `
        SELECT
            id_employee, e_fk_agent, e_fk_user, e_code, e_name, e_email, e_hire_date, e_number_of_dependants, e_gross_base_salary, e_role, e_updated_at, e_created_at
        FROM motify_agent_employees e
        INNER JOIN motify_agent a ON e.e_fk_agent = a.id_agent
        WHERE
            a_fk_integration = ? AND
            a_company_id = ? AND
            e_code = ?
    `, integrationFK, companyID, code)
	return &res, err
}

func (service *AgentService) GetEmployeeByID(ctx context.Context, modelID uint64) (*models.Employee, error) {
	res := models.Employee{}
	err := service.db.Get(&res, `
        SELECT id_employee, e_fk_agent, e_fk_user, e_code, e_name, e_email, e_hire_date, e_number_of_dependants, e_gross_base_salary, e_role, e_updated_at, e_created_at
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
	if _, exists := args["e_fk_user"]; exists {
		fkField = "e_fk_user, "
		fkValue = ":e_fk_user, "
	}
	sql := fmt.Sprintf(
		`INSERT INTO motify_agent_employees (e_fk_agent, %s e_code, e_name, e_email, e_hire_date, e_number_of_dependants, e_gross_base_salary, e_role) 
        VALUES (:e_fk_agent, %s :e_code, :e_name, :e_email, :e_hire_date, :e_number_of_dependants, :e_gross_base_salary, :e_role)`,
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
	if _, exists := args["e_fk_user"]; exists {
		fkField = "e_fk_user = :e_fk_user,"
	}
	sql := fmt.Sprintf(
		`UPDATE motify_agent_employees SET
                e_fk_agent = :e_fk_agent,
                %s
                e_code = :e_code,
                e_name = :e_name,
                e_email = :e_email,
                e_hire_date = :e_hire_date,
                e_number_of_dependants = :e_number_of_dependants,
                e_gross_base_salary = :e_gross_base_salary,
                e_role = :e_role
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
        SELECT id_setting, s_fk_agent, s_fk_agent_processed, s_fk_user, s_role, s_notifications_enabled, s_is_main_agent, s_updated_at, s_created_at
        FROM motify_agent_settings WHERE id_setting = ?
    `, modelID)
	return &res, err
}

func (service *AgentService) GetSettingByAgentAndUser(ctx context.Context, agentFK, userFK uint64) (*models.AgentSetting, error) {
	res := models.AgentSetting{}

	err := service.db.Get(&res, `
        SELECT id_setting, s_fk_agent, s_fk_agent_processed, s_fk_user, s_role, s_notifications_enabled, s_is_main_agent, s_updated_at, s_created_at
        FROM motify_agent_settings WHERE s_fk_agent = ? AND s_fk_user = ?
    `, agentFK, userFK)
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
	if _, exists := args["s_fk_user"]; exists {
		fkField += "s_fk_user, "
		fkValue += ":s_fk_user, "
	}
	if _, exists := args["s_fk_agent_processed"]; exists {
		fkField += "s_fk_agent_processed, "
		fkValue += ":s_fk_agent_processed, "
	}
	sql := fmt.Sprintf(
		`INSERT INTO motify_agent_settings (s_fk_agent, %s s_role, s_notifications_enabled, s_is_main_agent) 
        VALUES (:s_fk_agent, %s :s_role, :s_notifications_enabled, :s_is_main_agent)`,
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
	if _, exists := args["s_fk_user"]; exists {
		fkField += "s_fk_user = :s_fk_user,"
	}
	if _, exists := args["s_fk_agent_processed"]; exists {
		fkField += "s_fk_agent_processed = :s_fk_agent_processed,"
	}
	sql := fmt.Sprintf(
		`UPDATE motify_agent_settings SET
                s_fk_agent = :s_fk_agent,
                %s
                s_role = :s_role,
                s_notifications_enabled = :s_notifications_enabled,
                s_is_main_agent = :s_is_main_agent
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
