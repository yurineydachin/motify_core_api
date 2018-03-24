package payslip_list

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/service/logger"
)

type V1Args struct {
	UserID uint64  `key:"user_id" description:"User ID"`
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
	UpdateAt   string  `json:"updated_at"`
	CreatedAt  string  `json:"created_at"`
}

type V1ErrorTypes struct {
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "Payslip/Create/V1")
	cache.DisableTransportCache(ctx)

	limit := uint64(0)
	if opts.Limit != nil && *opts.Limit > 0 {
		limit = *opts.Limit
	}
	offset := uint64(0)
	if opts.Offset != nil && *opts.Offset > 0 {
		offset = *opts.Offset
	}

	payslips, err := handler.payslipService.GetListByUserID(ctx, opts.UserID, limit, offset)
	if err != nil {
		logger.Error(ctx, "Failed loading payslips by user %d: %v", opts.UserID, err)
		return nil, err
	}
	payslipsRes := make([]Payslip, 0, len(payslips))
	for i := range payslips {
		p := payslips[i]
		payslipsRes = append(payslipsRes, Payslip{
			ID:         p.ID,
			EmployeeFK: p.EmployeeFK,
			Title:      p.Title,
			Currency:   p.Currency,
			Amount:     p.Amount,
			UpdateAt:   p.UpdateAt,
			CreatedAt:  p.CreatedAt,
		})
	}

	return &V1Res{
		Payslips: payslipsRes,
	}, nil
}
