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
	"github.com/voxtmault/bank-integration/config"
	"github.com/voxtmault/bank-integration/models"
	"github.com/voxtmault/bank-integration/storage"
	"github.com/voxtmault/bank-integration/utils"
)

var envPath = "/home/andy/go-projects/github.com/voxtmault/bank-integration/.env"

func TestVerifySymmetricSignature(t *testing.T) {
	cfg := config.New(envPath)
	utils.InitValidator()
	storage.InitMariaDB(&cfg.MariaConfig)
	storage.InitRedis(&cfg.RedisConfig)

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
	body := models.BCAVARequestPayload{
		PartnerServiceID: "11223",
		CustomerNo:       "1234567890123456",
		VirtualAccountNo: "112231234567890123456",
		InquiryRequestID: "202410180000000000001",
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Errorf("Error marshalling body: %v", err)
	}
	mockRequest, err := http.NewRequestWithContext(context.Background(), http.MethodPost, cfg.BaseURL+cfg.BCAURLEndpoints.BalanceInquiryURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Errorf("Error creating mock request: %v", err)
	}

	mockRequest.Header.Set("Content-Type", "application/json")
	mockRequest.Header.Set("X-TIMESTAMP", "2024-10-21T23:58:32+07:00")
	mockRequest.Header.Set("Authorization", "0HDZIbSpVLDJHv-dfmez5QtX87OrBXdlVO2KNYEZitk95LBp8ChPaCPNcabsLoMV")
	mockRequest.Header.Set("X-SIGNATURE", "Nybi/ZID+rPwObe/aZJ3A71gULNldsiuy3G58yccs5UqSdy2YuJJ+T1HG0zGyBoBoz81npZlUX8KYmpJ5TiPww==")
	mockRequest.Header.Set("X-EXTERNAL-ID", "167123456")

	result, response := ingress.VerifySymmetricSignature(context.Background(), mockRequest, storage.GetRedisInstance(), body)
	if response != nil {
		t.Errorf("Error verifying symmetric signature: %v", response)
	}

	if !result {
		t.Errorf("Signature verification failed")
	}

	slog.Debug("response", "data", result)
}
