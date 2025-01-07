package bca_security

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	biConfig "github.com/voxtmault/bank-integration/config"
	biModels "github.com/voxtmault/bank-integration/models"
)

var envPath = "../../.env"

func TestCreateAsymmetricSignature(t *testing.T) {
	cfg := biConfig.New(envPath)
	security := NewBCASecurity(
		cfg,
	)

	security.PrivateKeyPath = "../../private-key.pem"
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
	cfg := biConfig.New(envPath)
	security := NewBCASecurity(
		cfg,
	)

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	security.PrivateKeyPath = "../../private-key.pem"
	security.BCAPublicKeyPath = "../../public-key.pem"
	security.ClientID = "c3e7fe0d-379c-4ce2-ad85-372fea661aa0"
	security.ClientSecret = "3fd9d63c-f4f1-4c26-8886-fecca45b1053"

	inputJSON := `{
		"partnerServiceId": "   15335",
		"customerNo": "112233445566778899",
		"virtualAccountNo": "   15335112233445566778899",
		"virtualAccountName": "Budi Sujipto",
		"virtualAccountEmail": "",
		"virtualAccountPhone": "",
		"trxId": "",
		"paymentRequestId": "202411141539271533500047652186",
		"channelCode": 6014,
		"hashedSourceAccountNo": "",
		"sourceBankCode": "014",
		"paidAmount": {
			"value": "15000.00",
			"currency": "IDR"
		},
		"cumulativePaymentAmount": null,
		"paidBills": "",
		"totalAmount": {
			"value": "15000.00",
			"currency": "IDR"
		},
		"trxDateTime": "2024-11-30T10:27:00+07:00",
		"referenceNo": "24657125601",
		"journalNum": "",
		"paymentType": "",
		"flagAdvise": "N",
		"subCompany": "00000",
		"billDetails": "",
		"freeTexts": "",
		"additionalInfo": {}
	}`
	fmt.Println(cfg.BCARequestedClientCredentials.ClientID)

	timestamp := time.Now().Format(time.RFC3339)
	signature, err := security.CreateSymmetricSignature(context.Background(), &biModels.SymmetricSignatureRequirement{
		HTTPMethod:  http.MethodPost,
		AccessToken: "trWIaRJ7Kk6ZrLqtZdja-XDq1lia1PideqpggT-BEQzCLxbp4vXRBsibNnhjvGKF",
		Timestamp:   "2024-11-30T16:17:54+07:00",
		RequestBody: []byte(inputJSON),
		RelativeURL: cfg.BCARequestedEndpoints.PaymentFlagURL,
	})
	if err != nil {
		t.Error(err)
	}

	log.Println("Timestamp: ", timestamp)
	log.Println("Signature: ", signature)
}

func TestVerifyAsymmetricSignature(t *testing.T) {
	cfg := biConfig.New(envPath)
	security := NewBCASecurity(
		cfg,
	)

	timeStamp := "2024-10-23T15:15:42+07:00"
	clientID := cfg.BCARequestedClientCredentials.ClientID
	signature := "F7Tbk9w42OTYKc2TtEypdIkzhAt56ieIxKFFwNLHIu5ItEbzYTsVpuhSrYJmtwFagqN6Jci6eyvkmnG6qRh0LCzB5yVlizN434LzxkJkH2Ug1EcRQzKl5APVXKYb/fybFDVJMV4BbvBA6lgerhL9AxJG2yHoLCOyr3CH+BjiNJF2tXLYi18jwcJKToBK0GY/POU8z16ykaAageqrAzFucuBq6cRq28Y99DSdFTAOrBysniCLY0I1TguVyJjXPIJEn7UJEqc2ZD7lYpS/g//bEfHeVsB01xzyuyuRWTXLL9KEFVqPRtaI3oUg3rZypTwpOUjiJZpBJJBAym/DGgjepQ=="
	result, err := security.VerifyAsymmetricSignature(context.Background(), timeStamp, clientID, signature)
	if err != nil {
		t.Error(err)
	}

	log.Println("Result: ", result)
}

func TestVerifySymmetricSignature(t *testing.T) {
	cfg := biConfig.New(envPath)
	security := NewBCASecurity(
		cfg,
	)

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	clientSecret := cfg.BCARequestedClientCredentials.ClientSecret

	signature := "7I5uQA06kpl0YK/ExygooNCf52b6CuKXCkrwZQx/nvdn+VjBr8wl2JR52LMJEPtgZr7Ppqvy1OUK8rteZOAb3w=="

	inputJSON := `{"partnerServiceId":"   15335","customerNo":"123456789012345678","virtualAccountNo":"   15335123456789012345678","trxDateInit":"2024-10-23T16:04:00+07:00","channelCode":6011,"language":"","amount":null,"hashedSourceAccountNo":"","sourceBankCode":"014","additionalInfo":{},"passApp":"","inquiryRequestId":"2024102345678984342"}`

	result, err := security.VerifySymmetricSignature(context.Background(), &biModels.SymmetricSignatureRequirement{
		HTTPMethod:  http.MethodPost,
		AccessToken: "P9r-p49whJWG1t5HECd-3kz3F4_MJh1KPs1x8dDoNH8xOGB8ujaOL098GlSjilNY",
		Timestamp:   "2024-10-23T16:04:23+07:00",
		RequestBody: []byte(inputJSON),
		RelativeURL: cfg.BCARequestedEndpoints.BillPresentmentURL,
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
	cfg := biConfig.New(envPath)
	security := NewBCASecurity(
		cfg,
	)

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	inputJson := `{ 
  "CorporateID" : "H2HAUTO009", 
  "SourceAccountNumber" : "0611104625", 
  "TransactionID" : "00177914", 
  "TransactionDate" : "2017-03-17", 
  "ReferenceID" : "1234567890098765", 
  "CurrencyCode" : "IDR", 
  "Amount" : "175000000", 
  "BeneficiaryAccountNumber" : "0613106704", 
  "Remark1" : "Pencairan Kredit", 
  "Remark2" : "1234567890098765" 
}`

	data, err := security.customMinifyJSON([]byte(inputJson))
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Minify Result: ", data)
}

func TestProcessingURL(t *testing.T) {
	cfg := biConfig.New(envPath)
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

func TestLoadPublicKey(t *testing.T) {

	path := "/home/andy/ssl/shifter-wallet/snap_sign.devapi.klikbca.com.pem"

	keyData, err := os.ReadFile(path)
	if err != nil {
		t.Error(err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "PUBLIC KEY" {
		t.Error(err)
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Error(err)
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		log.Println("Public Key: ", pub)
	default:
		t.Error("Not an RSA public key")
	}
}
