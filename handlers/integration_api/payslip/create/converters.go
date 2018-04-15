package payslip_create

import (
	"fmt"
	"strings"
)

func getErrorCountMessage(cnt int) string {
	if cnt <= 0 {
		return ""
	}
	return fmt.Sprintf("Errors: %d", cnt)
}

func (p *PayslipArgs) toPayslipData() (*PayslipData, int) {
	t, errTotalCount := p.Transaction.toTransaction()
	sections := make([]Section, 0, len(p.Sections))
	for i := range p.Sections {
		s, errCount := p.Sections[i].toSection()
		errTotalCount += errCount
		sections = append(sections, s)
	}
	res := &PayslipData{
		Status:      getErrorCountMessage(errTotalCount),
		Transaction: t,
		Sections:    sections,
	}
	if p.Footnote != nil {
		res.Footnote = *p.Footnote
	}
	return res, errTotalCount
}

func (t TransactionArgs) toTransaction() (Transaction, int) {
	sections := make([]TransactionSection, 0, len(t.Sections))
	errTotalCount := 0
	for i := range t.Sections {
		s, errCount := t.Sections[i].toTransactionSection()
		errTotalCount += errCount
		sections = append(sections, s)
	}
	res := Transaction{
		Status:      getErrorCountMessage(errTotalCount),
		Description: t.Description,
		Sections:    sections,
	}
	return res, errTotalCount
}

func (s TransactionSectionArgs) toTransactionSection() (TransactionSection, int) {
	rows, errTotalCount := rowArgsListToRowList(&s.Rows)
	errorMessages := []string{}
	if len(rows) == 0 {
		errorMessages = append(errorMessages, "No section rows")
		errTotalCount++
	}
	if errTotalCount > 0 {
		errorMessages = append(errorMessages, getErrorCountMessage(errTotalCount))
	}
	return TransactionSection{
		Status: strings.Join(errorMessages, ", "),
		Title:  s.Title,
		Rows:   rows,
	}, errTotalCount
}

func (s SectionArgs) toSection() (Section, int) {
	rows, errTotalCount := rowArgsListToRowList(&s.Rows)
	res := Section{
		Type:   strings.ToLower(s.Type),
		Title:  s.Title,
		Amount: s.Amount,
		Rows:   rows,
	}
	errorMessages := []string{}
	if len(rows) == 0 {
		errorMessages = append(errorMessages, "No section rows")
		errTotalCount++
	}
	if t, ok := sectionTypes[res.Type]; !ok || !t {
		errTotalCount++
		errorMessages = append(errorMessages, fmt.Sprintf("Wrong section_type: %s", res.Type))
	}
	if errTotalCount > 0 {
		errorMessages = append(errorMessages, getErrorCountMessage(errTotalCount))
	}
	res.Status = strings.Join(errorMessages, ", ")
	if s.Term != nil && *s.Term != "" {
		res.Term = *s.Term
	}
	if s.Definition != nil && *s.Definition != "" {
		res.Definition = *s.Definition
	}
	return res, errTotalCount
}

func rowArgsListToRowList(list *[]RowArgs) ([]Row, int) {
	if list == nil || len(*list) == 0 {
		return nil, 0
	}

	errTotalCount := 0
	rows := make([]Row, 0, len(*list))
	for i := range *list {
		r, errCount := (*list)[i].toRow()
		errTotalCount += errCount
		rows = append(rows, r)
	}
	return rows, errTotalCount
}

func (r RowArgs) toRow() (Row, int) {
	t := strings.ToLower(r.Type)
	validator, ok := rowTypes[t]
	if !ok || validator == nil {
		return Row{
			Title:  r.Title,
			Type:   t,
			Status: fmt.Sprintf("Wrong row_type: %s", t),
		}, 1
	}
	res, errTotalCount := validator(r)
	rowChildren, errCount := rowArgsListToRowList(r.Children)
	errTotalCount += errCount

	res.Status = getErrorCountMessage(errTotalCount)
	res.Title = r.Title
	res.Type = t
	res.Children = rowChildren
	if r.Term != nil {
		res.Term = *r.Term
	}
	if r.Description != nil {
		res.Description = *r.Description
	}
	if r.Footnote != nil {
		res.Footnote = *r.Footnote
	}
	return res, errTotalCount
}
