package payslip_create

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRowArgToRowWrongType(t *testing.T) {
	arg := RowArgs{
		Type:  "some_type",
		Title: "title",
	}
	r, e := arg.toRow()
	assert.Equal(t, 1, e)
	assert.Equal(t, "Wrong row_type: some_type", r.Status)
	assert.Equal(t, "some_type", r.Type)
	assert.Equal(t, "title", r.Title)
}

func TestRowArgToRowValidatorError(t *testing.T) {
	arg := RowArgs{
		Type:  RowText,
		Title: "title",
	}
	r, e := arg.toRow()
	assert.Equal(t, 1, e)
	assert.Equal(t, "Errors: 1", r.Status)
	assert.Equal(t, RowText, r.Type)
	assert.Equal(t, "title", r.Title)
}

func TestRowArgToRowValidatorOK(t *testing.T) {
	term := "term"
	description := "description"
	footnote := "footnote"
	text := "text"
	arg := RowArgs{
		Type:        RowText,
		Title:       "title",
		Text:        &text,
		Term:        &term,
		Footnote:    &footnote,
		Description: &description,
	}
	r, e := arg.toRow()
	assert.Equal(t, 0, e)
	assert.Equal(t, "", r.Status)
	assert.Equal(t, RowText, r.Type)
	assert.Equal(t, "title", r.Title)
	assert.Equal(t, text, r.Text)
	assert.Equal(t, term, r.Term)
	assert.Equal(t, footnote, r.Footnote)
	assert.Equal(t, description, r.Description)
}

func TestRowArgsListToRowListValidatorError(t *testing.T) {
	list := []RowArgs{
		{
			Type:  "some_type",
			Title: "title",
		},
		{
			Type:  RowText,
			Title: "title",
		},
	}
	rows, e := rowArgsListToRowList(&list)
	assert.Equal(t, 2, e)
	assert.Equal(t, 2, len(rows))
	assert.Equal(t, "Wrong row_type: some_type", rows[0].Status)
	assert.Equal(t, "Errors: 1", rows[1].Status)
}

func TestSectionArgToSectionWrongType(t *testing.T) {
	arg := SectionArgs{
		Type:  "some_type",
		Title: "title",
	}
	r, e := arg.toSection()
	assert.Equal(t, 2, e)
	assert.Equal(t, "No section rows, Wrong section_type: some_type, Errors: 2", r.Status)
	assert.Equal(t, "some_type", r.Type)
	assert.Equal(t, "title", r.Title)
}

func TestSectionArgToSectionWrongTypeAndRows(t *testing.T) {
	arg := SectionArgs{
		Type:  "some_type",
		Title: "title",
		Rows: []RowArgs{
			{
				Type:  "some_type",
				Title: "title",
			},
			{
				Type:  RowText,
				Title: "title",
			},
		},
	}
	r, e := arg.toSection()
	assert.Equal(t, 3, e)
	assert.Equal(t, "Wrong section_type: some_type, Errors: 3", r.Status)
	assert.Equal(t, "some_type", r.Type)
	assert.Equal(t, "title", r.Title)
}

func TestSectionArgToSectionNoRows(t *testing.T) {
	term := "term"
	description := "description"
	arg := SectionArgs{
		Type:       SectionPayslip,
		Title:      "title",
		Term:       &term,
		Definition: &description,
	}
	r, e := arg.toSection()
	assert.Equal(t, 1, e)
	assert.Equal(t, "No section rows, Errors: 1", r.Status)
	assert.Equal(t, SectionPayslip, r.Type)
	assert.Equal(t, "title", r.Title)
	assert.Equal(t, term, r.Term)
	assert.Equal(t, description, r.Definition)
}

func TestTransactionSectionArgToTransactionSectionNoRows(t *testing.T) {
	arg := TransactionSectionArgs{
		Title: "title",
	}
	r, e := arg.toTransactionSection()
	assert.Equal(t, 1, e)
	assert.Equal(t, "No section rows, Errors: 1", r.Status)
	assert.Equal(t, "title", r.Title)
}

func TestTransactionSectionArgToTransactionRows(t *testing.T) {
	arg := TransactionSectionArgs{
		Title: "title",
		Rows: []RowArgs{
			{
				Type:  "some_type",
				Title: "title",
			},
			{
				Type:  RowText,
				Title: "title",
			},
		},
	}
	r, e := arg.toTransactionSection()
	assert.Equal(t, 2, e)
	assert.Equal(t, "Errors: 2", r.Status)
	assert.Equal(t, "title", r.Title)
}
