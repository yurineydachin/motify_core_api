package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToPayslip(t *testing.T) {
	payslipExt := &PayslipExtended{
		CompanyName: "Company",
		Role:        "Developer",
	}
	payslipExt.ID = 1
	payslipExt.EmployeeFK = 1
	payslipExt.Currency = "USD"
	assert.Equal(t, payslipExt.Role, "Developer")
	payslip := payslipExt.ToPayslip()
	assert.Equal(t, payslip.ID, uint64(1))
}
