package agent_sync

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/mobapi_lib/token"
	"motify_core_api/godep_libs/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	List []AgentArg `key:"agent_list" description:"Agent list"`
}

type AgentArg struct {
	Name        string  `key:"name" description:"Name"`
	CompanyID   string  `key:"company_id"  description:"Company id"`
	Description *string `key:"description" description:"Description"`
	Logo        *string `key:"logo" description:"Logo"`
	Background  *string `key:"bg_image" description:"Background image"`
	Phone       *string `key:"phone" description:"Phone"`
	Email       *string `key:"email" description:"Email"`
	Address     *string `key:"address" description:"Address"`
	Site        *string `key:"site" description:"Site"`
}

type V1Res struct {
	List []ListItem `json:"list" description:"List of agents and employees"`
}

type ListItem struct {
	Agent   *AgentStatus        `json:"agent" description:"Agent"`
	Setting *AgentSettingStatus `json:"setting" description:"Setting"`
}

type AgentStatus struct {
	Hash      string `json:"hash"`
	Name      string `json:"name"`
	CompanyID string `json:"company_id"`
	Status    string `json:"status"`
}

type AgentSettingStatus struct {
	Status string `json:"status"`
}

type V1ErrorTypes struct {
	AGENT_NOT_UPDATED error `text:"Error updating agent"`
	AGENT_NOT_FOUND   error `text:"Agent not found"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.IToken) (*V1Res, error) {
	logger.Debug(ctx, "Agent/Sync/V1")
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
	mAgents := mapAgents(listData.List)

	resList := make([]ListItem, len(listData.List))
	for i := range opts.List {
		resList[i] = ListItem{
			Agent: &AgentStatus{
				Name:      opts.List[i].Name,
				CompanyID: opts.List[i].CompanyID,
				Status:    "-",
			},
			Setting: &AgentSettingStatus{
				Status: "-",
			},
		}
		agentID := uint64(0)

		var setting *coreApiAdapter.SettingListAgentSetting
		if listItem, ok := mAgents[opts.List[i].CompanyID]; ok && listItem.Agent != nil {

			agentID = listItem.Agent.ID
			resList[i].Agent.Hash = wrapToken.NewAgent(listItem.Agent.ID, listItem.Agent.IntegrationFK).Fixed().String()
			agentOptions := coreApiAdapter.AgentUpdateV1Args{
				ID:            agentID,
				IntegrationFK: integrationID,
				Name:          &opts.List[i].Name,
				CompanyID:     &opts.List[i].CompanyID,
				Description:   opts.List[i].Description,
				Logo:          opts.List[i].Logo,
				Background:    opts.List[i].Background,
				Phone:         opts.List[i].Phone,
				Email:         opts.List[i].Email,
				Address:       opts.List[i].Address,
				Site:          opts.List[i].Site,
			}
			agentData, err := handler.coreApi.AgentUpdateV1(ctx, agentOptions)
			if err != nil {
				resList[i].Agent.Status = err.Error()
				continue
			}
			if agentData.Agent == nil {
				resList[i].Agent.Status = "AGENT_NOT_UPDATED"
				continue
			}
			resList[i].Agent.Status = "UPDATED"
			setting = listItem.Setting
		} else {
			agentOptions := coreApiAdapter.AgentCreateV1Args{
				IntegrationFK: integrationID,
				Name:          &opts.List[i].Name,
				CompanyID:     opts.List[i].CompanyID,
				Description:   opts.List[i].Description,
				Logo:          opts.List[i].Logo,
				Background:    opts.List[i].Background,
				Phone:         opts.List[i].Phone,
				Email:         opts.List[i].Email,
				Address:       opts.List[i].Address,
				Site:          opts.List[i].Site,
			}
			agentData, err := handler.coreApi.AgentCreateV1(ctx, agentOptions)
			if err != nil {
				resList[i].Agent.Status = err.Error()
				continue
			}
			if agentData.Agent == nil {
				resList[i].Agent.Status = "AGENT_NOT_CREATED"
				continue
			}
			agentID = agentData.Agent.ID
			resList[i].Agent.Hash = wrapToken.NewAgent(agentData.Agent.ID, agentData.Agent.IntegrationFK).Fixed().String()
			resList[i].Agent.Status = "CREATED"
		}

		if setting == nil && agentID > 0 {
			coreOpts := coreApiAdapter.SettingCreateV1Args{
				AgentFK:               agentID,
				UserFK:                &userID,
				IsNotificationEnabled: false,
				IsMainAgent:           false,
			}

			settingData, err := handler.coreApi.SettingCreateV1(ctx, coreOpts)
			if err != nil {
				resList[i].Setting.Status = err.Error()
				continue
			}
			if settingData.User == nil || settingData.Agent == nil || settingData.Setting == nil {
				resList[i].Setting.Status = "SETTING_ADD_FAILED"
				continue
			}
			if settingData.Setting.UserFK == nil || *settingData.Setting.UserFK != userID {
				resList[i].Setting.Status = "SETTING_ADD_FAILED"
				continue
			}
			resList[i].Setting.Status = "CREATED"
		}
	}
	return &V1Res{
		List: resList,
	}, nil
}

func mapAgents(list []coreApiAdapter.SettingListListItem) map[string]coreApiAdapter.SettingListListItem {
	res := make(map[string]coreApiAdapter.SettingListListItem, len(list))
	for i := range list {
		if list[i].Agent != nil {
			res[list[i].Agent.CompanyID] = list[i]
		}
	}
	return res
}
