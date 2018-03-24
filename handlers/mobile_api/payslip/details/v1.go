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
	Agent    *Agent    `json:"agent"`
	Employee *Employee `json:"employee"`
	Payslip  Payslip   `json:"payslip"`
}

type Agent struct {
	ID          uint64 `json:"id_agent"`
	Name        string `json:"name"`
	CompanyID   string `json:"company_id"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	Site        string `json:"site"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}

type Employee struct {
	ID                 uint64  `json:"id_employee"`
	AgentFK            uint64  `json:"fk_agent"`
	UserFK             *uint64 `json:"fk_user"`
	Code               string  `json:"employee_code"`
	Name               string  `json:"name"`
	Role               string  `json:"role"`
	Email              string  `json:"email"`
	HireDate           string  `json:"hire_date"`
	NumberOfDepandants uint    `json:"number_of_dependants"`
	GrossBaseSalary    float64 `json:"gross_base_salary"`
	UpdatedAt          string  `json:"updated_at"`
	CreatedAt          string  `json:"created_at"`
}

type Payslip struct {
	ID         uint64      `json:"id_payslip"`
	EmployeeFK uint64      `json:"fk_employee"`
	Title      string      `json:"title"`
	Currency   string      `json:"currency"`
	Amount     float64     `json:"amount"`
	UpdatedAt  string      `json:"updated_at"`
	CreatedAt  string      `json:"created_at"`
	Data       PayslipData `json:"data"`
}

type PayslipData struct {
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
	Person     *Person      `json:"person,omitempty"`
	Company    *Company     `json:"company,omitempty"`
}

type Operation struct {
	Title      string       `json:"title,omitempty"`
	Term       string       `json:"term,omitempty"`
	Definition string       `json:"definition,omitempty"`
	Type       string       `json:"type,omitempty"`
	Footnote   string       `json:"footnote,omitempty"`
	Amount     *float64     `json:"amount,omitempty"`
	Float      *float64     `json:"float,omitempty"`
	Int        *int64       `json:"int,omitempty"`
	Text       string       `json:"text,omitempty"`
	Children   *[]Operation `json:"rows,omitempty"`
}

type Person struct {
	Name        string       `json:"name"`
	Avatar      string       `json:"avatar_image"`
	Role        string       `json:"job_title"`
	Description string       `json:"description"`
	Details     *[]Operation `json:"details,omitempty"`
	Contacts    []Contact    `json:"contacts"`
	Footnote    string       `json:"footnote,omitempty"`
}

type Company struct {
	Title       string    `json:"title"`
	Name        string    `json:"official_name"`
	Logo        string    `json:"logo_image,omitempty"`
	BGImage     string    `json:"bg_image,omitempty"`
	Description string    `json:"description,omitempty"`
	Contacts    []Contact `json:"contacts"`
	Footnote    string    `json:"footnote,omitempty"`
}

type Contact struct {
	Title string `json:"title"`
	Type  string `json:"type"`
	Text  string `json:"text"`
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
	userID := uint64(apiToken.GetCustomerID())
	if data.Employee.UserFK == nil || *data.Employee.UserFK != userID {
		logger.Error(ctx, "Payslip user (%#v) does not equals current user (%d)", data.Employee.UserFK, userID)
		return nil, v1Errors.PAYSLIP_NOT_FOUND
	}
	if data.Agent == nil {
		return nil, v1Errors.AGENT_NOT_FOUND
	}

	agent := data.Agent
	employee := data.Employee
	p := data.Payslip
	return &V1Res{
		Agent: &Agent{
			ID:          agent.ID,
			Name:        agent.Name,
			CompanyID:   agent.CompanyID,
			Description: agent.Description,
			Logo:        agent.Logo,
			Phone:       agent.Phone,
			Email:       agent.Email,
			Address:     agent.Address,
			Site:        agent.Site,
		},
		Employee: &Employee{
			ID:                 employee.ID,
			AgentFK:            employee.AgentFK,
			UserFK:             employee.UserFK,
			Code:               employee.Code,
			Name:               employee.Name,
			Role:               employee.Role,
			Email:              employee.Email,
			HireDate:           employee.HireDate,
			NumberOfDepandants: employee.NumberOfDepandants,
			GrossBaseSalary:    employee.GrossBaseSalary,
		},
		Payslip: Payslip{
			ID:         p.ID,
			EmployeeFK: p.EmployeeFK,
			Title:      p.Title,
			Currency:   p.Currency,
			Amount:     p.Amount,
			UpdatedAt:  p.UpdatedAt,
			CreatedAt:  p.CreatedAt,
			Data: PayslipData{
				Transaction: Transaction{
					Description: p.Data.Transaction.Description,
					Sections:    convertSections(p.Data.Transaction.Sections),
				},
				Sections: convertSections(p.Data.Sections),
				Footnote: p.Data.Footnote,
			},
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
			Person:     convertPerson(s.Person),
			Company:    convertCompany(s.Company),
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
			Title:      o.Title,
			Term:       o.Term,
			Definition: o.Definition,
			Type:       o.Type,
			Footnote:   o.Footnote,
			Amount:     o.Amount,
			Float:      o.Float,
			Int:        o.Int,
			Text:       o.Text,
			Children:   convertOperations(o.Children),
		})
	}
	return &res
}

func convertPerson(p *coreApiAdapter.PayslipDetailsPerson) *Person {
	if p == nil {
		return nil
	}
	return &Person{
		Name:        p.Name,
		Avatar:      p.Avatar,
		Role:        p.Role,
		Description: p.Description,
		Details:     convertOperations(p.Details),
		Contacts:    convertContacts(p.Contacts),
		Footnote:    p.Footnote,
	}
}

func convertCompany(c *coreApiAdapter.PayslipDetailsCompany) *Company {
	if c == nil {
		return nil
	}
	return &Company{
		Title:       c.Title,
		Name:        c.Name,
		Logo:        c.Logo,
		BGImage:     c.BGImage,
		Description: c.Description,
		Contacts:    convertContacts(c.Contacts),
		Footnote:    c.Footnote,
	}
}

func convertContacts(list []coreApiAdapter.PayslipDetailsContact) []Contact {
	res := make([]Contact, 0, len(list))
	for i := range list {
		c := list[i]
		res = append(res, Contact{
			Title: c.Title,
			Type:  c.Type,
			Text:  c.Text,
		})
	}
	return res
}
