package setting_update

import (
	"context"
	"strings"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/service/logger"

	"motify_core_api/models"
)

type V1Args struct {
	ID                    *uint64 `key:"id_setting" description:"Setting ID"`
	AgentFK               *uint64 `key:"fk_agent" description:"Agent ID"`
	UserFK                *uint64 `key:"fk_user" description:"User ID"`
	AgentProcessedFK      *uint64 `key:"fk_agent_processed" description:"Agent processed ID"`
	Role                  *string `key:"role" description:"Role"`
	IsNotificationEnabled *bool   `key:"notifications_enabled" description:"Is notification enabled"`
	IsMainAgent           *bool   `key:"is_main_agent" description:"Is main agent"`
}

type V1Res struct {
	Agent          *Agent        `json:"agent" description:"Agent"`
	AgentProcessed *Agent        `json:"agent_processed" description:"Agent processed"`
	Setting        *AgentSetting `json:"setting" description:"Setting"`
	User           *User         `json:"user" description:"User"`
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

type User struct {
	ID          uint64 `json:"id_user"`
	Name        string `json:"name"`
	Short       string `json:"p_description"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}

type V1ErrorTypes struct {
	NOT_ENOUGH_PARAMS      error `text:"need params id_setting or fk_agent and fk_user to find setting"`
	AGENT_NOT_FOUND        error `text:"agent not found"`
	SETTING_NOT_FOUND      error `text:"setting not found"`
	USER_NOT_FOUND         error `text:"user not found"`
	UPDATE_FAILED          error `text:"updating setting is failed"`
	SETTING_NOT_UPDATED    error `text:"updated setting not found"`
	SETTING_ALREADY_EXISTS error `text:"setting already exists for this agent and user"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "Setting/Update/V1")
	cache.DisableTransportCache(ctx)

	var setting *models.AgentSetting
	var err error
	if opts.ID != nil && *opts.ID > 0 {
		setting, err = handler.agentService.GetSettingByID(ctx, *opts.ID)
	} else if opts.AgentFK != nil && *opts.AgentFK > 0 && opts.UserFK != nil && *opts.UserFK > 0 {
		setting, err = handler.agentService.GetSettingByAgentAndUser(ctx, *opts.AgentFK, *opts.UserFK)
	} else {
		return nil, v1Errors.NOT_ENOUGH_PARAMS
	}
	if err != nil {
		logger.Error(ctx, "Failed loading from DB: %v", err)
		return nil, v1Errors.SETTING_NOT_FOUND
	}
	if setting == nil {
		logger.Error(ctx, "Failed loading setting is nil")
		return nil, v1Errors.SETTING_NOT_FOUND
	}

	needUpdate := false
	agentID := setting.AgentFK
	if opts.AgentFK != nil && *opts.AgentFK > 0 && setting.AgentFK != *opts.AgentFK {
		agentID = *opts.AgentFK
		needUpdate = true
	}

	agent, err := handler.agentService.GetAgentByID(ctx, agentID)
	if err != nil {
		logger.Error(ctx, "Failed loading agent %d: %v", agentID, err)
		return nil, v1Errors.AGENT_NOT_FOUND
	}
	if agent == nil {
		logger.Error(ctx, "Failed loading agent (%d) is nil", agentID)
		return nil, v1Errors.AGENT_NOT_FOUND
	}
	setting.AgentFK = agent.ID

	agentProcessedID := uint64(0)
	if opts.AgentProcessedFK != nil && *opts.AgentProcessedFK > 0 {
		agentProcessedID = *opts.AgentProcessedFK
		if setting.AgentProcessedFK != nil && *setting.AgentProcessedFK != *opts.AgentProcessedFK {
			needUpdate = true
		}
	} else if setting.AgentProcessedFK != nil && *setting.AgentProcessedFK > 0 {
		agentProcessedID = *setting.AgentProcessedFK
	}

	var agentProcessedRes *Agent
	if agentProcessedID > 0 {
		agentProcessed, err := handler.agentService.GetAgentByID(ctx, agentProcessedID)
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
	userID := uint64(0)
	if opts.UserFK != nil && *opts.UserFK > 0 {
		userID = *opts.UserFK
		if setting.UserFK != nil && *setting.UserFK != *opts.UserFK {
			needUpdate = true
		}
	} else if setting.UserFK != nil && *setting.UserFK > 0 {
		userID = *setting.UserFK
	}

	if userID > 0 {
		user, err := handler.userService.GetUserByID(ctx, userID)
		if err != nil {
			logger.Error(ctx, "Failed loading %v", err)
			return nil, v1Errors.USER_NOT_FOUND
		}
		if user == nil {
			logger.Error(ctx, "Failed loading user is nil")
			return nil, v1Errors.USER_NOT_FOUND
		}
		setting.UserFK = &user.ID
		userRes = &User{
			ID:          user.ID,
			Name:        user.Name,
			Short:       user.Short,
			Description: user.Description,
			Avatar:      user.Avatar,
			Phone:       user.Phone,
			Email:       user.Email,
			UpdatedAt:   user.UpdatedAt,
			CreatedAt:   user.CreatedAt,
		}
	}

	if opts.Role != nil && *opts.Role != "" && setting.Role != *opts.Role {
		setting.Role = *opts.Role
		needUpdate = true
	}
	if opts.IsNotificationEnabled != nil && setting.IsNotificationEnabled != *opts.IsNotificationEnabled {
		setting.IsNotificationEnabled = *opts.IsNotificationEnabled
		needUpdate = true
	}
	if opts.IsMainAgent != nil && setting.IsMainAgent != *opts.IsMainAgent {
		setting.IsMainAgent = *opts.IsMainAgent
		needUpdate = true
	}

	if needUpdate {
		settingID, err := handler.agentService.SetSetting(ctx, setting)
		if err != nil {
			if strings.Index(err.Error(), "uniq_fk_agent_fk_user") > -1 {
				return nil, v1Errors.SETTING_ALREADY_EXISTS
			}
			logger.Error(ctx, "Failed updating setting: %v", err)
			return nil, v1Errors.UPDATE_FAILED
		}
		if settingID == 0 {
			logger.Error(ctx, "Failed updating setting: settingID is 0")
			return nil, v1Errors.UPDATE_FAILED
		}

		setting, err = handler.agentService.GetSettingByID(ctx, settingID)
		if err != nil {
			logger.Error(ctx, "Failed loading from DB: %v", err)
			return nil, v1Errors.SETTING_NOT_UPDATED
		}
		if agent == nil {
			logger.Error(ctx, "Failed loading setting is nil")
			return nil, v1Errors.SETTING_NOT_UPDATED
		}
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
}
