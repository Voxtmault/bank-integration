package bca_security

import (
	"context"
	"log"
	"net/http"
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

	timestamp := time.Now().Format(time.RFC3339)
	signature, err := security.CreateSymmetricSignature(context.Background(), &models.SymetricSignatureRequirement{
		HTTPMethod:  http.MethodGet,
		AccessToken: "",
		Timestamp:   timestamp,
		RequestBody: "",
		RelativeURL: "",
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

	timeStamp := "2024-10-15T12:41:00+07:00"
	clientID := "02dcb2b1-bbd5-4d3f-9d2b-a7c95cce5bff"
	signature := "R6K3dwS7PJVgXh3ElrnlBtcvNZAEriP/N29asB5wlDcYgBLvVv/ZCx0zDBhWC+gYZTrVCphQwS0EttnCvH/MMeaLz4BP6Tn38YcsAvXIXGqhmVmtPnUmnLCYGjkegH3xUDQOzkci3GW9GUa5nRlR7FqZ0wm9z1v0fo0+5zFBAPoRX1vmPDcyvPapS6MjJMs4LO1PnxYjpAcE0QEs4dbQu51tVpUmXLlfm+yMe5HWLxn2EvRkfoQKuBR2Tjdy82oxLqN9eNbbLLaeo6KUtRuPpIjutYw0TX4lwSEBPpe1Y7cVuulqCqaCeKkerd66HuqEjx/p6J6ty79cCBzlg4SC3g=="
	result, err := security.VerifyAsymmetricSignature(context.Background(), timeStamp, clientID, signature)
	if err != nil {
		t.Error(err)
	}

	log.Println("Result: ", result)
}
