package integration_check

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/service/logger"
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
	INTEGRATION_NOT_FOUND error `text:"integration not found"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "Integration/Check/V1")
	cache.DisableTransportCache(ctx)

	integration, err := handler.integrationService.GetIntegrationByHash(ctx, opts.Hash)
	if err != nil {
		logger.Error(ctx, "Failed login: %v", err)
		return nil, v1Errors.INTEGRATION_NOT_FOUND
	}
	if integration == nil {
		logger.Error(ctx, "Failed login: integration is nil")
		return nil, v1Errors.INTEGRATION_NOT_FOUND
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
