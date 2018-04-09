package employeer_update

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/mobapi_lib/token"
	"godep.lzd.co/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	Name        *string `key:"name" description:"Name"`
	CompanyID   *string `key:"company_id" description:"Company id"`
	Description *string `key:"description" description:"Description"`
	Logo        *string `key:"logo" description:"Logo"`
	Background  *string `key:"bg_image" description:"Background image"`
	Phone       *string `keyjson:"phone" description:"Phone"`
	Email       *string `key:"email" description:"Email"`
	Address     *string `key:"address" description:"Address"`
	Site        *string `key:"site" description:"Site"`
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
	IsMainAgent           bool    `json:"is_employeer"`
}

type V1ErrorTypes struct {
	MISSED_REQUIRED_FIELDS error `text:"Need company_id"`
	AGENT_NOT_UPDATED      error `text:"Error updating agent"`
	AGENT_NOT_CREATED      error `text:"Error creating agent"`
	SETTING_ADD_FAILED     error `text:"Setting create failed"`
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
	agent, setting := findMainCompany(listData.List)
	if agent == nil {
		if opts.CompanyID == nil || *opts.CompanyID == "" {
			return nil, v1Errors.MISSED_REQUIRED_FIELDS
		}
		agentOptions := coreApiAdapter.AgentCreateV1Args{
			IntegrationFK: integrationID,
			Name:          opts.Name,
			CompanyID:     *opts.CompanyID,
			Description:   opts.Description,
			Logo:          opts.Logo,
			Background:    opts.Background,
			Phone:         opts.Phone,
			Email:         opts.Email,
			Address:       opts.Address,
			Site:          opts.Site,
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
		if agentData.Agent == nil {
			return nil, v1Errors.AGENT_NOT_CREATED
		}
		agent = convertAgentFromCreate(agentData.Agent)
	} else {
		agentOptions := coreApiAdapter.AgentUpdateV1Args{
			ID:            agent.id,
			IntegrationFK: integrationID,
			Name:          opts.Name,
			CompanyID:     opts.CompanyID,
			Description:   opts.Description,
			Logo:          opts.Logo,
			Background:    opts.Background,
			Phone:         opts.Phone,
			Email:         opts.Email,
			Address:       opts.Address,
			Site:          opts.Site,
		}
		agentData, err := handler.coreApi.AgentUpdateV1(ctx, agentOptions)
		if err != nil {
			if err.Error() == "MotifyCoreAPI: CREATE_FAILED" {
				return nil, v1Errors.AGENT_NOT_UPDATED
			} else if err.Error() == "MotifyCoreAPI: UPDATE_FAILED" {
				return nil, v1Errors.AGENT_NOT_UPDATED
			} else if err.Error() == "MotifyCoreAPI: AGENT_NOT_FOUND" {
				return nil, v1Errors.AGENT_NOT_UPDATED
			}
			return nil, v1Errors.AGENT_NOT_UPDATED
		}
		if agentData.Agent == nil {
			return nil, v1Errors.AGENT_NOT_UPDATED
		}
		agent = convertAgentFromUpdate(agentData.Agent)
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

func findMainCompany(list []coreApiAdapter.SettingListListItem) (*Agent, *AgentSetting) {
	for i := range list {
		if list[i].Agent != nil && list[i].Setting != nil && list[i].Setting.IsMainAgent {
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

func convertAgentFromUpdate(agent *coreApiAdapter.AgentUpdateAgent) *Agent {
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
