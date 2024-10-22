package bca_security

import (
	"context"
	"encoding/json"
	"fmt"
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

	inputJSON := `{"partnerServiceId":"   12345","customerNo":"123456789012345678","virtualAccountNo":"   12345123456789012345678","virtualAccountName":"Jokul Doe","virtualAccountEmail":"","virtualAccountPhone":"","trxId":"","paymentRequestId":"202202111031031234500001136962","channelCode":6011,"hashedSourceAccountNo":"","sourceBankCode":"014","paidAmount":{"value":"100000.00","currency":"IDR"},"cumulativePaymentAmount":null,"paidBills":"","totalAmount":{"value":"100000.00","currency":"IDR"},"trxDateTime":"2022-02-12T17:29:57+07:00","referenceNo":"00113696201","journalNum":"","paymentType":"","flagAdvise":"N","subCompany":"00000","billDetails":[{"billCode":"","billNo":"123456789012345678","billName":"","billShortName":"","billDescription":{"english":"Maintenance","indonesia":"Pemeliharaan"},"billSubCompany":"00000","billAmount":{"value":"100000.00","currency":"IDR"},"additionalInfo":{},"billReferenceNo":"00113696201"}],"freeTexts":[],"additionalInfo":{}}`
	var parsedJson map[string]interface{}
	if err := json.Unmarshal([]byte(inputJSON), &parsedJson); err != nil {
		t.Error(err)
	}
	slog.Debug("Parsed JSON", "data", parsedJson)

	payload, _ := json.Marshal(parsedJson)

	timestamp := time.Now().Format(time.RFC3339)
	signature, err := security.CreateSymmetricSignature(context.Background(), &models.SymmetricSignatureRequirement{
		HTTPMethod:  http.MethodPost,
		AccessToken: "PkEA2fLzAhkTEmUDdmG4eMcKNronHi8US-p5cGT_YMoqTqwwcNw9rizl57bvaMmk",
		Timestamp:   time.Now().Format(time.RFC3339),
		RequestBody: payload,
		RelativeURL: cfg.BCARequestedEndpoints.PaymentFlagURL,
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
	signature := "T2kXLJDL5pAuPyYPueGF5p4XDjNTAlDTRCaLiXPjuUqB3vEbE3r1TSypbb9Vp73CSPHIRX+LvHokG50WrJGvdg=="

	inputJSON := `{"partnerServiceId":"   12345","customerNo":"123456789012345678","virtualAccountNo":"   12345123456789012345678","virtualAccountName":"Jokul Doe","virtualAccountEmail":"","virtualAccountPhone":"","trxId":"","paymentRequestId":"202202111031031234500001136962","channelCode":6011,"hashedSourceAccountNo":"","sourceBankCode":"014","paidAmount":{"value":"100000.00","currency":"IDR"},"cumulativePaymentAmount":null,"paidBills":"","totalAmount":{"value":"100000.00","currency":"IDR"},"trxDateTime":"2022-02-12T17:29:57+07:00","referenceNo":"00113696201","journalNum":"","paymentType":"","flagAdvise":"N","subCompany":"00000","billDetails":[{"billCode":"","billNo":"123456789012345678","billName":"","billShortName":"","billDescription":{"english":"Maintenance","indonesia":"Pemeliharaan"},"billSubCompany":"00000","billAmount":{"value":"100000.00","currency":"IDR"},"additionalInfo":{},"billReferenceNo":"00113696201"}],"freeTexts":[],"additionalInfo":{}}`
	var parsedJson map[string]interface{}
	if err := json.Unmarshal([]byte(inputJSON), &parsedJson); err != nil {
		t.Error(err)
	}
	// slog.Debug("Parsed JSON", "data", parsedJson)

	payload, _ := json.Marshal(parsedJson)

	result, err := security.VerifySymmetricSignature(context.Background(), &models.SymmetricSignatureRequirement{
		HTTPMethod:  http.MethodPost,
		AccessToken: "PkEA2fLzAhkTEmUDdmG4eMcKNronHi8US-p5cGT_YMoqTqwwcNw9rizl57bvaMmk",
		Timestamp:   "2024-10-22T15:42:48+07:00",
		RequestBody: payload,
		RelativeURL: cfg.BCARequestedEndpoints.PaymentFlagURL,
	}, clientSecret, signature)
	if err != nil {
		t.Errorf("Error verifying symmetric signature: %v", err)
	}

	if !result {
		t.Error("Symmetric signature verification failed")
	}

	slog.Debug("Result: ", "data", result)
}

func TestMinifyJSON(t *testing.T) {
	cfg := config.New("/home/andy/go-projects/github.com/voxtmault/bank-integration/.env")
	security := NewBCASecurity(
		cfg,
	)

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	inputJson := `{
  "partnerServiceId": "        12345",
  "customerNo": "123456789012345678",
  "virtualAccountNo": "        12345123456789012345678",
  "trxDateInit": "2022-02-12T17:29:57+07:00",
  "channelCode": 6011,
  "language": "",
  "amount": null,
  "hashedSourceAccountNo": "",
  "sourceBankCode": "014",
  "additionalInfo": {},
  "passApp": "",
  "inquiryRequestId": "202202110909311234500001136962"
}`

	// Parse the input JSON string into a Go data structure
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(inputJson), &jsonData); err != nil {
		t.Fatalf("error un-marshaling input JSON: %v", err)
	}
	slog.Debug("jsonData", "data", jsonData)

	data, err := security.customMinifyJSON(jsonData)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Custom Result: ", data)

	data, err = security.minifyJSON(jsonData)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Minify Result: ", data)
}

func TestProcessingURL(t *testing.T) {
	cfg := config.New("/home/andy/go-projects/github.com/voxtmault/bank-integration/.env")
	security := NewBCASecurity(
		cfg,
	)

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	// data := `/banking/v2/corporates/h2hauto009/accounts/0611104625/statements?StartDate=2017-03-01&EndDate=2017-03-017`
	data := "/banking/v2/corporates/h2hauto009/accounts/0611104625,0613106704 "

	data, err := security.processRelativeURL(data)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Result: ", data)
}
