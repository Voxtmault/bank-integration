package models

import (
	"testing"
	"time"

	"github.com/voxtmault/bank-integration/config"
	"github.com/voxtmault/bank-integration/utils"
)

func TestVerifyBCARequestHeader(t *testing.T) {
	validator := utils.InitValidator()
	cfg := config.New("/home/andy/go-projects/github.com/voxtmault/bank-integration/.env")
	obj := BCARequestHeader{
		Timestamp:   time.Now().Format(time.RFC3339),
		ContentType: "application/json",
		Signature:   "123123123123",
		ClientKey:   cfg.ClientSecret,
		Origin:      "localhost",
	}

	if err := validator.Struct(obj); err != nil {
		t.Errorf("Error validating: %v", err)
	}
}