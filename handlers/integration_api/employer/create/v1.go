package employer_create

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/mobapi_lib/token"
	"godep.lzd.co/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	CompanyID string `key:"compnay_id" description:"Agent company id"`
}

type V1Res struct {
	Agent   *Agent        `json:"agent" description:"Agent"`
	Setting *AgentSetting `json:"setting" description:"Setting"`
}

type Agent struct {
	id          uint64
	Hash        string `json:"hash"`
	Name        string `json:"name"`
	CompanyID   string `json:"company_id"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	Background  string `json:"bg_image"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	Site        string `json:"site"`
}

type AgentSetting struct {
	AgentProcessedHash    *string `json:"agent_processed_hash"`
	IsNotificationEnabled bool    `json:"notifications_enabled"`
	IsMainAgent           bool    `json:"is_employer"`
}

type V1ErrorTypes struct {
	SETTING_ADD_FAILED error `text:"Setting create failed"`
	AGENT_NOT_CREATED  error `text:"Error creating agent"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.IToken) (*V1Res, error) {
	logger.Debug(ctx, "Agent/Create/V1")
	cache.DisableTransportCache(ctx)

	userID := apiToken.GetID()
	integrationID := apiToken.GetExtraID()
	coreOpts := coreApiAdapter.SettingListV1Args{
		UserID:        userID,
		IntegrationID: integrationID,
	}
	listData, err := handler.coreApi.SettingListV1(ctx, coreOpts)
	if err != nil {
		return nil, err
	}
	agent, setting := findByCompanyID(listData.List, opts.CompanyID)
	if agent == nil {
		agentOptions := coreApiAdapter.AgentCreateV1Args{
			IntegrationFK: apiToken.GetExtraID(),
			CompanyID:     opts.CompanyID,
		}
		agentData, err := handler.coreApi.AgentCreateV1(ctx, agentOptions)
		if err != nil {
			if err.Error() == "MotifyCoreAPI: CREATE_FAILED" {
				return nil, v1Errors.AGENT_NOT_CREATED
			} else if err.Error() == "MotifyCoreAPI: AGENT_NOT_CREATED" {
				return nil, v1Errors.AGENT_NOT_CREATED
			}
			return nil, v1Errors.AGENT_NOT_CREATED
		}
		if agentData.Agent != nil {
			agent = convertAgentFromCreate(agentData.Agent)
		} else {
			return nil, v1Errors.AGENT_NOT_CREATED
		}
	}

	if setting == nil {
		coreOpts := coreApiAdapter.SettingCreateV1Args{
			AgentFK:               agent.id,
			UserFK:                &userID,
			IsNotificationEnabled: false,
			IsMainAgent:           true,
		}

		settingData, err := handler.coreApi.SettingCreateV1(ctx, coreOpts)
		if err != nil {
			if err.Error() == "MotifyCoreAPI: AGENT_NOT_FOUND" {
				return nil, v1Errors.SETTING_ADD_FAILED
			} else if err.Error() == "MotifyCoreAPI: SETTING_NOT_CREATED" {
				return nil, v1Errors.SETTING_ADD_FAILED
			} else if err.Error() == "MotifyCoreAPI: SETTING_ALREADY_EXISTS" {
				return nil, v1Errors.SETTING_ADD_FAILED
			} else if err.Error() == "MotifyCoreAPI: USER_NOT_FOUND" {
				return nil, v1Errors.SETTING_ADD_FAILED
			} else if err.Error() == "MotifyCoreAPI: CREATE_FAILED" {
				return nil, v1Errors.SETTING_ADD_FAILED
			}
			return nil, err
		}
		if settingData.User == nil || settingData.Agent == nil || settingData.Setting == nil {
			return nil, v1Errors.SETTING_ADD_FAILED
		}
		if settingData.Setting.UserFK == nil || *settingData.Setting.UserFK != userID {
			return nil, v1Errors.SETTING_ADD_FAILED
		}
		setting = convertSettingFromCreate(settingData.Setting, integrationID)
	}

	return &V1Res{
		Agent:   agent,
		Setting: setting,
	}, nil
}

func findByCompanyID(list []coreApiAdapter.SettingListListItem, companyID string) (*Agent, *AgentSetting) {
	for i := range list {
		if list[i].Agent != nil && list[i].Agent.CompanyID == companyID {
			return convertAgentFromList(list[i].Agent), convertSettingFromList(list[i].Setting, list[i].Agent.IntegrationFK)
		}
	}
	return nil, nil
}

func convertAgentFromList(agent *coreApiAdapter.SettingListAgent) *Agent {
	return &Agent{
		id:          agent.ID,
		Hash:        wrapToken.NewAgent(agent.ID, agent.IntegrationFK).Fixed().String(),
		Name:        agent.Name,
		CompanyID:   agent.CompanyID,
		Description: agent.Description,
		Logo:        agent.Logo,
		Background:  agent.Background,
		Phone:       agent.Phone,
		Email:       agent.Email,
		Address:     agent.Address,
		Site:        agent.Site,
	}
}

func convertAgentFromCreate(agent *coreApiAdapter.AgentCreateAgent) *Agent {
	return &Agent{
		id:          agent.ID,
		Hash:        wrapToken.NewAgent(agent.ID, agent.IntegrationFK).Fixed().String(),
		Name:        agent.Name,
		CompanyID:   agent.CompanyID,
		Description: agent.Description,
		Logo:        agent.Logo,
		Background:  agent.Background,
		Phone:       agent.Phone,
		Email:       agent.Email,
		Address:     agent.Address,
		Site:        agent.Site,
	}
}

func convertSettingFromList(s *coreApiAdapter.SettingListAgentSetting, integrationID uint64) *AgentSetting {
	if s == nil {
		return nil
	}
	return &AgentSetting{
		AgentProcessedHash:    getAgentHashPointer(s.AgentProcessedFK, integrationID),
		IsNotificationEnabled: s.IsNotificationEnabled,
		IsMainAgent:           s.IsMainAgent,
	}
}

func convertSettingFromCreate(s *coreApiAdapter.SettingCreateAgentSetting, integrationID uint64) *AgentSetting {
	return &AgentSetting{
		AgentProcessedHash:    getAgentHashPointer(s.AgentProcessedFK, integrationID),
		IsNotificationEnabled: s.IsNotificationEnabled,
		IsMainAgent:           s.IsMainAgent,
	}
}

func getAgentHashPointer(id *uint64, integrationID uint64) *string {
	if id != nil && *id > 0 {
		str := wrapToken.NewAgent(*id, integrationID).Fixed().String()
		return &str
	}
	return nil
}
