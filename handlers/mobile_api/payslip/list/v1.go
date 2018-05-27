package payslip_list

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/mobapi_lib/token"
	"motify_core_api/godep_libs/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	Limit     *uint64 `key:"limit" description:"Limit"`
	Offset    *uint64 `key:"offset" description:"Offset"`
	DateAfter *uint64 `key:"date_after" description:"Way to get only new payslips with created_at after this date"`
}

type V1Res struct {
	List []Payslip `json:"list" description:"List of payslips"`
}

type Payslip struct {
	Hash      string  `json:"hash"`
	Title     string  `json:"title"`
	Logo      string  `json:"logo"`
	Currency  string  `json:"currency"`
	Amount    float64 `json:"amount"`
	UpdatedAt string  `json:"updated_at"`
	CreatedAt string  `json:"created_at"`
}

type V1ErrorTypes struct {
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.IToken) (*V1Res, error) {
	logger.Debug(ctx, "Payslip/List/V1")
	cache.DisableTransportCache(ctx)

	coreOpts := coreApiAdapter.PayslipListV1Args{
		UserID: uint64(apiToken.GetID()),
		Limit:  opts.Limit,
		Offset: opts.Offset,
		//DateAfter: opts.DateAfter,
	}
	data, err := handler.coreApi.PayslipListV1(ctx, coreOpts)
	if err != nil {
		return nil, err
	}

	res := V1Res{
		List: make([]Payslip, 0, len(data.List)),
	}
	for i := range data.List {
		agent := data.List[i].Agent
		p := data.List[i].Payslip
		res.List = append(res.List, Payslip{
			Hash:      wrapToken.NewPayslip(p.ID, agent.IntegrationFK).Fixed().String(),
			Title:     p.Title,
			Logo:      agent.Logo,
			Currency:  p.Currency,
			Amount:    p.Amount,
			UpdatedAt: p.UpdatedAt,
			CreatedAt: p.CreatedAt,
		})
	}

	return &res, nil
}
