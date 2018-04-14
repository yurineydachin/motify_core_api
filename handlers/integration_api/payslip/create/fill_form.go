package payslip_create

import (
	"strings"

	coreApiAdapter "motify_core_api/resources/motify_core_api"
)

func (p *PayslipArgs) getSectionTypes() map[string]bool {
	res := make(map[string]bool, len(p.Sections))
	for i := range p.Sections {
		res[strings.ToLower(p.Sections[i].Type)] = true
	}
	return res
}

func (p *PayslipArgs) addEmployer(agent *coreApiAdapter.SettingListAgent) {
	if agent == nil {
		return
	}
	sections := p.getSectionTypes()
	if _, exists := sections[SectionCompanySender]; exists {
		return
	}
	agentContactsRows := make([]RowArgs, 0, 4)
	if agent.Phone != "" {
		agentContactsRows = append(agentContactsRows, RowArgs{
			Type:  RowPhone,
			Text:  &agent.Phone,
			Title: "Phone",
		})
	}
	if agent.Email != "" {
		agentContactsRows = append(agentContactsRows, RowArgs{
			Type:  RowEmail,
			Text:  &agent.Email,
			Title: "Email",
		})
	}
	if agent.Address != "" {
		agentContactsRows = append(agentContactsRows, RowArgs{
			Type:  RowText,
			Text:  &agent.Address,
			Title: "Address",
		})
	}
	if agent.Site != "" {
		agentContactsRows = append(agentContactsRows, RowArgs{
			Type:  RowUrl,
			Text:  &agent.Site,
			Title: "Site",
		})
	}
	p.Sections = append(p.Sections, SectionArgs{
		Type:  SectionCompanySender,
		Title: "EMPLOYER",
		Rows: []RowArgs{
			{
				Type:        RowCompany,
				Title:       agent.Name,
				Avatar:      &agent.Logo,
				BGImage:     &agent.Background,
				Description: &agent.Description,
				Children:    &agentContactsRows,
			},
		},
	})
}

func (p *PayslipArgs) addEmployee(emp *coreApiAdapter.EmployeeDetailsEmployee) {
	if emp == nil {
		return
	}
	sections := p.getSectionTypes()
	if _, exists := sections[SectionPersonReceiver]; exists {
		return
	}
	numberOfDep := int64(emp.NumberOfDepandants)
	empDetailsRows := []RowArgs{
		{
			Type:  RowText,
			Text:  &emp.Code,
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
		empDetailsRows = append(empDetailsRows, RowArgs{
			Type:  RowText,
			Text:  &emp.HireDate,
			Title: "HireDate",
		})
	}
	p.Sections = append(p.Sections, SectionArgs{
		Type:  SectionPersonReceiver,
		Title: "EMPLOYEE",
		Rows: []RowArgs{
			{
				Type:  RowPerson,
				Title: emp.Name,
				//Avatar: user.Awatar,
				//Footnote: "* Employee details are relevant to this payslip only",
				Children: &empDetailsRows,
			},
		},
	})
}

func (p *PayslipArgs) addPersonPreparedBy(user *coreApiAdapter.UserUpdateUser) {
	if user == nil {
		return
	}
	sections := p.getSectionTypes()
	if _, exists := sections[SectionPersonProcesser]; exists {
		return
	}
	userDetailsRows := make([]RowArgs, 0, 2)
	if user.Phone != "" {
		userDetailsRows = append(userDetailsRows, RowArgs{
			Type:  RowPhone,
			Text:  &user.Phone,
			Title: "Phone",
		})
	}
	if user.Email != "" {
		userDetailsRows = append(userDetailsRows, RowArgs{
			Type:  RowEmail,
			Text:  &user.Email,
			Title: "Email",
		})
	}
	p.Sections = append(p.Sections, SectionArgs{
		Type:  SectionPersonProcesser,
		Title: "PREPARED BY",
		Rows: []RowArgs{
			{
				Type:        RowPerson,
				Title:       user.Name,
				Avatar:      &user.Awatar,
				Role:        &user.Short,
				Description: &user.Description,
				//Footnote: "* Employee details are relevant to this payslip only",
				Children: &userDetailsRows,
			},
		},
	})
}

func (p *PayslipArgs) addCompanyProceccedBy(agent *coreApiAdapter.SettingListAgent) {
	if agent == nil {
		return
	}
	sections := p.getSectionTypes()
	if _, exists := sections[SectionCompanyProcesser]; exists {
		return
	}
	agentContactsRows := make([]RowArgs, 0, 4)
	if agent.Phone != "" {
		agentContactsRows = append(agentContactsRows, RowArgs{
			Type:  RowPhone,
			Text:  &agent.Phone,
			Title: "Phone",
		})
	}
	if agent.Email != "" {
		agentContactsRows = append(agentContactsRows, RowArgs{
			Type:  RowEmail,
			Text:  &agent.Email,
			Title: "Email",
		})
	}
	if agent.Address != "" {
		agentContactsRows = append(agentContactsRows, RowArgs{
			Type:  RowText,
			Text:  &agent.Address,
			Title: "Address",
		})
	}
	if agent.Site != "" {
		agentContactsRows = append(agentContactsRows, RowArgs{
			Type:  RowUrl,
			Text:  &agent.Site,
			Title: "Site",
		})
	}
	p.Sections = append(p.Sections, SectionArgs{
		Type:  SectionCompanyProcesser,
		Title: "PROCESSED BY",
		Rows: []RowArgs{
			{
				Type:        RowCompany,
				Title:       agent.Name,
				Avatar:      &agent.Logo,
				BGImage:     &agent.Background,
				Description: &agent.Description,
				Children:    &agentContactsRows,
			},
		},
	})
}
