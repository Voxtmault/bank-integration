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

// Register custom validation rule

// Custom validation tag name to be used in struct tag
type CustomValidatorName string

const (
	BCAPartnerServiceID     CustomValidatorName = "bcaPartnerServiceID"
	BCAVirtualAccountNumber CustomValidatorName = "bcaVA"
)

// ValidatePartnerServiceID is used AFTER checking for required & startswith tag / rule
func ValidatePartnerServiceID(fl validator.FieldLevel) bool {
	partnerServiceID := fl.Field().String()

	// Check if the partner service ID is number (after the 3 white spaces)
	for _, c := range partnerServiceID[3:] {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

func ValidateBCAVirtualAccountNumber(fl validator.FieldLevel) bool {
	virtualAccountNumber := fl.Field().String()

	// Check if the virtual account number is number (after the 3 white spaces)
	for _, c := range virtualAccountNumber[3:] {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}
