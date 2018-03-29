package payslip_details

import (
	"context"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/mobapi_lib/token"
	"godep.lzd.co/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
)

type V1Args struct {
	ID uint64 `key:"payslip_id" description:"Payslip id"`
}

type V1Res struct {
	Payslip Payslip `json:"payslip"`
}

type Payslip struct {
	ID          uint64      `json:"id_payslip"`
	Title       string      `json:"title"`
	Currency    string      `json:"currency"`
	Amount      float64     `json:"amount"`
	UpdatedAt   string      `json:"updated_at"`
	CreatedAt   string      `json:"created_at"`
	Transaction Transaction `json:"transaction"`
	Sections    []Section   `json:"sections"`
	Footnote    string      `json:"footnote,omitempty"`
}

type Transaction struct {
	Description string    `json:"description"`
	Sections    []Section `json:"sections"`
}

type Section struct {
	Type       string       `json:"section_type,omitempty"`
	Title      string       `json:"title,omitempty"`
	Term       string       `json:"term,omitempty"`
	Definition string       `json:"definition,omitempty"`
	Amount     *float64     `json:"amount,omitempty"`
	Operations *[]Operation `json:"rows,omitempty"`
}

type Operation struct {
	Title       string       `json:"title,omitempty"`
	Term        string       `json:"term,omitempty"`
	Type        string       `json:"type,omitempty"`
	Description string       `json:"description,omitempty"`
	Footnote    string       `json:"footnote,omitempty"`
	Amount      *float64     `json:"amount,omitempty"`
	Float       *float64     `json:"float,omitempty"`
	Int         *int64       `json:"int,omitempty"`
	Text        string       `json:"text,omitempty"`
	Children    *[]Operation `json:"rows,omitempty"`

	Name    string `json:"name,omitempty"`
	Role    string `json:"job_title,omitempty"`
	Avatar  string `json:"avatar_image,omitempty"`
	BGImage string `json:"bg_image,omitempty"`
}

type V1ErrorTypes struct {
	AGENT_NOT_FOUND       error `text:"agent not found"`
	EMPLOYEE_NOT_FOUND    error `text:"employee not found"`
	PAYSLIP_NOT_FOUND     error `text:"payslip not found"`
	ERROR_PARSING_PAYSLIP error `text:"error parsing payslip"`
}

var v1Errors V1ErrorTypes

func (*Handler) V1ErrorsVar() *V1ErrorTypes {
	return &v1Errors
}

func (handler *Handler) V1(ctx context.Context, opts *V1Args, apiToken token.IToken) (*V1Res, error) {
	logger.Debug(ctx, "Payslip/List/V1")
	cache.DisableTransportCache(ctx)

	coreOpts := coreApiAdapter.PayslipDetailsV1Args{
		ID: opts.ID,
	}
	data, err := handler.coreApi.PayslipDetailsV1(ctx, coreOpts)
	if err != nil {
		if err.Error() == "MotifyCoreAPI: AGENT_NOT_FOUND" {
			return nil, v1Errors.AGENT_NOT_FOUND
		} else if err.Error() == "MotifyCoreAPI: EMPLOYEE_NOT_FOUND" {
			return nil, v1Errors.EMPLOYEE_NOT_FOUND
		} else if err.Error() == "MotifyCoreAPI: PAYSLIP_NOT_FOUND" {
			return nil, v1Errors.PAYSLIP_NOT_FOUND
		} else if err.Error() == "MotifyCoreAPI: ERROR_PARSING_PAYSLIP" {
			return nil, v1Errors.ERROR_PARSING_PAYSLIP
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

	p := data.Payslip
	return &V1Res{
		Payslip: Payslip{
			ID:        p.ID,
			Title:     p.Title,
			Currency:  p.Currency,
			Amount:    p.Amount,
			UpdatedAt: p.UpdatedAt,
			CreatedAt: p.CreatedAt,
			Transaction: Transaction{
				Description: p.Data.Transaction.Description,
				Sections:    convertSections(p.Data.Transaction.Sections),
			},
			Sections: convertSections(p.Data.Sections),
			Footnote: p.Data.Footnote,
		},
	}, nil
}

func convertSections(list []coreApiAdapter.PayslipDetailsSection) []Section {
	res := make([]Section, 0, len(list))
	for i := range list {
		s := list[i]
		res = append(res, Section{
			Type:       s.Type,
			Title:      s.Title,
			Term:       s.Term,
			Definition: s.Definition,
			Amount:     s.Amount,
			Operations: convertOperations(s.Operations),
		})
	}
	return res
}

func convertOperations(list *[]coreApiAdapter.PayslipDetailsOperation) *[]Operation {
	if list == nil || len(*list) == 0 {
		return nil
	}
	res := make([]Operation, 0, len(*list))
	for i := range *list {
		o := (*list)[i]
		res = append(res, Operation{
			Title:       o.Title,
			Term:        o.Term,
			Description: o.Description,
			Type:        o.Type,
			Footnote:    o.Footnote,
			Amount:      o.Amount,
			Float:       o.Float,
			Int:         o.Int,
			Text:        o.Text,
			Children:    convertOperations(o.Children),

			Name:    o.Name,
			Avatar:  o.Avatar,
			Role:    o.Role,
			BGImage: o.BGImage,
		})
	}
	return &res
}