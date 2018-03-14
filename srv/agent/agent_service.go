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

func (service *AgentService) GetSettingsListByUserID(ctx context.Context, userID, limit, offset uint64) ([]*models.AgentWithSettings, error) {
	if limit == 0 || limit > defaultLimit {
		limit = defaultLimit
	}
	res := []*models.AgentWithSettings{}

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
