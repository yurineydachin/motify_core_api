package payslip_details

import (
	"context"
	"encoding/json"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"motify_core_api/godep_libs/mobapi_lib/token"
	"motify_core_api/godep_libs/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

type V1Args struct {
	PayslipHash string `key:"payslip_hash" description:"Payslip hash"`
}

type V1Res struct {
	Payslip Payslip `json:"payslip"`
}

type Payslip struct {
	Hash        string      `json:"hash"`
	Title       string      `json:"title"`
	Currency    string      `json:"currency"`
	Amount      float64     `json:"amount"`
	UpdatedAt   string      `json:"updated_at"`
	CreatedAt   string      `json:"created_at"`
	Transaction Transaction `json:"transaction"`
	Sections    []Section   `json:"sections"`
	Footnote    string      `json:"footnote,omitempty"`
}

type PayslipData struct {
	Transaction Transaction `json:"transaction"`
	Sections    []Section   `json:"sections"`
	Footnote    string      `json:"footnote,omitempty"`
}

type Transaction struct {
	Description string               `json:"description"`
	Sections    []TransactionSection `json:"sections"`
}

type TransactionSection struct {
	Title string `json:"title"`
	Rows  []Row  `json:"rows"`
}

type Section struct {
	Type       string   `json:"section_type"`
	Title      string   `json:"title"`
	Term       string   `json:"term,omitempty"`
	Definition string   `json:"definition,omitempty"`
	Amount     *float64 `json:"amount,omitempty"`
	Rows       []Row    `json:"rows"`
}

type Row struct {
	Type        string   `json:"row_type"`
	Title       string   `json:"title"`
	Term        string   `json:"term,omitempty"`
	Description string   `json:"description,omitempty"`
	Footnote    string   `json:"footnote,omitempty"`
	Role        string   `json:"role,omitempty"`
	Avatar      string   `json:"avatar_image,omitempty"`
	BGImage     string   `json:"bg_image,omitempty"`
	Amount      *float64 `json:"amount,omitempty"`
	Float       *float64 `json:"float,omitempty"`
	Int         *int64   `json:"int,omitempty"`
	Text        string   `json:"text,omitempty"`
	DateFrom    string   `json:"date_from,omitempty"`
	DateTo      string   `json:"date_to,omitempty"`
	Children    []Row    `json:"rows,omitempty"`
}

type V1ErrorTypes struct {
	AGENT_NOT_FOUND       error `text:"agent not found"`
	EMPLOYEE_NOT_FOUND    error `text:"employee not found"`
	PAYSLIP_NOT_FOUND     error `text:"payslip not found"`
	ERROR_PARSING_PAYSLIP error `text:"error parsing payslip"`
	ERROR_PARSING_HASH    error `text:"Error parsing hash"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.IToken) (*V1Res, error) {
	logger.Debug(ctx, "Payslip/List/V1")
	cache.DisableTransportCache(ctx)

	t, err := wrapToken.ParsePayslip(opts.PayslipHash)
	if err != nil {
		logger.Error(ctx, "Error parse payslip hash: ", err)
		return nil, v1Errors.ERROR_PARSING_HASH
	}

	coreOpts := coreApiAdapter.PayslipDetailsV1Args{
		ID: t.GetID(),
	}
	data, err := handler.coreApi.PayslipDetailsV1(ctx, coreOpts)
	if err != nil {
		if err.Error() == "MotifyCoreAPI: AGENT_NOT_FOUND" {
			return nil, v1Errors.AGENT_NOT_FOUND
		} else if err.Error() == "MotifyCoreAPI: EMPLOYEE_NOT_FOUND" {
			return nil, v1Errors.EMPLOYEE_NOT_FOUND
		} else if err.Error() == "MotifyCoreAPI: PAYSLIP_NOT_FOUND" {
			return nil, v1Errors.PAYSLIP_NOT_FOUND
		}
		return nil, err
	}

	if data.Employee == nil {
		return nil, v1Errors.EMPLOYEE_NOT_FOUND
	}
	userID := uint64(apiToken.GetID())
	if data.Employee.UserFK == nil || *data.Employee.UserFK != userID {
		logger.Error(ctx, "Payslip user (%#v) does not equals current user (%d)", data.Employee.UserFK, userID)
		return nil, v1Errors.PAYSLIP_NOT_FOUND
	}
	if data.Agent == nil {
		return nil, v1Errors.AGENT_NOT_FOUND
	}

	payslipData := PayslipData{}
	if len(data.Payslip.Data) > 0 {
		err := json.Unmarshal([]byte(data.Payslip.Data), &payslipData)
		if err != nil {
			logger.Error(ctx, "Error parsing payslip data: %v", err)
			return nil, v1Errors.ERROR_PARSING_PAYSLIP
		}
	}

	p := data.Payslip
	return &V1Res{
		Payslip: Payslip{
			Hash:        wrapToken.NewPayslip(p.ID, data.Agent.IntegrationFK).Fixed().String(),
			Title:       p.Title,
			Currency:    p.Currency,
			Amount:      p.Amount,
			UpdatedAt:   p.UpdatedAt,
			CreatedAt:   p.CreatedAt,
			Transaction: payslipData.Transaction,
			Sections:    payslipData.Sections,
			Footnote:    payslipData.Footnote,
		},
	}, nil
}
