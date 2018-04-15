package payslip_create

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateEmailFail(t *testing.T) {
	r, e := validateEmail(RowArgs{})
	assert.Equal(t, 1, e)
	assert.Equal(t, "", r.Text)
}

func TestValidateEmailWrong(t *testing.T) {
	text := "blabla"
	r, e := validateEmail(RowArgs{
		Text: &text,
	})
	assert.Equal(t, 1, e)
	assert.Equal(t, text, r.Text)
}

func TestValidateEmailOk(t *testing.T) {
	text := "blabla@bla.com"
	r, e := validateEmail(RowArgs{
		Text: &text,
	})
	assert.Equal(t, 0, e)
	assert.Equal(t, text, r.Text)
}

func TestValidateUrlFail(t *testing.T) {
	r, e := validateUrl(RowArgs{})
	assert.Equal(t, 1, e)
	assert.Equal(t, "", r.Text)
}

func TestValidateUrlWrong(t *testing.T) {
	text := "blabla"
	r, e := validateUrl(RowArgs{
		Text: &text,
	})
	assert.Equal(t, 1, e)
	assert.Equal(t, text, r.Text)
}

func TestValidateUrlOk(t *testing.T) {
	text := "http://blabla.com"
	r, e := validateUrl(RowArgs{
		Text: &text,
	})
	assert.Equal(t, 0, e)
	assert.Equal(t, text, r.Text)
}

func TestValidatePhoneFail(t *testing.T) {
	r, e := validatePhone(RowArgs{})
	assert.Equal(t, 1, e)
	assert.Equal(t, "", r.Text)
}

func TestValidatePhoneWrong(t *testing.T) {
	text := "blabla"
	r, e := validatePhone(RowArgs{
		Text: &text,
	})
	assert.Equal(t, 1, e)
	assert.Equal(t, text, r.Text)
}

func TestValidatePhoneOk(t *testing.T) {
	text := "+74951234567"
	r, e := validatePhone(RowArgs{
		Text: &text,
	})
	assert.Equal(t, 0, e)
	assert.Equal(t, text, r.Text)
}

func TestValidateTextFail(t *testing.T) {
	r, e := validateText(RowArgs{})
	assert.Equal(t, 1, e)
	assert.Equal(t, "", r.Text)
}

func TestValidateTextOk(t *testing.T) {
	text := "some text of digits 000111"
	r, e := validateText(RowArgs{
		Text: &text,
	})
	assert.Equal(t, 0, e)
	assert.Equal(t, text, r.Text)
}

func TestValidateCurrencyFail(t *testing.T) {
	r, e := validateCurrency(RowArgs{})
	assert.Equal(t, 1, e)
	assert.Nil(t, r.Amount)
}

func TestValidateCurrencyOk(t *testing.T) {
	float := 1234.56
	r, e := validateCurrency(RowArgs{
		Amount: &float,
	})
	assert.Equal(t, 0, e)
	if assert.NotNil(t, r.Amount) {
		assert.Equal(t, float, *r.Amount)
	}
}

func TestValidateIntFail(t *testing.T) {
	r, e := validateInt(RowArgs{})
	assert.Equal(t, 1, e)
	assert.Nil(t, r.Int)
}

func TestValidateIntOk(t *testing.T) {
	int1 := int64(1234)
	r, e := validateInt(RowArgs{
		Int: &int1,
	})
	assert.Equal(t, 0, e)
	if assert.NotNil(t, r.Int) {
		assert.Equal(t, int1, *r.Int)
	}
}

func TestValidateFloatFail(t *testing.T) {
	r, e := validateFloat(RowArgs{})
	assert.Equal(t, 1, e)
	assert.Nil(t, r.Float)
}

func TestValidateFloatOk(t *testing.T) {
	float := 1234.56
	r, e := validateFloat(RowArgs{
		Float: &float,
	})
	assert.Equal(t, 0, e)
	if assert.NotNil(t, r.Float) {
		assert.Equal(t, float, *r.Float)
	}
}

func TestValidateDateFail(t *testing.T) {
	r, e := validateDate(RowArgs{})
	assert.Equal(t, 1, e)
	assert.Equal(t, "", r.Text)
}

func TestValidateDateWrong(t *testing.T) {
	text := "2018-04-01T12:30:01"
	r, e := validateDate(RowArgs{
		Text: &text,
	})
	assert.Equal(t, 1, e)
	assert.Equal(t, text, r.Text)
}

func TestValidateDateOk(t *testing.T) {
	text := "2018-04-01T12:30:01+03:00"
	r, e := validateDate(RowArgs{
		Text: &text,
	})
	assert.Equal(t, 0, e)
	assert.Equal(t, text, r.Text)
}

func TestValidateDateRangeFail(t *testing.T) {
	r, e := validateDateRange(RowArgs{})
	assert.Equal(t, 1, e)
	assert.Equal(t, "", r.DateFrom)
	assert.Equal(t, "", r.DateTo)
}

func TestValidateDateRangeWrong(t *testing.T) {
	dateFrom := "2018-04-01T12:30:01"
	dateTo := "2018-05-01T12:30:01"
	r, e := validateDateRange(RowArgs{
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
	})
	assert.Equal(t, 2, e)
	assert.Equal(t, dateFrom, r.DateFrom)
	assert.Equal(t, dateTo, r.DateTo)
}

func TestValidateDateRangeOk(t *testing.T) {
	dateFrom := "2018-04-01T12:30:01+03:00"
	dateTo := "2018-05-01T12:30:01+03:00"
	r, e := validateDateRange(RowArgs{
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
	})
	assert.Equal(t, 0, e)
	assert.Equal(t, dateFrom, r.DateFrom)
	assert.Equal(t, dateTo, r.DateTo)
}

func TestValidatePersonNullOk(t *testing.T) {
	r, e := validatePerson(RowArgs{})
	assert.Equal(t, 0, e)
	assert.Equal(t, "", r.Avatar)
	assert.Equal(t, "", r.Role)
}

func TestValidatePersonWrong(t *testing.T) {
	avatar := "blabla"
	role := "developer"
	r, e := validatePerson(RowArgs{
		Avatar: &avatar,
		Role:   &role,
	})
	assert.Equal(t, 1, e)
	assert.Equal(t, avatar, r.Avatar)
	assert.Equal(t, role, r.Role)
}

func TestValidatePersonOk(t *testing.T) {
	avatar := "http://blabla.com/image.jpg"
	role := "developer"
	r, e := validatePerson(RowArgs{
		Avatar: &avatar,
		Role:   &role,
	})
	assert.Equal(t, 0, e)
	assert.Equal(t, avatar, r.Avatar)
	assert.Equal(t, role, r.Role)
}

func TestValidateCompanyNullOk(t *testing.T) {
	r, e := validateCompany(RowArgs{})
	assert.Equal(t, 0, e)
	assert.Equal(t, "", r.Avatar)
	assert.Equal(t, "", r.BGImage)
}

func TestValidateCompanyWrong(t *testing.T) {
	avatar := "blabla"
	bgImage := "foofoo"
	r, e := validateCompany(RowArgs{
		Avatar:  &avatar,
		BGImage: &bgImage,
	})
	assert.Equal(t, 2, e)
	assert.Equal(t, avatar, r.Avatar)
	assert.Equal(t, bgImage, r.BGImage)
}

func TestValidateCompanyOk(t *testing.T) {
	avatar := "http://blabla.com/image.jpg"
	bgImage := "http://blabla.com/bg.png"
	r, e := validateCompany(RowArgs{
		Avatar:  &avatar,
		BGImage: &bgImage,
	})
	assert.Equal(t, 0, e)
	assert.Equal(t, avatar, r.Avatar)
	assert.Equal(t, bgImage, r.BGImage)
}
