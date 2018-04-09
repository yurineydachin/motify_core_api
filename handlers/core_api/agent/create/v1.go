package agent_create

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/service/logger"

	"motify_core_api/models"
)

type V1Args struct {
	IntegrationFK uint64  `key:"fk_integration" description:"Integration ID"`
	Name          *string `key:"name" description:"Name"`
	CompanyID     string  `key:"company_id" description:"Company number"`
	Description   *string `key:"description" description:"Long Description"`
	Logo          *string `key:"logo" description:"Logo url"`
	Background    *string `key:"bg_image" description:"Background image url"`
	Phone         *string `key:"phone" description:"Phone number"`
	Email         *string `key:"email" description:"Email"`
	Address       *string `key:"address" description:"Address"`
	Site          *string `key:"site" description:"Site"`
}

type V1Res struct {
	Agent *Agent `json:"agent" description:"Agent"`
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

type V1ErrorTypes struct {
	CREATE_FAILED     error `text:"creating agent is failed"`
	AGENT_NOT_CREATED error `text:"created agent not found"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "Agent/Create/V1")
	cache.DisableTransportCache(ctx)

	newAgent := &models.Agent{
		IntegrationFK: opts.IntegrationFK,
		CompanyID:     opts.CompanyID,
	}
	if opts.Name != nil && *opts.Name != "" {
		newAgent.Name = *opts.Name
	}
	if opts.Description != nil && *opts.Description != "" {
		newAgent.Description = *opts.Description
	}
	if opts.Logo != nil && *opts.Logo != "" {
		newAgent.Logo = *opts.Logo
	}
	if opts.Background != nil && *opts.Background != "" {
		newAgent.Background = *opts.Background
	}
	if opts.Phone != nil && *opts.Phone != "" {
		newAgent.Phone = *opts.Phone
	}
	if opts.Email != nil && *opts.Email != "" {
		newAgent.Email = *opts.Email
	}
	if opts.Address != nil && *opts.Address != "" {
		newAgent.Address = *opts.Address
	}
	if opts.Site != nil && *opts.Site != "" {
		newAgent.Site = *opts.Site
	}

	agentID, err := handler.agentService.SetAgent(ctx, newAgent)
	if err != nil {
		logger.Error(ctx, "Failed creating agent: %v", err)
		return nil, v1Errors.CREATE_FAILED
	}
	if agentID == 0 {
		logger.Error(ctx, "Failed creating agent: agentID is 0")
		return nil, v1Errors.CREATE_FAILED
	}

	agent, err := handler.agentService.GetAgentByID(ctx, agentID)
	if err != nil {
		logger.Error(ctx, "Failed login: %v", err)
		return nil, v1Errors.AGENT_NOT_CREATED
	}
	if agent == nil {
		logger.Error(ctx, "Failed login: agent is nil")
		return nil, v1Errors.AGENT_NOT_CREATED
	}

	return &V1Res{
		Agent: &Agent{
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
		},
	}, nil
}
