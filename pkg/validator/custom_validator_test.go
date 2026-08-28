package validator_test

import (
	"testing"

	customValidator "go-hris-payroll-system/pkg/validator"

	"github.com/go-playground/validator/v10"
)

type TestStruct struct {
	Email string `validate:"company_email"`
}

func TestCompanyEmailValidator(t *testing.T) {
	v := validator.New()
	_ = v.RegisterValidation("company_email", customValidator.CompanyEmailValidator)

	tests := []struct {
		name    string
		email   string
		isValid bool
	}{
		{"Valid Company Email", "employee@company.co.id", true},
		{"Valid HRD Email", "admin.hrd@company.co.id", true},
		{"Uppercase Domain", "USER@COMPANY.CO.ID", true},
		{"Gmail Domain", "employee@gmail.com", false},
		{"Other Domain", "user@company.com", false},
		{"Empty Email", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TestStruct{Email: tt.email}
			err := v.Struct(s)
			if tt.isValid && err != nil {
				t.Errorf("expected valid for %s, got error: %v", tt.email, err)
			}
			if !tt.isValid && err == nil {
				t.Errorf("expected invalid for %s, got no error", tt.email)
			}
		})
	}
}
