package bca_security

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
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
	security.ClientID = ""
	security.ClientSecret = ""

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
	var parsedJson map[string]interface{}
	if err := json.Unmarshal([]byte(inputJSON), &parsedJson); err != nil {
		t.Error(err)
	}
	slog.Debug("Parsed JSON", "data", parsedJson)

	payload, _ := json.Marshal(parsedJson)

	timestamp := time.Now().Format(time.RFC3339)
	signature, err := security.CreateSymmetricSignature(context.Background(), &biModels.SymmetricSignatureRequirement{
		HTTPMethod:  http.MethodPost,
		AccessToken: "6lKrec35Vrs4upmJuGEfctqGIJImYwQm_FDFhbRz076DvqQsLTpCH4HMk7sNE6XG",
		Timestamp:   time.Now().Format(time.RFC3339),
		RequestBody: payload,
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

	timeStamp := "2024-10-23T15:09:52+07:00"
	clientID := "694facb0-f562-44ae-840f-07253ac52f00"
	signature := "vY2eS7bjseT9S90ClgllXVtT+4Mlyuad9joU7c8k4AMs9sdkZYEd+Mwh+C7vaEc1csZdwsjg9gwseWjM3+c8ZMYBZA7j3HhQWwmJnpd/Z0+uGrTZTASj81+vEXrunlawisFElm4OyzZVzAHtoWmcDVwIH3nAcvidZnFgzrHVKhWpjgQo3vfXUvhi7vV4hDfGqRuL/Vu/INAQlN3Q8Htdhz+/W5npQ2y8nL5tVMU2dgG8nYsjmkzJ4Q5+IFF4m6xdz908CeCq0/Mr1KeurYktb66owVQXM9JEy3mOZzlxmKUR0A+Fff9/TjuPybluOMzj0Oh876hOX4etMUQY9IfTkg=="
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

	clientSecret := "3fd9d63c-f4f1-4c26-8886-fecca45b1053"
	signature := "T2kXLJDL5pAuPyYPueGF5p4XDjNTAlDTRCaLiXPjuUqB3vEbE3r1TSypbb9Vp73CSPHIRX+LvHokG50WrJGvdg=="

	inputJSON := `{"partnerServiceId":"   12345","customerNo":"123456789012345678","virtualAccountNo":"   12345123456789012345678","virtualAccountName":"Jokul Doe","virtualAccountEmail":"","virtualAccountPhone":"","trxId":"","paymentRequestId":"202202111031031234500001136962","channelCode":6011,"hashedSourceAccountNo":"","sourceBankCode":"014","paidAmount":{"value":"100000.00","currency":"IDR"},"cumulativePaymentAmount":null,"paidBills":"","totalAmount":{"value":"100000.00","currency":"IDR"},"trxDateTime":"2022-02-12T17:29:57+07:00","referenceNo":"00113696201","journalNum":"","paymentType":"","flagAdvise":"N","subCompany":"00000","billDetails":[{"billCode":"","billNo":"123456789012345678","billName":"","billShortName":"","billDescription":{"english":"Maintenance","indonesia":"Pemeliharaan"},"billSubCompany":"00000","billAmount":{"value":"100000.00","currency":"IDR"},"additionalInfo":{},"billReferenceNo":"00113696201"}],"freeTexts":[],"additionalInfo":{}}`
	var parsedJson map[string]interface{}
	if err := json.Unmarshal([]byte(inputJSON), &parsedJson); err != nil {
		t.Error(err)
	}
	// slog.Debug("Parsed JSON", "data", parsedJson)

	payload, _ := json.Marshal(parsedJson)

	result, err := security.VerifySymmetricSignature(context.Background(), &biModels.SymmetricSignatureRequirement{
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
