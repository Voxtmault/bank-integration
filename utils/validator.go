package bank_integration_utils

import (
	"context"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func InitValidator() *validator.Validate {
	validate = validator.New()

	return validate
}

func GetValidator() *validator.Validate {
	if validate == nil {
		return InitValidator()
	} else {
		return validate
	}
}

func ValidateStruct(ctx context.Context, s interface{}) error {
	return validate.Struct(s)
}
