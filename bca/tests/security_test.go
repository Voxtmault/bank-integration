package bca_test

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
	"testing"
	"time"

	biModels "github.com/voxtmault/bank-integration/models"
)

func TestCreateAsymmetricSignature(t *testing.T) {
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	timestamp := time.Now().Format(time.RFC3339)
	// timestamp = "2025-01-13T22:58:39+07:00"
	signature, err := security.CreateAsymmetricSignature(context.Background(), timestamp)
	if err != nil {
		t.Error(err)
	}

	log.Println("Timestamp: ", timestamp)
	log.Println("Signature: ", signature)

	// Validate the signature
	result, err := security.VerifyAsymmetricSignature(context.Background(), timestamp, bCfg.BankCredential.ClientID, signature)
	if err != nil {
		t.Error(err)
	}

	log.Println("Result: ", result)

}

func TestCreateSymmetricSignature(t *testing.T) {
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	inputJSON := `{
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
	fmt.Println(bCfg.BankCredential.ClientID)

	timestamp := time.Now().Format(time.RFC3339)
	// timestamp := "2024-10-31T13:34:49+07:00"
	signature, err := security.CreateSymmetricSignature(context.Background(), &biModels.SymmetricSignatureRequirement{
		HTTPMethod:  http.MethodPost,
		AccessToken: "nxVWo8j3QCpvXAiYbQ7UTyEOTNxdnWKl7DUZ8FPPLWJ8Put5dBZHyDO3y_nwCBMQ",
		Timestamp:   timestamp,
		RequestBody: []byte(inputJSON),
		RelativeURL: bCfg.RequestedEndpoints.PaymentFlagURL,
	})
	if err != nil {
		t.Error(err)
	}

	log.Println("Timestamp: ", timestamp)
	log.Println("Signature: ", signature)

	clientSecret := bCfg.BankCredential.ClientSecret
	result, err := security.VerifySymmetricSignature(context.Background(), &biModels.SymmetricSignatureRequirement{
		HTTPMethod:  http.MethodPost,
		AccessToken: "nxVWo8j3QCpvXAiYbQ7UTyEOTNxdnWKl7DUZ8FPPLWJ8Put5dBZHyDO3y_nwCBMQ",
		Timestamp:   timestamp,
		RequestBody: []byte(inputJSON),
		RelativeURL: bCfg.RequestedEndpoints.PaymentFlagURL,
	}, clientSecret, signature)
	if err != nil {
		t.Error(err)
	}

	log.Println("Result: ", result)
}

func TestVerifyAsymmetricSignature(t *testing.T) {
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	timeStamp := "2025-01-25T17:29:24+07:00"
	clientID := bCfg.BankRequestedCredentials.ClientID
	signature := "aB54qttjJIXSMyArnHywqjjT/BGULn3Eldi2XmSTHnDXmesZ3Jakk8Gl+matSb1cs3u2ngByzWF2ZCzrbtUPwK5FmVwH0PzQLAQH1Mv/zgEgJstg26T0OXWhI8gh0Pd1tEqtk4NLI/x9gUSOo0YPS4c9lN+bKi+3lHUekqrgB0dEy60a324zTq3WXvyO1LOXl93KHau8m+Z3qvMAldTYxWsns26suOsRRt1CUmhqSzYgI3wtOE1Tsei3L7qCTnAuxFk8OA52LtX+wIWzKBui9kf08GNLOc4D1sP5UybcVjQkcXvARBJJ52AmTbiPOPQlTKIs7wAaNb74khWgfxf0gQ=="
	result, err := security.VerifyAsymmetricSignature(context.Background(), timeStamp, clientID, signature)
	if err != nil {
		t.Error(err)
	}

	log.Println("Result: ", result)
}

func TestVerifySymmetricSignatureSecurity(t *testing.T) {
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	clientSecret := bCfg.BankRequestedCredentials.ClientSecret

	signature := "7I5uQA06kpl0YK/ExygooNCf52b6CuKXCkrwZQx/nvdn+VjBr8wl2JR52LMJEPtgZr7Ppqvy1OUK8rteZOAb3w=="

	inputJSON := `{"partnerServiceId":"   15335","customerNo":"123456789012345678","virtualAccountNo":"   15335123456789012345678","trxDateInit":"2024-10-23T16:04:00+07:00","channelCode":6011,"language":"","amount":null,"hashedSourceAccountNo":"","sourceBankCode":"014","additionalInfo":{},"passApp":"","inquiryRequestId":"2024102345678984342"}`

	result, err := security.VerifySymmetricSignature(context.Background(), &biModels.SymmetricSignatureRequirement{
		HTTPMethod:  http.MethodPost,
		AccessToken: "P9r-p49whJWG1t5HECd-3kz3F4_MJh1KPs1x8dDoNH8xOGB8ujaOL098GlSjilNY",
		Timestamp:   "2024-10-23T16:04:23+07:00",
		RequestBody: []byte(inputJSON),
		RelativeURL: bCfg.RequestedEndpoints.BillPresentmentURL,
	}, clientSecret, signature)
	if err != nil {
		t.Errorf("Error verifying symmetric signature: %v", err)
	}

	if !result {
		t.Error("Symmetric signature verification failed")
	}

	slog.Debug("Result: ", "data", result)
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
