package integration_create

import (
	"context"
	"strings"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/service/logger"

	"motify_core_api/models"
)

type V1Args struct {
	Hash string `key:"hash" description:"Hash"`
}

type V1Res struct {
	Integration *Integration `json:"integration" description:"integration"`
}

type Integration struct {
	ID        uint64 `json:"id_integration"`
	Hash      string `json:"hash"`
	UpdatedAt string `json:"updated_at"`
	CreatedAt string `json:"created_at"`
}

type V1ErrorTypes struct {
	CREATE_FAILED           error `text:"creating integration is failed"`
	DUBLICATE_HASH          error `text:"dublicate hash"`
	INTEGRATION_NOT_CREATED error `text:"created integration not found"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "Integration/Create/V1")
	cache.DisableTransportCache(ctx)

	newIntegration := &models.Integration{
		Hash: opts.Hash,
	}

	integrationID, err := handler.integrationService.SetIntegration(ctx, newIntegration)
	if err != nil {
		logger.Error(ctx, "Failed creating integration: %v", err)
		if strings.Index(err.Error(), "Duplicate entry") > 0 {
			return nil, v1Errors.DUBLICATE_HASH
		}
		return nil, v1Errors.CREATE_FAILED
	}
	if integrationID == 0 {
		logger.Error(ctx, "Failed creating integration: integrationID is 0")
		return nil, v1Errors.CREATE_FAILED
	}

	integration, err := handler.integrationService.GetIntegrationByID(ctx, integrationID)
	if err != nil {
		logger.Error(ctx, "Failed login: %v", err)
		return nil, v1Errors.INTEGRATION_NOT_CREATED
	}
	if integration == nil {
		logger.Error(ctx, "Failed login: integration is nil")
		return nil, v1Errors.INTEGRATION_NOT_CREATED
	}

	return &V1Res{
		Integration: &Integration{
			ID:        integration.ID,
			Hash:      integration.Hash,
			UpdatedAt: integration.UpdatedAt,
			CreatedAt: integration.CreatedAt,
		},
	}, nil
}
