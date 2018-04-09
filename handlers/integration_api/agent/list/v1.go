package agent_list

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/mobapi_lib/token"
	"godep.lzd.co/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
}

type V1Res struct {
	List []ListItem `json:"list" description:"List of agents and employees"`
}

type ListItem struct {
	Agent   *Agent        `json:"agent" description:"Agent"`
	Setting *AgentSetting `json:"setting" description:"Setting"`
}

type Agent struct {
	Hash        string `json:"hash"`
	Name        string `json:"name"`
	CompanyID   string `json:"company_id"`
	Description string `json:"description"`
	Logo        string `json:"Logo"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	Site        string `json:"site"`
}

type AgentSetting struct {
	Hash                  string  `json:"hash"`
	AgentProcessedHash    *string `json:"agent_processed_hash"`
	Role                  string  `json:"role"`
	IsNotificationEnabled bool    `json:"notifications_enabled"`
	IsMainAgent           bool    `json:"is_employeer"`
}

type V1ErrorTypes struct {
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

	res := V1Res{
		List: make([]ListItem, 0, len(listData.List)),
	}
	for i := range listData.List {
		if listData.List[i].Agent != nil && listData.List[i].Setting != nil {
			res.List = append(res.List, ListItem{
				Agent:   convertAgentFromList(listData.List[i].Agent),
				Setting: convertSettingFromList(listData.List[i].Setting, integrationID),
			})
		}
	}
	return &res, nil
}

func convertAgentFromList(agent *coreApiAdapter.SettingListAgent) *Agent {
	return &Agent{
		Hash:        wrapToken.NewAgent(agent.ID, agent.IntegrationFK).Fixed().String(),
		Name:        agent.Name,
		CompanyID:   agent.CompanyID,
		Description: agent.Description,
		Logo:        agent.Logo,
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
		Hash:               wrapToken.NewSetting(s.ID, integrationID).Fixed().String(),
		AgentProcessedHash: getAgentHashPointer(s.AgentProcessedFK, integrationID),
		Role:               s.Role,
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
