package bca_request

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	bcaSecurity "github.com/voxtmault/bank-integration/bca/security"
	biConfig "github.com/voxtmault/bank-integration/config"
	biStorage "github.com/voxtmault/bank-integration/storage"
	biUtil "github.com/voxtmault/bank-integration/utils"
)

var envPath = "/home/andy/go-projects/github.com/voxtmault/bank-integration/.env"

func TestVerifySymmetricSignature(t *testing.T) {
	cfg := biConfig.New(envPath)
	biUtil.InitValidator()
	biStorage.InitMariaDB(&cfg.MariaConfig)
	biStorage.InitRedis(&cfg.RedisConfig)

	security := bcaSecurity.NewBCASecurity(
		cfg,
	)

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	// Set to use mock info
	security.PrivateKeyPath = "/home/andy/ssl/shifter-wallet/mock_private.pem"
	security.BCAPublicKeyPath = "/home/andy/ssl/shifter-wallet/mock_public.pub"
	security.ClientID = "c3e7fe0d-379c-4ce2-ad85-372fea661aa0"
	security.ClientSecret = "3fd9d63c-f4f1-4c26-8886-fecca45b1053"

	ingress := NewBCAIngress(security)

	// Generate the mock http request
	body := `{
  "partnerServiceId": "        11223",
  "customerNo": "1234567890123456",
  "virtualAccountNo": "        112231234567890123456",
  "trxDateInit": "2022-02-12T17:29:57+07:00",
  "channelCode": 6011,
  "language": "",
  "amount": null,
  "hashedSourceAccountNo": "",
  "sourceBankCode": "014",
  "additionalInfo": {},
  "passApp": "",
  "inquiryRequestId": "202410180000000000001"
}`

	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(body), &jsonData); err != nil {
		slog.Debug("error unmarshalling request body", "error", err)
		t.Error(err)
	}

	payload, _ := json.Marshal(jsonData)
	mockRequest, err := http.NewRequestWithContext(context.Background(), http.MethodPost, cfg.BaseURL+cfg.BCARequestedEndpoints.BillPresentmentURL, bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error creating mock request: %v", err)
	}

	mockRequest.Header.Set("Content-Type", "application/json")
	mockRequest.Header.Set("X-TIMESTAMP", "2024-10-22T12:50:37+07:00")
	mockRequest.Header.Set("Authorization", "PkEA2fLzAhkTEmUDdmG4eMcKNronHi8US-p5cGT_YMoqTqwwcNw9rizl57bvaMmk")
	mockRequest.Header.Set("X-SIGNATURE", "NV54FMmgdpMuwshlUCgIMSXlJpH3s/X3bj3IzHqpHVmaA/PAIIgq5ICIZlwm5nM8/y503+h88Q1pP3NO5nlVLA==")
	mockRequest.Header.Set("X-EXTERNAL-ID", "32131")

	result, response := ingress.VerifySymmetricSignature(context.Background(), mockRequest, biStorage.GetRedisInstance(), payload)
	if response != nil {
		t.Errorf("Error verifying symmetric signature: %v", response)
	}

	if !result {
		t.Errorf("Signature verification failed")
	}

	slog.Debug("response", "data", result)
}
