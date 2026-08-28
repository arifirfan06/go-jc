package validator

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

// CompanyEmailValidator adalah fungsi custom validator Gin/go-playground.
// Memastikan email diakhiri dengan domain "@company.co.id".
func CompanyEmailValidator(fl validator.FieldLevel) bool {
	// FLOW 1: Ambil nilai string dari field email yang divalidasi
	email := fl.Field().String()

	// FLOW 2: Periksa apakah email diakhiri dengan suffix "@company.co.id"
	return strings.HasSuffix(strings.ToLower(email), "@company.co.id")
}
