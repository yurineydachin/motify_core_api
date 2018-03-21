package setting_create

import (
	"context"
	"strings"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/service/logger"

	"motify_core_api/models"
)

type V1Args struct {
	AgentFK               uint64  `key:"fk_agent" description:"Agent ID"`
	UserFK                *uint64 `key:"fk_user" description:"User ID"`
	AgentProcessedFK      *uint64 `key:"fk_agent_processed" description:"Agent processed ID"`
	Role                  string  `key:"role" description:"Role"`
	IsNotificationEnabled bool    `key:"notifications_enabled" description:"Is notification enabled"`
	IsMainAgent           bool    `key:"is_main_agent" description:"Is main agent"`
}

type V1Res struct {
	Agent          *Agent        `json:"agent" description:"Agent"`
	AgentProcessed *Agent        `json:"agent_processed" description:"Agent processed"`
	Setting        *AgentSetting `json:"setting" description:"Setting"`
	User           *User         `json:"user" description:"User"`
}

type Agent struct {
	ID          uint64 `json:"id_agent"`
	Name        string `json:"name"`
	CompanyID   string `json:"company_id"`
	Description string `json:"description"`
	Logo        string `json:"Logo"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	Site        string `json:"site"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
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

type User struct {
	ID          uint64 `json:"id_user"`
	Name        string `json:"name"`
	Short       string `json:"p_description"`
	Description string `json:"description"`
	Awatar      string `json:"awatar"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}

type V1ErrorTypes struct {
	AGENT_NOT_FOUND        error `text:"agent not found"`
	USER_NOT_FOUND         error `text:"user not found"`
	CREATE_FAILED          error `text:"creating setting is failed"`
	SETTING_NOT_CREATED    error `text:"created setting not found"`
	SETTING_ALREADY_EXISTS error `text:"setting already exists for this agent and user"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "Setting/Create/V1")
	cache.DisableTransportCache(ctx)

	agent, err := handler.agentService.GetAgentByID(ctx, opts.AgentFK)
	if err != nil {
		logger.Error(ctx, "Failed login: %v", err)
		return nil, v1Errors.AGENT_NOT_FOUND
	}
	if agent == nil {
		logger.Error(ctx, "Failed login: agent is nil")
		return nil, v1Errors.AGENT_NOT_FOUND
	}

	var agentProcessedRes *Agent
	if opts.AgentProcessedFK != nil && *opts.AgentProcessedFK > 0 {
		agentProcessed, err := handler.agentService.GetAgentByID(ctx, *opts.AgentProcessedFK)
		if err != nil {
			logger.Error(ctx, "Failed login: %v", err)
			return nil, v1Errors.AGENT_NOT_FOUND
		}
		if agentProcessed == nil {
			logger.Error(ctx, "Failed login: agent is nil")
			return nil, v1Errors.AGENT_NOT_FOUND
		}
		agentProcessedRes = convertAgent(agentProcessed)
	}

	var userRes *User
	if opts.UserFK != nil && *opts.UserFK > 0 {
		user, err := handler.userService.GetUserByID(ctx, *opts.UserFK)
		if err != nil {
			logger.Error(ctx, "Failed login: %v", err)
			return nil, v1Errors.USER_NOT_FOUND
		}
		if user == nil {
			logger.Error(ctx, "Failed login: user is nil")
			return nil, v1Errors.USER_NOT_FOUND
		}
		userRes = &User{
			ID:          user.ID,
			Name:        user.Name,
			Short:       user.Short,
			Description: user.Description,
			Awatar:      user.Awatar,
			Phone:       user.Phone,
			Email:       user.Email,
			UpdatedAt:   user.UpdatedAt,
			CreatedAt:   user.CreatedAt,
		}
	}

	setting := &models.AgentSetting{
		AgentFK:          opts.AgentFK,
		AgentProcessedFK: opts.AgentProcessedFK,
		UserFK:           opts.UserFK,
		Role:             opts.Role,
		IsNotificationEnabled: opts.IsNotificationEnabled,
		IsMainAgent:           opts.IsNotificationEnabled,
	}

	settingID, err := handler.agentService.SetSetting(ctx, setting)
	if err != nil {
		if strings.Index(err.Error(), "uniq_fk_agent_fk_user") > -1 {
			return nil, v1Errors.SETTING_ALREADY_EXISTS
		}
		logger.Error(ctx, "Failed creating setting: %v", err)
		return nil, v1Errors.CREATE_FAILED
	}
	if settingID == 0 {
		logger.Error(ctx, "Failed creating setting: settingID is 0")
		return nil, v1Errors.CREATE_FAILED
	}

	setting, err = handler.agentService.GetSettingByID(ctx, settingID)
	if err != nil {
		logger.Error(ctx, "Failed loading from DB: %v", err)
		return nil, v1Errors.SETTING_NOT_CREATED
	}
	if setting == nil {
		logger.Error(ctx, "Failed login: setting is nil")
		return nil, v1Errors.SETTING_NOT_CREATED
	}

	return &V1Res{
		Agent:          convertAgent(agent),
		AgentProcessed: agentProcessedRes,
		Setting: &AgentSetting{
			ID:               setting.ID,
			AgentFK:          setting.AgentFK,
			AgentProcessedFK: setting.AgentProcessedFK,
			UserFK:           setting.UserFK,
			Role:             setting.Role,
			IsNotificationEnabled: setting.IsNotificationEnabled,
			IsMainAgent:           setting.IsMainAgent,
			UpdatedAt:             setting.UpdatedAt,
			CreatedAt:             setting.CreatedAt,
		},
		User: userRes,
	}, nil
}

func convertAgent(agent *models.Agent) *Agent {
	return &Agent{
		ID:          agent.ID,
		Name:        agent.Name,
		CompanyID:   agent.CompanyID,
		Description: agent.Description,
		Logo:        agent.Logo,
		Phone:       agent.Phone,
		Email:       agent.Email,
		Address:     agent.Address,
		Site:        agent.Site,
		UpdatedAt:   agent.UpdatedAt,
		CreatedAt:   agent.CreatedAt,
	}
}
