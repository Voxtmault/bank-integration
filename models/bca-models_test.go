package bank_integration_models

import (
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	biConfig "github.com/voxtmault/bank-integration/config"
	biUtil "github.com/voxtmault/bank-integration/utils"
)

func TestVerifyBCARequestHeader(t *testing.T) {
	validator := biUtil.InitValidator()
	cfg := biConfig.New("/home/andy/go-projects/github.com/voxtmault/bank-integration/.env")
	obj := BCARequestHeader{
		Timestamp:   time.Now().Format(time.RFC3339),
		ContentType: "application/json",
		Signature:   "123123123123",
		ClientKey:   cfg.BCAConfig.ClientID,
		Origin:      "localhost",
	}

	if err := validator.Struct(obj); err != nil {
		t.Errorf("Error validating: %v", err)
	}
}

func TestVerifyBCABillInquiryRequestPayload(t *testing.T) {
	validate := biUtil.InitValidator()
	validate.RegisterValidation("bcaPartnerServiceID", biUtil.ValidatePartnerServiceID)
	validate.RegisterValidation("bcaVA", biUtil.ValidateBCAVirtualAccountNumber)

	sampleRequest := `
	{ 
	"partnerServiceId": "   a1234", 
	"customerNo": "123456789012345678", 
	"virtualAccountNo": "  12345123456789012345678",
	"trxDateInit": "2022-02-12T17:29:57+07:00", 
	"channelCode": 6011, 
	"language": "", 
	"amount": null, 
	"hashedSourceAccountNo": "", 
	"sourceBankCode": "014", 
	"additionalInfo": {}, 
	"passApp": "", 
	"inquiryRequestId": "202202110909311234500001136962" 
	}
	`
	var obj BCAVARequestPayload
	if err := json.Unmarshal([]byte(sampleRequest), &obj); err != nil {
		t.Errorf("error parsing sample request %v", err)
		return
	}

	if err := validate.Struct(obj); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			if err.Tag() == "required" || err.Tag() == "startswith" {
				slog.Warn("invalid mandatory field", "field", err.Field())
			} else {
				slog.Warn("invalid field format", "field", err.Field())
			}
		}
		t.Errorf("error validating struct %s", err)
		return
	}
}
