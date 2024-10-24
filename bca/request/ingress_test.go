package bca_request

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"regexp"
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

	// // Set to use mock info
	// security.PrivateKeyPath = "/home/andy/ssl/shifter-wallet/mock_private.pem"
	// security.BCAPublicKeyPath = "/home/andy/ssl/shifter-wallet/mock_public.pub"
	// security.ClientID = cfg.BCARequestedClientCredentials.ClientID
	// security.ClientSecret = cfg.BCARequestedClientCredentials.ClientSecret

	ingress := NewBCAIngress(security)

	// Generate the mock http request
	body := `{"partnerServiceId":"   15335","customerNo":"123456789012345678","virtualAccountNo":"   15335123456789012345678","trxDateInit":"2024-10-23T16:26:00+07:00","channelCode":6011,"language":"","amount":null,"hashedSourceAccountNo":"","sourceBankCode":"014","additionalInfo":{},"passApp":"","inquiryRequestId":"20241023456763236"}`

	mockRequest, err := http.NewRequestWithContext(context.Background(), http.MethodPost, cfg.BaseURL+cfg.BCARequestedEndpoints.BillPresentmentURL, bytes.NewBufferString(body))
	if err != nil {
		t.Errorf("Error creating mock request: %v", err)
	}

	mockRequest.Header.Set("Content-Type", "application/json")
	mockRequest.Header.Set("X-TIMESTAMP", "2024-10-23T16:26:04+07:00")
	mockRequest.Header.Set("Authorization", "Bearer M30N2QBIiM9GKRtT8_XjdDI5eoP7ozN3Sf-xjmgN6oLFhThJXCmHkuiP6QUfd4Mo")
	mockRequest.Header.Set("X-SIGNATURE", "gymKBa1U4RQ8SN5pG02XmdKrXhKBAGy6Dkzxl+NLQ5xcCMADP/KtL9P48eE+lx0hHGliVzxqad/crjPOcOE8PQ==")
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
