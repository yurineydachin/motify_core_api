package payslip_create

import (
	"fmt"
	"strings"
)

func getErrorCountMessage(cnt int) string {
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
		Transaction: t,
		Sections:    sections,
	}
	if p.Footnote != nil {
		res.Footnote = *p.Footnote
	}
	if errTotalCount > 0 {
		res.Status = getErrorCountMessage(errTotalCount)
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
		Description: t.Description,
		Sections:    sections,
	}
	if errTotalCount > 0 {
		res.Status = getErrorCountMessage(errTotalCount)
	}
	return res, errTotalCount
}

func (s TransactionSectionArgs) toTransactionSection() (TransactionSection, int) {
	rows := make([]Row, 0, len(s.Rows))
	errTotalCount := 0
	for i := range s.Rows {
		r, errCount := s.Rows[i].toRow()
		errTotalCount += errCount
		rows = append(rows, r)
	}
	res := TransactionSection{
		Title: s.Title,
		Rows:  rows,
	}
	if errTotalCount > 0 {
		res.Status = getErrorCountMessage(errTotalCount)
	}
	return res, errTotalCount
}

func (s SectionArgs) toSection() (Section, int) {
	rows := make([]Row, 0, len(s.Rows))
	errTotalCount := 0
	for i := range s.Rows {
		r, errCount := s.Rows[i].toRow()
		errTotalCount += errCount
		rows = append(rows, r)
	}
	res := Section{
		Title:  s.Title,
		Type:   strings.ToLower(s.Type),
		Rows:   rows,
		Amount: s.Amount,
	}
	if errTotalCount > 0 {
		res.Status = getErrorCountMessage(errTotalCount)
	}
	if t, ok := sectionTypes[res.Type]; !ok || !t {
		message := fmt.Sprintf("Wrong section_type: %s", res.Type)
		if res.Status != "" {
			res.Status += ", " + message
		} else {
			res.Status = message
		}
	}
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
	if errTotalCount > 0 {
		res.Status = getErrorCountMessage(errTotalCount)
	}
	return res, errTotalCount
}
