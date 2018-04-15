package payslip_create

import (
	"motify_core_api/utils/validators"
)

func validateEmail(r RowArgs) (Row, int) {
	if r.Text == nil {
		return Row{}, 1
	}
	errCount := 0
	if !validators.IsValidEmail(*r.Text) {
		errCount++
	}
	return Row{
		Text: *r.Text,
	}, errCount
}

func validatePhone(r RowArgs) (Row, int) {
	if r.Text == nil {
		return Row{}, 1
	}
	errCount := 0
	if !validators.IsValidPhone(*r.Text) {
		errCount++
	}
	return Row{
		Text: *r.Text,
	}, errCount
}

func validateUrl(r RowArgs) (Row, int) {
	if r.Text == nil {
		return Row{}, 1
	}
	errCount := 0
	if !validators.IsValidUrl(*r.Text) {
		errCount++
	}
	return Row{
		Text: *r.Text,
	}, errCount
}

func validateText(r RowArgs) (Row, int) {
	if r.Text == nil {
		return Row{}, 1
	}
	return Row{
		Text: *r.Text,
	}, 0
}

func validateCurrency(r RowArgs) (Row, int) {
	if r.Amount == nil {
		return Row{}, 1
	}
	return Row{
		Amount: r.Amount,
	}, 0
}

func validateInt(r RowArgs) (Row, int) {
	if r.Int == nil {
		return Row{}, 1
	}
	return Row{
		Int: r.Int,
	}, 0
}

func validateFloat(r RowArgs) (Row, int) {
	if r.Float == nil {
		return Row{}, 1
	}
	return Row{
		Float: r.Float,
	}, 0
}

func validateDate(r RowArgs) (Row, int) {
	if r.Text == nil {
		return Row{}, 1
	}
	errCount := 0
	if !validators.IsValidDatetime(*r.Text) {
		errCount++
	}
	return Row{
		Text: *r.Text,
	}, errCount
}

func validateDateRange(r RowArgs) (Row, int) {
	if r.DateFrom == nil || r.DateTo == nil {
		return Row{}, 1
	}
	errCount := 0
	if !validators.IsValidDatetime(*r.DateFrom) {
		errCount++
	}
	if !validators.IsValidDatetime(*r.DateTo) {
		errCount++
	}
	return Row{
		DateFrom: *r.DateFrom,
		DateTo:   *r.DateTo,
	}, errCount
}

func validatePerson(r RowArgs) (Row, int) {
	res := Row{}
	if r.Avatar != nil {
		res.Avatar = *r.Avatar
	}
	if r.Role != nil {
		res.Role = *r.Role
	}
	return res, 0
}

func validateCompany(r RowArgs) (Row, int) {
	res := Row{}
	if r.Avatar != nil {
		res.Avatar = *r.Avatar
	}
	if r.BGImage != nil {
		res.BGImage = *r.BGImage
	}
	return res, 0
}
