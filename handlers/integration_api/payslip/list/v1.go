package payslip_list

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/mobapi_lib/token"
	"godep.lzd.co/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	EmployeeHash string  `key:"employee_hash" description:"Employee hash"`
	Limit        *uint64 `key:"limit" description:"Limit"`
	Offset       *uint64 `key:"offset" description:"Offset"`
}

type V1Res struct {
	List []ListItem `json:"list" description:"List of agents and employees"`
}

type ListItem struct {
	Payslip Payslip `json:"payslip" description:"Payslip"`
}

type Payslip struct {
	Hash      string  `json:"hash"`
	Title     string  `json:"title"`
	Currency  string  `json:"currency"`
	Amount    float64 `json:"amount"`
	UpdatedAt string  `json:"updated_at"`
	CreatedAt string  `json:"created_at"`
}

type V1ErrorTypes struct {
	ERROR_PARSING_HASH error `text:"Error parsing hash"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.IToken) (*V1Res, error) {
	logger.Debug(ctx, "Payslip/List/V1")
	cache.DisableTransportCache(ctx)

	integrationID := apiToken.GetExtraID()

	t, err := wrapToken.ParseEmployee(opts.EmployeeHash)
	employeeID := t.GetID()
	if err != nil {
		logger.Error(ctx, "Error parse employee hash: ", err)
		return nil, v1Errors.ERROR_PARSING_HASH
	} else if t.GetExtraID() != integrationID {
		logger.Error(ctx, "Wrong employee hash (integration_id not equal): %d != %d", t.GetExtraID(), integrationID)
		return nil, v1Errors.ERROR_PARSING_HASH
	} else if employeeID == 0 {
		logger.Error(ctx, "Wrong employee hash (employeeID = 0)")
		return nil, v1Errors.ERROR_PARSING_HASH
	}

	coreOpts := coreApiAdapter.PayslipListByEmployeeV1Args{
		EmployeeID: employeeID,
		Limit:      opts.Limit,
		Offset:     opts.Offset,
	}
	data, err := handler.coreApi.PayslipListByEmployeeV1(ctx, coreOpts)
	if err != nil {
		return nil, err
	}

	res := V1Res{
		List: make([]ListItem, 0, len(data.List)),
	}
	for i := range data.List {
		p := data.List[i].Payslip
		res.List = append(res.List, ListItem{
			Payslip: Payslip{
				Hash:      wrapToken.NewPayslip(p.ID, integrationID).Fixed().String(),
				Title:     p.Title,
				Currency:  p.Currency,
				Amount:    p.Amount,
				UpdatedAt: p.UpdatedAt,
				CreatedAt: p.CreatedAt,
			},
		})
	}

	return &res, nil
}
