package agent_details

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/mobapi_lib/token"
	"motify_core_api/godep_libs/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	AgentHash string `key:"agent_hash" description:"Agent hash"`
}

type V1Res struct {
	Agent   *Agent        `json:"agent" description:"Agent"`
	Setting *AgentSetting `json:"setting" description:"Setting"`
}

type Agent struct {
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
	ERROR_PARSING_HASH error `text:"Error parsing hash"`
	AGENT_NOT_FOUND    error `text:"Agent not found"`
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

	t, err := wrapToken.ParseAgent(opts.AgentHash)
	if err != nil {
		logger.Error(ctx, "Error parse agent hash: ", err)
		return nil, v1Errors.ERROR_PARSING_HASH
	} else if t.GetExtraID() != integrationID {
		logger.Error(ctx, "Wrong agent hash (integration_id not equal): %d != %d", t.GetExtraID(), integrationID)
		return nil, v1Errors.ERROR_PARSING_HASH
	}
	agentID := t.GetID()

	coreOpts := coreApiAdapter.SettingListV1Args{
		UserID:        userID,
		IntegrationID: integrationID,
	}
	listData, err := handler.coreApi.SettingListV1(ctx, coreOpts)
	if err != nil {
		return nil, err
	}

	agent, setting := findAgentByID(listData.List, agentID)
	if agent == nil || setting == nil {
		return nil, v1Errors.AGENT_NOT_FOUND
	}

	return &V1Res{
		Agent:   agent,
		Setting: setting,
	}, nil
}

func findAgentByID(list []coreApiAdapter.SettingListListItem, id uint64) (*Agent, *AgentSetting) {
	for i := range list {
		if list[i].Agent != nil && list[i].Setting != nil && list[i].Agent.ID == id {
			return convertAgentFromList(list[i].Agent), convertSettingFromList(list[i].Setting, list[i].Agent.IntegrationFK)
		}
	}
	return nil, nil
}

func convertAgentFromList(agent *coreApiAdapter.SettingListAgent) *Agent {
	return &Agent{
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

func getAgentHashPointer(id *uint64, integrationID uint64) *string {
	if id != nil && *id > 0 {
		str := wrapToken.NewAgent(*id, integrationID).Fixed().String()
		return &str
	}
	return nil
}
