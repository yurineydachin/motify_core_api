package payslip_create

import (
	"context"
	"encoding/json"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/mobapi_lib/token"
	"godep.lzd.co/service/logger"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
	wrapToken "motify_core_api/utils/token"
)

const (
	SectionPersonReceiver   = "person_receiver"
	SectionPersonProcesser  = "person_processer"
	SectionCompanySender    = "company_sender"
	SectionCompanyProcesser = "company_processer"

	RowEmail     = "email"
	RowPhone     = "phone"
	RowUrl       = "url"
	RowText      = "text"
	RowCurrency  = "currency"
	RowInt       = "int"
	RowFloat     = "float"
	RowDate      = "date"
	RowDateRange = "daterange"
	RowPerson    = "person"
	RowCompany   = "company"
)

var rowTypes = map[string]func(RowArgs) (Row, uint64){
	RowEmail:     validateEmail,
	RowPhone:     validatePhone,
	RowUrl:       validateUrl,
	RowText:      validateText,
	RowCurrency:  validateCurrency,
	RowInt:       validateInt,
	RowFloat:     validateFloat,
	RowDate:      validateDate,
	RowDateRange: validateDateRange,
	RowPerson:    validatePerson,
	RowCompany:   validateCompany,
}

var sectionTypes = map[string]bool{
	SectionPersonReceiver:   true,
	SectionPersonProcesser:  true,
	SectionCompanySender:    true,
	SectionCompanyProcesser: true,
	"details":               true,
	"payslip":               true,
}

type V1Args struct {
	CompanyID string       `key:"company_id" description:"Company id"`
	Code      string       `key:"employee_code" description:"employee code"`
	Payslip   *PayslipArgs `key:"payslip" description:"payslip"`
}

type PayslipArgs struct {
	Title       string          `key:"title" description:"Title"`
	Currency    string          `key:"currency" description:"Currency"`
	Amount      float64         `key:"amount" description:"Amount"`
	Transaction TransactionArgs `key:"transaction" description:"transaction"`
	Sections    []SectionArgs   `key:"sections" description:"sections"`
	Footnote    *string         `key:"footnote" description:"footnote"`
}

type TransactionArgs struct {
	Description string                   `key:"description" description:"description"`
	Sections    []TransactionSectionArgs `key:"sections" description:"sections"`
}

type TransactionSectionArgs struct {
	Title string    `key:"title" description:"title"`
	Rows  []RowArgs `key:"rows" description:"rows"`
}

type SectionArgs struct {
	Type       string    `key:"section_type" description:"type"`
	Title      string    `key:"title" description:"title"`
	Term       *string   `key:"term" description:"term"`
	Definition *string   `key:"definition" description:"definition"`
	Amount     *float64  `key:"amount" description:"amount"`
	Rows       []RowArgs `key:"rows" description:"rows"`
}

type RowArgs struct {
	Type        string     `key:"row_type" description:"type"`
	Title       string     `key:"title" description:"title"`
	Term        *string    `key:"term" description:"term"`
	Description *string    `key:"description" description:"description"`
	Footnote    *string    `key:"footnote" description:"footnote"`
	Role        *string    `key:"role" description:"role"`
	Avatar      *string    `key:"avatar_image" description:"avatar"`
	BGImage     *string    `key:"bg_image" description:"bg_image"`
	Amount      *float64   `key:"amount" description:"amount"`
	Float       *float64   `key:"float" description:"float"`
	Int         *int64     `key:"int" description:"int"`
	Text        *string    `key:"text" description:"text"`
	DateFrom    *string    `key:"date_from" description:"date from"`
	DateTo      *string    `key:"date_to" description:"date to"`
	Children    *[]RowArgs `key:"rows" description:"rows"`
}

type V1Res struct {
	Status  string  `json:"status"`
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
	Footnote    string      `json:"footnote"`
}

type PayslipData struct {
	Status      string      `json:"status,omitempty"`
	Transaction Transaction `json:"transaction"`
	Sections    []Section   `json:"sections"`
	Footnote    string      `json:"footnote,omitempty"`
}

type Transaction struct {
	Status      string               `json:"status,omitempty"`
	Description string               `json:"description"`
	Sections    []TransactionSection `json:"sections"`
}

type TransactionSection struct {
	Status string `json:"status,omitempty"`
	Title  string `json:"title"`
	Rows   []Row  `json:"rows"`
}

type Section struct {
	Status     string   `json:"status,omitempty"`
	Type       string   `json:"section_type"`
	Title      string   `json:"title"`
	Term       string   `json:"term,omitempty"`
	Definition string   `json:"definition,omitempty"`
	Amount     *float64 `json:"amount,omitempty"`
	Rows       []Row    `json:"rows,omitempty"`
}

type Row struct {
	Status      string   `json:"status,omitempty"`
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
	CREATE_FAILED         error `text:"creating payslip is failed"`
	USER_AGENT_NOT_FOUND  error `text:"user agent not found"`
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

	userID := apiToken.GetID()
	integrationID := apiToken.GetExtraID()
	coreSettingListOpts := coreApiAdapter.SettingListV1Args{
		UserID:        userID,
		IntegrationID: integrationID,
	}
	listData, err := handler.coreApi.SettingListV1(ctx, coreSettingListOpts)
	if err != nil {
		return nil, err
	}

	agent, setting := findByCompanyID(listData.List, opts.CompanyID)
	if agent == nil {
		return nil, v1Errors.AGENT_NOT_FOUND
	}

	coreEmployeeDetailsOpts := coreApiAdapter.EmployeeDetailsV1Args{
		AgentFK: &agent.ID,
		Code:    &opts.Code,
	}

	dataEmp, err := handler.coreApi.EmployeeDetailsV1(ctx, coreEmployeeDetailsOpts)
	if err != nil {
		return nil, err
	}
	if dataEmp == nil || dataEmp.Employee == nil {
		return nil, v1Errors.EMPLOYEE_NOT_FOUND
	}

	var agentProcessed *coreApiAdapter.SettingListAgent
	if setting.AgentProcessedFK != nil && *setting.AgentProcessedFK > 0 {
		agentProcessed, _ = findAgentByID(listData.List, *setting.AgentProcessedFK)
	}

	userCoreOpts := coreApiAdapter.UserUpdateV1Args{
		ID: uint64(apiToken.GetID()),
	}
	userData, err := handler.coreApi.UserUpdateV1(ctx, userCoreOpts)
	if err != nil || userData.User == nil {
		return nil, v1Errors.USER_AGENT_NOT_FOUND
	}

	payslipData, errCount := opts.Payslip.toPayslipData()

	if errCount > 0 {
		return &V1Res{
			Status: payslipData.Status,
			Payslip: Payslip{
				Title:       opts.Payslip.Title,
				Currency:    opts.Payslip.Currency,
				Amount:      opts.Payslip.Amount,
				Transaction: payslipData.Transaction,
				Sections:    payslipData.Sections,
				Footnote:    payslipData.Footnote,
			},
		}, nil
	}

	payslipData.addEmployer(agent)
	payslipData.addEmployee(dataEmp.Employee)
	payslipData.addPersonPreparedBy(userData.User)
	payslipData.addCompanyProceccedBy(agentProcessed)

	payslipDataText, err := json.Marshal(payslipData)
	if err != nil {
		return nil, v1Errors.ERROR_PARSING_PAYSLIP
	}

	corePayslipCreateOpts := coreApiAdapter.PayslipCreateV1Args{
		Payslip: coreApiAdapter.PayslipCreatePayslipArgs{
			EmployeeFK: dataEmp.Employee.ID,
			Title:      opts.Payslip.Title,
			Currency:   opts.Payslip.Currency,
			Amount:     opts.Payslip.Amount,
			Data:       string(payslipDataText),
		},
	}

	data, err := handler.coreApi.PayslipCreateV1(ctx, corePayslipCreateOpts)
	if err != nil {
		if err.Error() == "MotifyCoreAPI: CREATE_FAILED" {
			return nil, v1Errors.CREATE_FAILED
		} else if err.Error() == "MotifyCoreAPI: EMPLOYEE_NOT_FOUND" {
			return nil, v1Errors.EMPLOYEE_NOT_FOUND
		} else if err.Error() == "MotifyCoreAPI: PAYSLIP_NOT_CREATED" {
			return nil, v1Errors.CREATE_FAILED
		}
		return nil, err
	}
	if data.Employee == nil {
		return nil, v1Errors.EMPLOYEE_NOT_FOUND
	}
	if data.Payslip == nil {
		return nil, v1Errors.CREATE_FAILED
	}

	p := data.Payslip
	return &V1Res{
		Payslip: Payslip{
			Hash:        wrapToken.NewPayslip(p.ID, agent.IntegrationFK).Fixed().String(),
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

func findByCompanyID(list []coreApiAdapter.SettingListListItem, companyID string) (*coreApiAdapter.SettingListAgent, *coreApiAdapter.SettingListAgentSetting) {
	for i := range list {
		if list[i].Agent != nil && list[i].Agent.CompanyID == companyID {
			return list[i].Agent, list[i].Setting
		}
	}
	return nil, nil
}

func findAgentByID(list []coreApiAdapter.SettingListListItem, agentFK uint64) (*coreApiAdapter.SettingListAgent, *coreApiAdapter.SettingListAgentSetting) {
	for i := range list {
		if list[i].Agent != nil && list[i].Agent.ID == agentFK {
			return list[i].Agent, list[i].Setting
		}
	}
	return nil, nil
}
