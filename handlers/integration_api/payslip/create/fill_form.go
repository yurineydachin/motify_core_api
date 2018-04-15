package payslip_create

import (
	"strings"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
)

func (p *PayslipData) getSectionTypes() map[string]bool {
	res := make(map[string]bool, len(p.Sections))
	for i := range p.Sections {
		res[strings.ToLower(p.Sections[i].Type)] = true
	}
	return res
}

func (p *PayslipData) addEmployer(agent *coreApiAdapter.SettingListAgent) {
	if agent == nil {
		return
	}
	sections := p.getSectionTypes()
	if _, exists := sections[SectionCompanySender]; exists {
		return
	}
	agentContactsRows := make([]Row, 0, 4)
	if agent.Phone != "" {
		agentContactsRows = append(agentContactsRows, Row{
			Type:  RowPhone,
			Text:  agent.Phone,
			Title: "Phone",
		})
	}
	if agent.Email != "" {
		agentContactsRows = append(agentContactsRows, Row{
			Type:  RowEmail,
			Text:  agent.Email,
			Title: "Email",
		})
	}
	if agent.Address != "" {
		agentContactsRows = append(agentContactsRows, Row{
			Type:  RowText,
			Text:  agent.Address,
			Title: "Address",
		})
	}
	if agent.Site != "" {
		agentContactsRows = append(agentContactsRows, Row{
			Type:  RowUrl,
			Text:  agent.Site,
			Title: "Site",
		})
	}
	p.Sections = append(p.Sections, Section{
		Type:  SectionCompanySender,
		Title: "EMPLOYER",
		Rows: []Row{
			{
				Type:        RowCompany,
				Title:       agent.Name,
				Avatar:      agent.Logo,
				BGImage:     agent.Background,
				Description: agent.Description,
				Children:    agentContactsRows,
			},
		},
	})
}

func (p *PayslipData) addEmployee(emp *coreApiAdapter.EmployeeDetailsEmployee) {
	if emp == nil {
		return
	}
	sections := p.getSectionTypes()
	if _, exists := sections[SectionPersonReceiver]; exists {
		return
	}
	numberOfDep := int64(emp.NumberOfDepandants)
	empDetailsRows := []Row{
		{
			Type:  RowText,
			Text:  emp.Code,
			Title: "Employee code",
		},
		{
			Type:   RowCurrency,
			Amount: &emp.GrossBaseSalary,
			Title:  "Gross Base Salary",
		},
		{
			Type:  RowInt,
			Int:   &numberOfDep,
			Title: "Number of depandants",
		},
	}
	if emp.HireDate != "" {
		empDetailsRows = append(empDetailsRows, Row{
			Type:  RowText,
			Text:  emp.HireDate,
			Title: "HireDate",
		})
	}
	p.Sections = append(p.Sections, Section{
		Type:  SectionPersonReceiver,
		Title: "EMPLOYEE",
		Rows: []Row{
			{
				Type:  RowPerson,
				Title: emp.Name,
				//Avatar: user.Awatar,
				//Footnote: "* Employee details are relevant to this payslip only",
				Children: empDetailsRows,
			},
		},
	})
}

func (p *PayslipData) addPersonPreparedBy(user *coreApiAdapter.UserUpdateUser) {
	if user == nil {
		return
	}
	sections := p.getSectionTypes()
	if _, exists := sections[SectionPersonProcesser]; exists {
		return
	}
	userDetailsRows := make([]Row, 0, 2)
	if user.Phone != "" {
		userDetailsRows = append(userDetailsRows, Row{
			Type:  RowPhone,
			Text:  user.Phone,
			Title: "Phone",
		})
	}
	if user.Email != "" {
		userDetailsRows = append(userDetailsRows, Row{
			Type:  RowEmail,
			Text:  user.Email,
			Title: "Email",
		})
	}
	p.Sections = append(p.Sections, Section{
		Type:  SectionPersonProcesser,
		Title: "PREPARED BY",
		Rows: []Row{
			{
				Type:        RowPerson,
				Title:       user.Name,
				Avatar:      user.Awatar,
				Role:        user.Short,
				Description: user.Description,
				//Footnote: "* Employee details are relevant to this payslip only",
				Children: userDetailsRows,
			},
		},
	})
}

func (p *PayslipData) addCompanyProceccedBy(agent *coreApiAdapter.SettingListAgent) {
	if agent == nil {
		return
	}
	sections := p.getSectionTypes()
	if _, exists := sections[SectionCompanyProcesser]; exists {
		return
	}
	agentContactsRows := make([]Row, 0, 4)
	if agent.Phone != "" {
		agentContactsRows = append(agentContactsRows, Row{
			Type:  RowPhone,
			Text:  agent.Phone,
			Title: "Phone",
		})
	}
	if agent.Email != "" {
		agentContactsRows = append(agentContactsRows, Row{
			Type:  RowEmail,
			Text:  agent.Email,
			Title: "Email",
		})
	}
	if agent.Address != "" {
		agentContactsRows = append(agentContactsRows, Row{
			Type:  RowText,
			Text:  agent.Address,
			Title: "Address",
		})
	}
	if agent.Site != "" {
		agentContactsRows = append(agentContactsRows, Row{
			Type:  RowUrl,
			Text:  agent.Site,
			Title: "Site",
		})
	}
	p.Sections = append(p.Sections, Section{
		Type:  SectionCompanyProcesser,
		Title: "PROCESSED BY",
		Rows: []Row{
			{
				Type:        RowCompany,
				Title:       agent.Name,
				Avatar:      agent.Logo,
				BGImage:     agent.Background,
				Description: agent.Description,
				Children:    agentContactsRows,
			},
		},
	})
}
