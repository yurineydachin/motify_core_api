package setting_list

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/service/logger"
)

type V1Args struct {
	IntegrationID uint64 `key:"integration_id" description:"Integration id"`
	UserID        uint64 `key:"user_id" description:"User id"`
}

type V1Res struct {
	List []ListItem `json:"list" description:"List of agents and employees"`
}

type ListItem struct {
	Agent   *Agent        `json:"agent" description:"Agent"`
	Setting *AgentSetting `json:"setting" description:"Setting"`
}

type Agent struct {
	ID            uint64 `json:"id_agent"`
	IntegrationFK uint64 `json:"fk_integration"`
	Name          string `json:"name"`
	CompanyID     string `json:"company_id"`
	Description   string `json:"description"`
	Logo          string `json:"logo"`
	Background    string `json:"bg_image"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	Site          string `json:"site"`
	UpdatedAt     string `json:"updated_at"`
	CreatedAt     string `json:"created_at"`
}

type AgentSetting struct {
	ID                    uint64  `json:"id_setting"`
	AgentFK               uint64  `json:"fk_agent"`
	AgentProcessedFK      *uint64 `json:"fk_agent_processed"`
	UserFK                *uint64 `json:"fk_user"`
	Role                  string  `json:"role"`
	IsNotificationEnabled bool    `json:"notifications_enabled"`
	IsMainAgent           bool    `json:"is_main_agent"`
	UpdatedAt             string  `json:"updated_at"`
	CreatedAt             string  `json:"created_at"`
}

type V1ErrorTypes struct {
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "Agent/List/V1")
	cache.DisableTransportCache(ctx)

	list, err := handler.agentService.GetAgentWithSettingsListByIntegrationIDAndUserID(ctx, opts.IntegrationID, opts.UserID)

	if err != nil {
		return nil, err
	}

	res := V1Res{
		List: make([]ListItem, len(list)),
	}
	for i := range list {
		agent := list[i].Agent
		res.List[i].Agent = &Agent{
			ID:            agent.ID,
			IntegrationFK: agent.IntegrationFK,
			Name:          agent.Name,
			CompanyID:     agent.CompanyID,
			Description:   agent.Description,
			Logo:          agent.Logo,
			Background:    agent.Background,
			Phone:         agent.Phone,
			Email:         agent.Email,
			Address:       agent.Address,
			Site:          agent.Site,
			UpdatedAt:     agent.UpdatedAt,
			CreatedAt:     agent.CreatedAt,
		}
		if list[i].AgentSetting != nil {
			setting := list[i].AgentSetting
			res.List[i].Setting = &AgentSetting{
				ID:               setting.ID,
				AgentFK:          setting.AgentFK,
				AgentProcessedFK: setting.AgentProcessedFK,
				UserFK:           setting.UserFK,
				Role:             setting.Role,
				IsNotificationEnabled: setting.IsNotificationEnabled,
				IsMainAgent:           setting.IsMainAgent,
				UpdatedAt:             setting.UpdatedAt,
				CreatedAt:             setting.CreatedAt,
			}
		}
	}

	return &res, nil
}
