package payslip_list

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/mobapi_lib/token"
	"godep.lzd.co/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
)

type V1Args struct {
	Limit  *uint64 `key:"limit" description:"Limit"`
	Offset *uint64 `key:"offset" description:"Offset"`
}

type V1Res struct {
	Payslips []Payslip `json:"payslips" description:"Payslips"`
}

type Payslip struct {
	ID         uint64  `json:"id_payslip"`
	EmployeeFK uint64  `json:"fk_employee"`
	Title      string  `json:"title"`
	Currency   string  `json:"currency"`
	Amount     float64 `json:"amount"`
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
		UserID: uint64(apiToken.GetCustomerID()),
		Limit:  opts.Limit,
		Offset: opts.Offset,
	}
	data, err := handler.coreApi.PayslipListV1(ctx, coreOpts)
	if err != nil {
		return nil, err
	}

	payslipsRes := make([]Payslip, 0, len(data.Payslips))
	for i := range data.Payslips {
		p := data.Payslips[i]
		payslipsRes = append(payslipsRes, Payslip{
			ID:         p.ID,
			EmployeeFK: p.EmployeeFK,
			Title:      p.Title,
			Currency:   p.Currency,
			Amount:     p.Amount,
		})
	}

	return &V1Res{
		Payslips: payslipsRes,
	}, nil
}
