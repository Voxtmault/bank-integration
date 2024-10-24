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

var envPath = "/home/andy/go-projects/github.com/voxtmault/bank-integration/.env"

func TestCreateAsymmetricSignature(t *testing.T) {
	cfg := biConfig.New(envPath)
	security := NewBCASecurity(
		cfg,
	)

	security.PrivateKeyPath = "/home/andy/ssl/shifter-wallet/mock_private.pem"
	security.ClientID = ""

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

	security.PrivateKeyPath = "/home/andy/ssl/shifter-wallet/mock_private.pem"
	security.BCAPublicKeyPath = "/home/andy/ssl/shifter-wallet/mock_public.pub"
	security.ClientID = cfg.BCARequestedClientCredentials.ClientID
	security.ClientSecret = cfg.BCARequestedClientCredentials.ClientSecret

	inputJSON := `
{
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
}
	`

	timestamp := time.Now().Format(time.RFC3339)
	signature, err := security.CreateSymmetricSignature(context.Background(), &biModels.SymmetricSignatureRequirement{
		HTTPMethod:  http.MethodPost,
		AccessToken: "6lKrec35Vrs4upmJuGEfctqGIJImYwQm_FDFhbRz076DvqQsLTpCH4HMk7sNE6XG",
		Timestamp:   time.Now().Format(time.RFC3339),
		RequestBody: []byte(inputJSON),
		RelativeURL: cfg.BCARequestedEndpoints.BillPresentmentURL,
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
	signature := "oTNgXCLPXEkiqV1UVV9qRodxukUHhcixToOqfdWhWkfFrygOFjjmPtzG/ec2ZZrVLCGtIHoQUwf9FmKNvh7WvVddAqLa08zvPvzrkWWBPEYcOrJgtmrQbmWOk+CTMEcO9CDHHbz7NfwXQwnj2gEz2oeSWj0yadZxjbhv1ar578ukQ8hxiItk0bHdAnc+M2OtTl3fK8NaADpaZg+7ZOdNh4uiF4jxlNEVqQ0F9+MgIW+pbP73ynMC+WaJ17f4O/k8nUuB81sekeqpd9hSG6gJvx/DF4D9NCbzn3Ty5p+c4t0AUJh5WzEowBJ7l0WwTVHQJr+/IjV98HANMklqVwaU7Q=="
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
