package validator

import (
	"github.com/go-playground/validator/v10"
)

// CustomValidator is a custom validator struct for echo
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator creates a new validator
func NewValidator() *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
	}
}

// Validate validates the given struct
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}