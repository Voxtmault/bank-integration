package bca_service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	request "github.com/voxtmault/bank-integration/bca/request"
	security "github.com/voxtmault/bank-integration/bca/security"
	biConfig "github.com/voxtmault/bank-integration/config"
	biModels "github.com/voxtmault/bank-integration/models"
	biStorage "github.com/voxtmault/bank-integration/storage"
	biUtil "github.com/voxtmault/bank-integration/utils"
)

// var envPath = "/home/andy/go-projects/github.com/voxtmault/bank-integration/.env"
var envPath = "../../.env"

func TestGetAccessToken(t *testing.T) {
	cfg := biConfig.New(envPath)
	biUtil.InitValidator()
	biStorage.InitMariaDB(&cfg.MariaConfig)
	biStorage.InitRedis(&cfg.RedisConfig)

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	security := security.NewBCASecurity(cfg)

	service, _ := NewBCAService(
		request.NewBCAEgress(security),
		request.NewBCAIngress(security),
		cfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)

	if err := service.GetAccessToken(context.Background()); err != nil {
		t.Errorf("Error getting access token: %v", err)
	}
}

func TestBalanceInquiry(t *testing.T) {
	cfg := biConfig.New(envPath)
	biUtil.InitValidator()
	biStorage.InitMariaDB(&cfg.MariaConfig)
	biStorage.InitRedis(&cfg.RedisConfig)

	security := security.NewBCASecurity(cfg)

	service, _ := NewBCAService(
		request.NewBCAEgress(security),
		request.NewBCAIngress(security),
		cfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)

	data, err := service.BalanceInquiry(context.Background(), &biModels.BCABalanceInquiry{})
	if err != nil {
		t.Errorf("Error getting access token: %v", err)
	}

	log.Println(data)
}

func TestBalanceInquiryUnmarshal(t *testing.T) {

	// sample := `
	// {"responseCode":"4001102","responseMessage":"Invalid Mandatory Field accountNo","referenceNo":"","partnerReferenceNo":"2020102900000000000001","accountNo":"1234567890","name":"","accountInfos":{"balanceType":"","amount":{"value":"","currency":""},"floatAmount":{"value":"","currency":""},"holdAmount":{"value":"","currency":""},"availableBalance":{"value":"","currency":""},"ledgerBalance":{"value":"","currency":""},"currentMultilateralLimit":{"value":"","currency":""},"registrationStatusCode":"","status":""}}
	// `
	// sample := `
	// {"responseCode":"5001100","responseMessage":"General error","referenceNo":"","partnerReferenceNo":"2020102900000000000001","accountNo":"1234567890","name":"","accountInfos":{"balanceType":"","amount":{"value":"","currency":""},"floatAmount":{"value":"","currency":""},"holdAmount":{"value":"","currency":""},"availableBalance":{"value":"","currency":""},"ledgerBalance":{"value":"","currency":""},"currentMultilateralLimit":{"value":"","currency":""},"registrationStatusCode":"","status":""}}
	// `
	sample := `
	{"responseCode":"2001100","responseMessage":"Successful","referenceNo":"2020102977770000000009","partnerReferenceNo":"2020102900000000000001","accountNo":"1234567890","name":"ANDHIKA","accountInfos":{"balanceType":"Cash","amount":{"value":"100000.00","currency":"IDR"},"floatAmount":{"value":"500000.00","currency":"IDR"},"holdAmount":{"value":"200000.00","currency":"IDR"},"availableBalance":{"value":"200000.00","currency":"IDR"},"ledgerBalance":{"value":"200000.00","currency":"IDR"},"currentMultilateralLimit":{"value":"200000.00","currency":"IDR"},"registrationStatusCode":"0001","status":"0001"}}
	`

	var obj biModels.BCAAccountBalance
	if err := json.Unmarshal([]byte(sample), &obj); err != nil {
		t.Errorf("Error unmarshalling: %v", err)
	}

	log.Printf("%+v\n", obj)
}

func TestBillPresentmentCore(t *testing.T) {
	cfg := biConfig.New(envPath)
	biUtil.InitValidator()
	biStorage.InitMariaDB(&cfg.MariaConfig)

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	s := BCAService{DB: biStorage.GetDBConnection()}
	bodyReq := `{"partnerServiceId":"   15335","customerNo":"123456789012345678","virtualAccountNo":"   15335081234567891234567","trxDateInit":"2024-10-24T11:31:00+07:00","channelCode":6011,"language":"","amount":null,"hashedSourceAccountNo":"","sourceBankCode":"014","additionalInfo":{},"passApp":"","inquiryRequestId":"20241024568326673"}`

	var payload biModels.BCAVARequestPayload
	if err := json.Unmarshal([]byte(bodyReq), &payload); err != nil {
		t.Errorf("Error unmarshalling: %v", err)
	}

	var response biModels.VAResponsePayload
	response.VirtualAccountData = &biModels.VABCAResponseData{
		BillDetails:           []biModels.BillInfo{},
		FreeTexts:             []biModels.FreeText{},
		InquiryReason:         biModels.InquiryReason{},
		SubCompany:            "00000",
		VirtualAccountTrxType: "C",
		CustomerNo:            payload.CustomerNo,
		VirtualAccountNo:      payload.VirtualAccountNo,
		PartnerServiceID:      payload.PartnerServiceID,
		InquiryRequestID:      payload.InquiryRequestID,
		InquiryStatus:         "01", // Default to failure
		TotalAmount:           biModels.Amount{},
		AdditionalInfo:        map[string]interface{}{},
	}

	err := s.BillPresentmentCore(context.Background(), &response, &payload)
	if err != nil {
		t.Errorf("Error From function Bill: %v", err)
	}

	result, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Error From Marshal: %v", err)
	}

	fmt.Println(string(result))
}

func TestInquiryVa(t *testing.T) {
	cfg := biConfig.New(envPath)
	biUtil.InitValidator()
	biStorage.InitMariaDB(&cfg.MariaConfig)
	biStorage.InitRedis(&cfg.RedisConfig)

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	security := security.NewBCASecurity(cfg)

	service, _ := NewBCAService(
		request.NewBCAEgress(security),
		request.NewBCAIngress(security),
		cfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)

	body := `{"partnerServiceId":"   15335","customerNo":"050000000000000012","virtualAccountNo":"   15335050000000000000012","virtualAccountName":"Pemesanan-12","virtualAccountEmail":"","virtualAccountPhone":"","trxId":"","paymentRequestId":"20241028345467246571256","channelCode":6014,"hashedSourceAccountNo":"","sourceBankCode":"014","paidAmount":{"value":"15000.00","currency":"IDR"},"cumulativePaymentAmount":null,"paidBills":"","totalAmount":{"value":"15000.00","currency":"IDR"},"trxDateTime":"2024-10-31T10:27:00+07:00","referenceNo":"24657125601","journalNum":"","paymentType":"","flagAdvise":"N","subCompany":"00000","billDetails":"","freeTexts":"","additionalInfo":""}`

	mockRequest, err := http.NewRequestWithContext(context.Background(), http.MethodPost, cfg.BaseURL+cfg.BCARequestedEndpoints.PaymentFlagURL, bytes.NewBuffer([]byte(body)))
	if err != nil {
		t.Errorf("Error creating mock request: %v", err)
	}

	mockRequest.Header.Set("Content-Type", "application/json")
	mockRequest.Header.Set("X-PARTNER-ID", "15335")
	mockRequest.Header.Set("CHANNEL-ID", "95231")
	mockRequest.Header.Set("X-EXTERNAL-ID", "546785453654")
	mockRequest.Header.Set("X-TIMESTAMP", "2024-10-31T13:34:49+07:00")
	mockRequest.Header.Set("Authorization", "Bearer GPOGvZrlLcvs_Dhi7ju9nKkIkXlOXi-2C4Capr3PaXzCdvAsQS-OtnWgVTN04o3I")
	mockRequest.Header.Set("X-SIGNATURE", "KraW/u8f2622zI3wrF68EIeP4Z873SzoH9zodc8NWL4uMM8VwLifNEDDBDIqNgLa8Mjesvi9uw0/AD2wn4Xgkw==")

	data, err := service.InquiryVA(context.Background(), mockRequest)
	if err != nil {
		t.Errorf("Error getting access token: %v", err)
	}

	result, _ := json.Marshal(data)
	fmt.Println(string(result))

}

func TestInquiryVACore(t *testing.T) {
	cfg := biConfig.New(envPath)
	biUtil.InitValidator()
	biStorage.InitMariaDB(&cfg.MariaConfig)

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	s := BCAService{DB: biStorage.GetDBConnection()}

	bodyReq := `{"partnerServiceId":"   15335","customerNo":"050000000000000012","virtualAccountNo":"   15335123456789012345678","virtualAccountName":"Pemesanan-12","virtualAccountEmail":"","virtualAccountPhone":"","trxId":"","paymentRequestId":"20241024568326673","channelCode":6014,"hashedSourceAccountNo":"","sourceBankCode":"014","paidAmount":{"value":"10000.00","currency":"IDR"},"cumulativePaymentAmount":null,"paidBills":"","totalAmount":{"value":"15000.00","currency":"IDR"},"trxDateTime":"2024-10-31T10:27:00+07:00","referenceNo":"24657125601","journalNum":"","paymentType":"","flagAdvise":"N","subCompany":"00000","billDetails":"","freeTexts":"","additionalInfo":""}`
	var obj biModels.BCAInquiryRequest
	if err := json.Unmarshal([]byte(bodyReq), &obj); err != nil {
		t.Errorf("Error unmarshalling: %v", err)
	}

	var res biModels.BCAInquiryVAResponse
	res.VirtualAccountData = &biModels.VirtualAccountDataInquiry{
		BillDetails:       obj.BillDetails,
		FreeTexts:         obj.FreeTexts,
		PaymentRequestID:  obj.PaymentRequestID,
		ReferenceNo:       obj.ReferenceNo,
		CustomerNo:        obj.CustomerNo,
		VirtualAccountNo:  obj.VirtualAccountNo,
		TrxDateTime:       obj.TrxDateTime,
		PartnerServiceID:  obj.PartnerServiceID,
		PaidAmount:        biModels.Amount{},
		TotalAmount:       biModels.Amount{},
		PaymentFlagReason: biModels.Reason{},
		FlagAdvise:        "N",
	}
	res.AdditionalInfo = obj.AdditionalInfo

	if err := s.InquiryVACore(context.Background(), &res, &obj); err != nil {
		t.Errorf("Error From function Bill: %v", err)
	}

	result, _ := json.Marshal(res)

	fmt.Println(string(result))
}

func TestGenerateAccessToken(t *testing.T) {
	// Logic
	// 1. Load the configs
	// 2. Initiate required services (DB, Validations, etc...)
	// 3. Initiate the BCA Service
	// 4. Generate a mock signature using the BCA Requested Client ID and Secret
	// 5. Generate a mock http request using the data received from (4)
	// 6. Call the Generate Access Token function

	cfg := biConfig.New(envPath)
	biUtil.InitValidator()
	biStorage.InitMariaDB(&cfg.MariaConfig)
	biStorage.InitRedis(&cfg.RedisConfig)

	// Load Registered Banks

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	bcaSec := security.NewBCASecurity(cfg)

	service, _ := NewBCAService(
		request.NewBCAEgress(bcaSec),
		request.NewBCAIngress(bcaSec),
		cfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)

	mockSecurity := security.NewBCASecurity(
		cfg,
	)
	mockSecurity.PrivateKeyPath = "/home/andy/ssl/shifter-wallet/mock_private.pem"
	mockSecurity.BCAPublicKeyPath = "/home/andy/ssl/shifter-wallet/snap_sign.devapi.klikbca.com.pem"
	mockSecurity.ClientID = cfg.BCARequestedClientCredentials.ClientID
	mockSecurity.ClientSecret = cfg.BCARequestedClientCredentials.ClientSecret

	// Generate the mock signature
	timeStamp := "2024-10-23T15:15:42+07:00"
	mockSignature, err := mockSecurity.CreateAsymmetricSignature(context.Background(), timeStamp)
	if err != nil {
		t.Errorf("Error generating mock signature: %v", err)
	}

	// Generate the mock http request
	body := biModels.GrantType{
		GrantType: "client_credentials",
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Errorf("Error marshalling body: %v", err)
	}
	mockRequest, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Errorf("Error creating mock request: %v", err)
	}

	mockRequest.Header.Set("Content-Type", "application/json")
	mockRequest.Header.Set("X-TIMESTAMP", timeStamp)
	mockRequest.Header.Set("X-CLIENT-KEY", mockSecurity.ClientID)
	mockRequest.Header.Set("X-SIGNATURE", "oTNgXCLPXEkiqV1UVV9qRodxukUHhcixToOqfdWhWkfFrygOFjjmPtzG/ec2ZZrVLCGtIHoQUwf9FmKNvh7WvVddAqLa08zvPvzrkWWBPEYcOrJgtmrQbmWOk+CTMEcO9CDHHbz7NfwXQwnj2gEz2oeSWj0yadZxjbhv1ar578ukQ8hxiItk0bHdAnc+M2OtTl3fK8NaADpaZg+7ZOdNh4uiF4jxlNEVqQ0F9+MgIW+pbP73ynMC+WaJ17f4O/k8nUuB81sekeqpd9hSG6gJvx/DF4D9NCbzn3Ty5p+c4t0AUJh5WzEowBJ7l0WwTVHQJr+/IjV98HANMklqVwaU7Q==")

	// Call the Generate Access Token function
	data, err := service.GenerateAccessToken(context.Background(), mockRequest)
	if err != nil {
		t.Errorf("Error generating access token: %v", err)
	}

	slog.Debug("timestamp", "data", timeStamp)
	slog.Debug("signature", "data", mockSignature)
	marshalled, _ := json.Marshal(data)
	slog.Debug("response", "data", marshalled)
}

func TestValidateAccessToken(t *testing.T) {
	cfg := biConfig.New(envPath)
	biUtil.InitValidator()
	biStorage.InitMariaDB(&cfg.MariaConfig)
	biStorage.InitRedis(&cfg.RedisConfig)

	// Load Registered Banks

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	security := security.NewBCASecurity(cfg)

	service, _ := NewBCAService(
		request.NewBCAEgress(security),
		request.NewBCAIngress(security),
		cfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)

	result, err := service.Ingress.ValidateAccessToken(context.Background(), biStorage.GetRedisInstance(), "QyAuKj2Ph0dYkwZ-zozRTg85FC86nfd43qFPqj_dwAKnCIrKg1I4TxSxOeFiZt1F")
	if err != nil {
		t.Errorf("Error validating access token: %v", err)
	}

	slog.Debug("response", "data", fmt.Sprintf("%+v", result))
}

func TestMockRequest(t *testing.T) {
	cfg := biConfig.New(envPath)
	biUtil.InitValidator()
	biStorage.InitMariaDB(&cfg.MariaConfig)
	biStorage.InitRedis(&cfg.RedisConfig)

	// Load Registered Banks

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	security := security.NewBCASecurity(cfg)

	service, _ := NewBCAService(
		request.NewBCAEgress(security),
		request.NewBCAIngress(security),
		cfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)

	mockSecurity := security
	mockSecurity.PrivateKeyPath = "/home/andy/ssl/shifter-wallet/mock_private.pem"
	mockSecurity.BCAPublicKeyPath = "/home/andy/ssl/shifter-wallet/mock_public.pem"
	mockSecurity.ClientID = "c3e7fe0d-379c-4ce2-ad85-372fea661aa0"
	mockSecurity.ClientSecret = "3fd9d63c-f4f1-4c26-8886-fecca45b1053"

	// Generate the mock signature
	timeStamp := time.Now().Format(time.RFC3339)
	mockSignature, err := mockSecurity.CreateSymmetricSignature(context.Background(), &biModels.SymmetricSignatureRequirement{
		HTTPMethod:  http.MethodPost,
		AccessToken: "mAClHNe62u6L6jUuHjzkQ37YTzP49YRGaikd9d0A_hc2uunz44x6554H1ZkZOeAs",
		Timestamp:   "2024-10-21T15:38:03+07:00",
		RelativeURL: "/payment-api/v1.0/transfer-va/inquiry",
		RequestBody: []byte(""),
	})
	if err != nil {
		t.Errorf("Error generating mock signature: %v", err)
	}

	// Generate the mock http request
	body := `{
	"partnerServiceId": " 11223",
	"customerNo": "1234567890123456",
	"virtualAccountNo": " 112231234567890123457",
	"inquiryRequestId": "202410180000000000001"
	}`

	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Errorf("Error marshalling body: %v", err)
	}
	mockRequest, err := http.NewRequestWithContext(context.Background(), http.MethodPost, cfg.BaseURL+cfg.BCAURLEndpoints.BalanceInquiryURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Errorf("Error creating mock request: %v", err)
	}

	mockRequest.Header.Set("Content-Type", "application/json")
	mockRequest.Header.Set("X-TIMESTAMP", timeStamp)
	mockRequest.Header.Set("Authorization", "mAClHNe62u6L6jUuHjzkQ37YTzP49YRGaikd9d0A_hc2uunz44x6554H1ZkZOeAs")
	mockRequest.Header.Set("X-SIGNATURE", mockSignature)
	mockRequest.Header.Set("X-EXTERNAL-ID", "12312321312")

	// Call the validate symmetric signature function
	result, response := service.Ingress.VerifySymmetricSignature(context.Background(), mockRequest, biStorage.GetRedisInstance(), []byte(body))

	slog.Debug("response", "data", fmt.Sprintf("%+v", result))
	slog.Debug("response", "data", fmt.Sprintf("%+v", response))

	if response != nil && response.HTTPStatusCode != http.StatusOK {
		t.Errorf("Error validating symmetric signature: %v", response)
	}
}

// func TestBillPresentmentIntegration(t *testing.T) {
// 	cfg := biConfig.New(envPath)
// 	biUtil.InitValidator()
// 	biStorage.InitMariaDB(&cfg.MariaConfig)
// 	biStorage.InitRedis(&cfg.RedisConfig)

// 	// Load Registered Banks

// 	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
// 		slog.SetLogLoggerLevel(slog.LevelDebug)
// 	} else {
// 		slog.SetLogLoggerLevel(slog.LevelInfo)
// 	}

// 	security := security.NewBCASecurity(cfg)

// 	service := NewBCAService(
// 		request.NewBCAEgress(security),
// 		request.NewBCAIngress(security),
// 		cfg,
// 		biStorage.GetDBConnection(),
// 		biStorage.GetRedisInstance(),
// 	)

// 	// Generate the mock http request
// 	body := `{"partnerServiceId":"   15335","customerNo":"123456789012345678","virtualAccountNo":"   15335123456789012345678","trxDateInit":"2024-10-23T16:23:00+07:00","channelCode":6011,"language":"","amount":null,"hashedSourceAccountNo":"","sourceBankCode":"014","additionalInfo":{},"passApp":"","inquiryRequestId":"20241023566563457"}`

// 	mockRequest, err := http.NewRequestWithContext(context.Background(), http.MethodPost, cfg.BaseURL+cfg.BCARequestedEndpoints.BillPresentmentURL, bytes.NewBufferString(body))
// 	if err != nil {
// 		t.Errorf("Error creating mock request: %v", err)
// 	}

// 	mockRequest.Header.Set("Content-Type", "application/json")
// 	mockRequest.Header.Set("X-TIMESTAMP", "2024-10-23T16:23:40+07:00")
// 	mockRequest.Header.Set("Authorization", "Bearer M30N2QBIiM9GKRtT8_XjdDI5eoP7ozN3Sf-xjmgN6oLFhThJXCmHkuiP6QUfd4Mo")
// 	mockRequest.Header.Set("X-SIGNATURE", "/IZrbCFa/X1kdI1B0IqEvipJ9eKI0eHv8GzzXUzI00qGpUPGRJJTP+Czg687UMkYBx5hZgAapU8KVFOmoOuSEg==")
// 	mockRequest.Header.Set("X-EXTERNAL-ID", "21234")

// 	_, err = service.BillPresentment(context.Background(), data)
// 	if err != nil {
// 		t.Errorf("Error bill presentment: %v", err)
// 	}
// }

func TestInquiryVAIntegration(t *testing.T) {
	cfg := biConfig.New(envPath)
	biUtil.InitValidator()
	biStorage.InitMariaDB(&cfg.MariaConfig)
	biStorage.InitRedis(&cfg.RedisConfig)

	// Load Registered Banks

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	security := security.NewBCASecurity(cfg)

	service, _ := NewBCAService(
		request.NewBCAEgress(security),
		request.NewBCAIngress(security),
		cfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)

	// Generate the mock http request
	body := `{"partnerServiceId":"   12345","customerNo":"123456789012345678","virtualAccountNo":"   12345123456789012345678","virtualAccountName":"Jokul Doe","virtualAccountEmail":"","virtualAccountPhone":"","trxId":"","paymentRequestId":"202202111031031234500001136962","channelCode":6011,"hashedSourceAccountNo":"","sourceBankCode":"014","paidAmount":{"value":"100000.00","currency":"IDR"},"cumulativePaymentAmount":null,"paidBills":"","totalAmount":{"value":"100000.00","currency":"IDR"},"trxDateTime":"2022-02-12T17:29:57+07:00","referenceNo":"00113696201","journalNum":"","paymentType":"","flagAdvise":"N","subCompany":"00000","billDetails":[{"billCode":"","billNo":"123456789012345678","billName":"","billShortName":"","billDescription":{"english":"Maintenance","indonesia":"Pemeliharaan"},"billSubCompany":"00000","billAmount":{"value":"100000.00","currency":"IDR"},"additionalInfo":{},"billReferenceNo":"00113696201"}],"freeTexts":[],"additionalInfo":{}}`

	// payload, _ := json.Marshal(jsonData)
	mockRequest, err := http.NewRequestWithContext(context.Background(), http.MethodPost, cfg.BaseURL+cfg.BCARequestedEndpoints.PaymentFlagURL, bytes.NewBuffer([]byte(body)))
	if err != nil {
		t.Errorf("Error creating mock request: %v", err)
	}

	mockRequest.Header.Set("Content-Type", "application/json")
	mockRequest.Header.Set("X-TIMESTAMP", "2024-10-22T15:42:48+07:00")
	mockRequest.Header.Set("Authorization", "PkEA2fLzAhkTEmUDdmG4eMcKNronHi8US-p5cGT_YMoqTqwwcNw9rizl57bvaMmk")
	mockRequest.Header.Set("X-SIGNATURE", "T2kXLJDL5pAuPyYPueGF5p4XDjNTAlDTRCaLiXPjuUqB3vEbE3r1TSypbb9Vp73CSPHIRX+LvHokG50WrJGvdg==")
	mockRequest.Header.Set("X-EXTERNAL-ID", "765443")

	result, err := service.InquiryVA(context.Background(), mockRequest)
	if err != nil {
		t.Errorf("Error bill presentment: %v", err)
	}

	slog.Debug("response", "data", fmt.Sprintf("%+v", result))
}

func TestCreateVA(t *testing.T) {
	cfg := biConfig.New(envPath)
	biUtil.InitValidator()
	biStorage.InitMariaDB(&cfg.MariaConfig)
	biStorage.InitRedis(&cfg.RedisConfig)

	// Load Registered Banks

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	security := security.NewBCASecurity(cfg)

	service, _ := NewBCAService(
		request.NewBCAEgress(security),
		request.NewBCAIngress(security),
		cfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)

	// Generate CreateVA object
	data := biModels.CreateVAReq{
		IdUser:           12,
		NamaUser:         "Joko Dono",
		IdJenisPembelian: 1,
		IdJenisUser:      1,
		JumlahPembayaran: 100000,
	}

	err := service.CreateVA(context.Background(), &data)
	if err != nil {
		t.Errorf("Error bill presentment: %v", err)
	}
}

func TestGetWatchedTransaction(t *testing.T) {
	cfg := biConfig.New(envPath)
	biUtil.InitValidator()
	biStorage.InitMariaDB(&cfg.MariaConfig)
	biStorage.InitRedis(&cfg.RedisConfig)
	os.Setenv("TZ", "Asia/Jakarta")

	// Load Registered Banks

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	security := security.NewBCASecurity(cfg)

	service, _ := NewBCAService(
		request.NewBCAEgress(security),
		request.NewBCAIngress(security),
		cfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)

	data := service.GetWatchedTransaction(context.Background())
	for _, watcher := range data {
		slog.Debug("watcher", "data", watcher)
	}

	time.Sleep(60 * time.Minute)
}
