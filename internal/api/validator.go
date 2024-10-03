package api

import (
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	ValidatorProvider *validator.Validate
}

// In production we don't expose detailed validation error messages.
func (v *Validator) Validate(i interface{}) error {

	return v.ValidatorProvider.Struct(i)
}
