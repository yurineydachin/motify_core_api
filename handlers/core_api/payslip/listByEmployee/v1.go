package payslip_list_by_employee

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/service/logger"
)

type V1Args struct {
	EmployeeID uint64  `key:"employee_id" description:"Employee ID"`
	Limit      *uint64 `key:"limit" description:"Limit"`
	Offset     *uint64 `key:"offset" description:"Offset"`
}

type V1Res struct {
	List []ListItem `json:"list" description:"List"`
}

type ListItem struct {
	Payslip Payslip `json:"payslip" description:"Payslip"`
}

type Payslip struct {
	ID         uint64  `json:"id_payslip"`
	EmployeeFK uint64  `json:"fk_employee"`
	Title      string  `json:"title"`
	Currency   string  `json:"currency"`
	Amount     float64 `json:"amount"`
	UpdatedAt  string  `json:"updated_at"`
	CreatedAt  string  `json:"created_at"`
}

type V1ErrorTypes struct {
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args) (*V1Res, error) {
	logger.Debug(ctx, "Payslip/ListByEmployee/V1")
	cache.DisableTransportCache(ctx)

	limit := uint64(0)
	if opts.Limit != nil && *opts.Limit > 0 {
		limit = *opts.Limit
	}
	offset := uint64(0)
	if opts.Offset != nil && *opts.Offset > 0 {
		offset = *opts.Offset
	}

	list, err := handler.payslipService.GetListByEmployeeID(ctx, opts.EmployeeID, limit, offset)
	if err != nil {
		logger.Error(ctx, "Failed loading payslips by employee %d: %v", opts.EmployeeID, err)
		return nil, err
	}
	res := V1Res{
		List: make([]ListItem, 0, len(list)),
	}
	for i := range list {
		p := list[i]
		res.List = append(res.List, ListItem{
			Payslip: Payslip{
				ID:         p.ID,
				EmployeeFK: p.EmployeeFK,
				Title:      p.Title,
				Currency:   p.Currency,
				Amount:     p.Amount,
				UpdatedAt:  p.UpdatedAt,
				CreatedAt:  p.CreatedAt,
			},
		})
	}

	return &res, nil
}
