package bca_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"regexp"
	"testing"

	bca_request "github.com/voxtmault/bank-integration/bca/request"
	biStorage "github.com/voxtmault/bank-integration/storage"
)

func TestVerifySymmetricSignature(t *testing.T) {
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	ingress := bca_request.NewBCAIngress(security)

	// Generate the mock http request
	body := `{
		"partnerServiceId": "   15335",
		"customerNo": "00100001010000000042",
		"virtualAccountNo": "   1533500100001010000000042",
		"trxDateInit": "2024-10-24T11:31:00+07:00",
		"channelCode": 6011,
		"language": "",
		"amount": {},
		"hashedSourceAccountNo": "",
		"sourceBankCode": "014",
		"additionalInfo": {},
		"passApp": "",
		"inquiryRequestId": "202411141537491533500047652185"
	}`

	mockRequest, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost, cfg.AppHost+bCfg.RequestedEndpoints.BillPresentmentURL, bytes.NewBufferString(body))
	if err != nil {
		t.Errorf("Error creating mock request: %v", err)
	}

	mockRequest.Header.Set("Content-Type", "application/json")
	mockRequest.Header.Set("X-TIMESTAMP", "2025-01-26T12:03:48+07:00")
	mockRequest.Header.Set("Authorization", "Bearer nxVWo8j3QCpvXAiYbQ7UTyEOTNxdnWKl7DUZ8FPPLWJ8Put5dBZHyDO3y_nwCBMQ")
	mockRequest.Header.Set("X-SIGNATURE", "G4D89JqbqOmloKq3jQSGEQx5xM58eb+xCaDtgl8qVZEzlggpbF75ortYTH32Ua353j+uw6dVv+9h0lfQYPXuGg==")
	mockRequest.Header.Set("X-EXTERNAL-ID", "456763236123")

	result, response := ingress.VerifySymmetricSignature(context.Background(), mockRequest, biStorage.GetRedisInstance(), []byte(body))
	if response != nil {
		t.Errorf("Error verifying symmetric signature: %v", response)
	}

	if !result {
		t.Errorf("Signature verification failed")
	}

	slog.Debug("response", "data", result)
}

func TestValidateExternalID(t *testing.T) {
	externalID := "485899127136373621119798753420345267"

	// Check if the external ID is numeric using regex
	isNumeric := regexp.MustCompile(`^\d+$`).MatchString(externalID)
	if !isNumeric {
		t.Errorf("Expected numeric externalID to pass, but it failed")
	} else {
		slog.Info("externalId is numeric")
	}
}
