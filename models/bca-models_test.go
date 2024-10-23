package bank_integration_models

import (
	"testing"
	"time"

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
