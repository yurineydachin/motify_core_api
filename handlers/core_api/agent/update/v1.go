package agent_update

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/service/logger"
)

type V1Args struct {
	ID            uint64  `key:"id_agent" description:"Agent ID"`
	IntegrationFK uint64  `key:"fk_integration" description:"Integration ID"`
	Name          *string `key:"name" description:"Name"`
	CompanyID     *string `key:"company_id" description:"Company number"`
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
	AGENT_NOT_FOUND   error `text:"agent not found"`
	UPDATE_FAILED     error `text:"updating agent is failed"`
	AGENT_NOT_UPDATED error `text:"updated agent not found"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "Agent/Update/V1")
	cache.DisableTransportCache(ctx)

	agent, err := handler.agentService.GetAgentByID(ctx, opts.ID)
	if err != nil {
		logger.Error(ctx, "Fail loading agent: %v", err)
		return nil, v1Errors.AGENT_NOT_FOUND
	}
	if agent == nil {
		logger.Error(ctx, "Fail loading agent is nil")
		return nil, v1Errors.AGENT_NOT_FOUND
	}

	needUpdate := false
	if opts.CompanyID != nil && *opts.CompanyID != "" && *opts.CompanyID != agent.CompanyID {
		agent.CompanyID = *opts.CompanyID
		needUpdate = true
	}
	if opts.Name != nil && *opts.Name != "" && *opts.Name != agent.Name {
		agent.Name = *opts.Name
		needUpdate = true
	}
	if opts.Description != nil && *opts.Description != agent.Description {
		agent.Description = *opts.Description
		needUpdate = true
	}
	if opts.Logo != nil && *opts.Logo != agent.Logo {
		agent.Logo = *opts.Logo
		needUpdate = true
	}
	if opts.Background != nil && *opts.Background != agent.Background {
		agent.Background = *opts.Background
		needUpdate = true
	}
	if opts.Phone != nil && *opts.Phone != agent.Phone {
		agent.Phone = *opts.Phone
		needUpdate = true
	}
	if opts.Email != nil && *opts.Email != agent.Email {
		agent.Email = *opts.Email
		needUpdate = true
	}
	if opts.Address != nil && *opts.Address != agent.Address {
		agent.Address = *opts.Address
		needUpdate = true
	}
	if opts.Site != nil && *opts.Site != agent.Site {
		agent.Site = *opts.Site
		needUpdate = true
	}

	if needUpdate {
		agentID, err := handler.agentService.SetAgent(ctx, agent)
		if err != nil {
			logger.Error(ctx, "Failed updating agent: %v", err)
			return nil, v1Errors.UPDATE_FAILED
		}
		if agentID == 0 {
			logger.Error(ctx, "Failed updating agent: agentID is 0")
			return nil, v1Errors.UPDATE_FAILED
		}

		agent, err = handler.agentService.GetAgentByID(ctx, agentID)
		if err != nil {
			logger.Error(ctx, "Fail loading agent: %v", err)
			return nil, v1Errors.AGENT_NOT_FOUND
		}
		if agent == nil {
			logger.Error(ctx, "Fail loading agent is nil")
			return nil, v1Errors.AGENT_NOT_FOUND
		}
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
