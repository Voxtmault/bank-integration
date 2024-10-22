package bca_security

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/voxtmault/bank-integration/config"
	"github.com/voxtmault/bank-integration/models"
)

func TestCreateAsymmetricSignature(t *testing.T) {
	cfg := config.New("/home/andy/go-projects/github.com/voxtmault/bank-integration/.env")
	security := NewBCASecurity(
		cfg,
	)

	security.PrivateKeyPath = "/home/andy/ssl/shifter-wallet/mock_private.pem"
	security.ClientID = "c3e7fe0d-379c-4ce2-ad85-372fea661aa0"

	timestamp := time.Now().Format(time.RFC3339)
	signature, err := security.CreateAsymmetricSignature(context.Background(), timestamp)
	if err != nil {
		t.Error(err)
	}

	log.Println("Timestamp: ", timestamp)
	log.Println("Signature: ", signature)
}

func TestCreateSymmetricSignature(t *testing.T) {
	cfg := config.New("/home/andy/go-projects/github.com/voxtmault/bank-integration/.env")
	security := NewBCASecurity(
		cfg,
	)

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	security.PrivateKeyPath = "/home/andy/ssl/shifter-wallet/mock_private.pem"
	security.BCAPublicKeyPath = "/home/andy/ssl/shifter-wallet/mock_public.pub"
	security.ClientID = "c3e7fe0d-379c-4ce2-ad85-372fea661aa0"
	security.ClientSecret = "3fd9d63c-f4f1-4c26-8886-fecca45b1053"

	timestamp := time.Now().Format(time.RFC3339)
	signature, err := security.CreateSymmetricSignature(context.Background(), &models.SymmetricSignatureRequirement{
		HTTPMethod:  http.MethodPost,
		AccessToken: "66XsNgG-NddWt_XbSKMwgrfLGlte0tbQQpnVyTeGrT4zxkQCZM7y4AQV_i3sqG2y",
		Timestamp:   time.Now().Format(time.RFC3339),
		RequestBody: models.BCAVARequestPayload{
			PartnerServiceID: "11223",
			CustomerNo:       "1234567890123456",
			VirtualAccountNo: "112231234567890123456",
			InquiryRequestID: "202410180000000000001",
		},
		RelativeURL: cfg.BCARequestedEndpoints.BillPresentmentURL,
	})
	if err != nil {
		t.Error(err)
	}

	log.Println("Timestamp: ", timestamp)
	log.Println("Signature: ", signature)
}

func TestVerifyAsymmetricSignature(t *testing.T) {
	cfg := config.New("/home/andy/go-projects/github.com/voxtmault/bank-integration/.env")
	security := NewBCASecurity(
		cfg,
	)

	timeStamp := "2024-10-21T22:39:08+07:00"
	clientID := "c3e7fe0d-379c-4ce2-ad85-372fea661aa0"
	signature := "F2gzBvhRsulzNGzFeqLQ1jsZNTqv4TSNii8MJ2n7qe50fUPnepUZghbSTKDCLDPZ"
	result, err := security.VerifyAsymmetricSignature(context.Background(), timeStamp, clientID, signature)
	if err != nil {
		t.Error(err)
	}

	log.Println("Result: ", result)
}

func TestVerifySymmetricSignature(t *testing.T) {
	cfg := config.New("/home/andy/go-projects/github.com/voxtmault/bank-integration/.env")
	security := NewBCASecurity(
		cfg,
	)

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	clientSecret := "3fd9d63c-f4f1-4c26-8886-fecca45b1053"
	signature := "3SQ6tjbXGHnycGiaufUT8AwivHY9OrsEOuZ3ZMk7qWkezPiT1T3h4NrhH4Ak7BvbHteM+uOzkdneojZoR9Te2w=="

	result, err := security.VerifySymmetricSignature(context.Background(), &models.SymmetricSignatureRequirement{
		HTTPMethod:  http.MethodPost,
		AccessToken: "mnC56Gqr5CquF6vLNJuX4waBlg_yAItts6ytPMtytAPqnNsymitxRkw3fwlBhFLj",
		Timestamp:   "2024-10-21T23:26:36+07:00",
		RequestBody: `{
    "partnerServiceId": " 11223",
    "customerNo": "1234567890123456",
    "virtualAccountNo": " 112231234567890123457",
    "inquiryRequestId": "202410180000000000001"
}`,
		RelativeURL: "/payment-api/v1.0/transfer-va/inquiry",
	}, clientSecret, signature)
	if err != nil {
		t.Errorf("Error verifying symmetric signature: %v", err)
	}

	if !result {
		t.Error("Symmetric signature verification failed")
	}

	slog.Debug("Result: ", "data", result)
}
