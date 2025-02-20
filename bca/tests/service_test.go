package bca_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	request "github.com/voxtmault/bank-integration/bca/request"
	bca_service "github.com/voxtmault/bank-integration/bca/service"
	biModels "github.com/voxtmault/bank-integration/models"
	biStorage "github.com/voxtmault/bank-integration/storage"
)

func TestGetAccessToken(t *testing.T) {
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	service, _ := bca_service.NewBCAService(
		request.NewBCAEgress(security, bCfg, cfg),
		request.NewBCAIngress(security),
		cfg,
		bCfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)

	if err := service.GetAccessToken(context.Background()); err != nil {
		t.Errorf("Error getting access token: %v", err)
	}

	time.Sleep(time.Second * 3)
}

func TestBalanceInquiry(t *testing.T) {
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	service, _ := bca_service.NewBCAService(
		request.NewBCAEgress(security, bCfg, cfg),
		request.NewBCAIngress(security),
		cfg,
		bCfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)

	data, err := service.BalanceInquiry(context.Background())
	if err != nil {
		t.Errorf("Error getting access token: %v", err)
	}

	jsonStr, _ := json.Marshal(data)

	log.Println(string(jsonStr))
}

func TestBankStatement(t *testing.T) {
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	service, _ := bca_service.NewBCAService(
		request.NewBCAEgress(security, bCfg, cfg),
		request.NewBCAIngress(security),
		cfg,
		bCfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)

	fromDateTime := time.Now().AddDate(0, 0, -13).Format(time.RFC3339)
	toDateTime := time.Now().Format(time.RFC3339)
	data, err := service.BankStatement(context.Background(), fromDateTime, toDateTime)
	if err != nil {
		t.Errorf("Error getting bank statement: %v", err)
	}

	jsonStr, _ := json.Marshal(data)

	log.Println(string(jsonStr))
}

func TestGetVAPaymentStatus(t *testing.T) {
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	service, _ := bca_service.NewBCAService(
		request.NewBCAEgress(security, bCfg, cfg),
		request.NewBCAIngress(security),
		cfg,
		bCfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)

	data, err := service.GetVAPaymentStatus(context.Background(), "   7510020221007001")
	if err != nil {
		t.Errorf("Error getting va payment status: %v", err)
	}

	jsonStr, _ := json.Marshal(data)

	log.Println(string(jsonStr))
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

func TestBillPresentmentIntegration(t *testing.T) {
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	service, _ := bca_service.NewBCAService(
		request.NewBCAEgress(security, bCfg, cfg),
		request.NewBCAIngress(security),
		cfg,
		bCfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)

	body := `{
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

	mockRequest, err := http.NewRequestWithContext(context.Background(), http.MethodPost, cfg.AppHost+bCfg.RequestedEndpoints.PaymentFlagURL, bytes.NewBuffer([]byte(body)))
	if err != nil {
		t.Errorf("Error creating mock request: %v", err)
	}

	mockRequest.Header.Set("Content-Type", "application/json")
	mockRequest.Header.Set("X-PARTNER-ID", "15335")
	mockRequest.Header.Set("CHANNEL-ID", "95231")
	mockRequest.Header.Set("X-EXTERNAL-ID", "546785453654")
	mockRequest.Header.Set("X-TIMESTAMP", "2025-01-26T11:58:00+07:00")
	mockRequest.Header.Set("Authorization", "Bearer nxVWo8j3QCpvXAiYbQ7UTyEOTNxdnWKl7DUZ8FPPLWJ8Put5dBZHyDO3y_nwCBMQ")
	mockRequest.Header.Set("X-SIGNATURE", "HOJ+4/qdN66V3HxINFvl/EcBRgZNMyFsx9JEhUbwNwesgoel9luARHDAFOsXNlOQZ2/c+Jzv/XQvNe1raXkjiA==")

	data, err := service.BillPresentment(context.Background(), mockRequest)
	if err != nil {
		t.Errorf("Error in bill presentment: %v", err)
	}

	result, _ := json.Marshal(data)
	fmt.Println(string(result))
}

func TestBillPresentmentCore(t *testing.T) {
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	s, err := bca_service.NewBCAService(
		request.NewBCAEgress(security, bCfg, cfg),
		request.NewBCAIngress(security),
		cfg,
		bCfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)
	if err != nil {
		t.Errorf("Error creating BCA Service: %v", err)
	}

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

	err = s.BillPresentmentCore(context.Background(), &response, &payload)
	if err != nil {
		t.Errorf("Error From function Bill: %v", err)
	}

	result, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Error From Marshal: %v", err)
	}

	fmt.Println(string(result))
}

func TestInquiryVAIntegration(t *testing.T) {
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	s, err := bca_service.NewBCAService(
		request.NewBCAEgress(security, bCfg, cfg),
		request.NewBCAIngress(security),
		cfg,
		bCfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)
	if err != nil {
		t.Errorf("Error creating BCA Service: %v", err)
	}

	body := `
	{
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
	}
	`

	mockRequest, err := http.NewRequestWithContext(context.Background(), http.MethodPost, cfg.AppHost+bCfg.RequestedEndpoints.PaymentFlagURL, bytes.NewBuffer([]byte(body)))
	if err != nil {
		t.Errorf("Error creating mock request: %v", err)
	}

	mockRequest.Header.Set("Content-Type", "application/json")
	mockRequest.Header.Set("X-PARTNER-ID", "15335")
	mockRequest.Header.Set("CHANNEL-ID", "95231")
	mockRequest.Header.Set("X-EXTERNAL-ID", "546785453654")
	mockRequest.Header.Set("X-TIMESTAMP", "2024-11-30T15:26:39+07:00")
	mockRequest.Header.Set("Authorization", "Bearer B9njeONWOirGwBce05lU6CLZn6tL1g8gqvO194zTBVfEL107RSsRjNMdvJOqiXcc")
	mockRequest.Header.Set("X-SIGNATURE", "FvjgnqcharXlNjI8ZvRywF44smfbYI/mIGr3JBH5QrLSZhIckV+VE5Ipaifir7isilUE/xVGQaEHWu6+e54X/g==")

	data, err := s.InquiryVA(context.Background(), mockRequest)
	if err != nil {
		t.Errorf("Error getting access token: %v", err)
	}

	result, _ := json.Marshal(data)
	fmt.Println(string(result))
}

func TestInquiryVACore(t *testing.T) {
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	s, err := bca_service.NewBCAService(
		request.NewBCAEgress(security, bCfg, cfg),
		request.NewBCAIngress(security),
		cfg,
		bCfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)
	if err != nil {
		t.Errorf("Error creating BCA Service: %v", err)
	}

	bodyReq := `{
		"partnerServiceId": "   15335",
		"customerNo": "001000020050000016",
		"virtualAccountNo": "   15335001000020050000016",
		"virtualAccountName": "Atiko Dono Dono",
		"virtualAccountEmail": "",
		"virtualAccountPhone": "",
		"trxId": "",
		"paymentRequestId": "202411141539271533500047652187",
		"channelCode": 6014,
		"hashedSourceAccountNo": "",
		"sourceBankCode": "014",
		"paidAmount": {
			"value": "10000.00",
			"currency": "IDR"
		},
		"cumulativePaymentAmount": null,
		"paidBills": "",
		"totalAmount": {
			"value": "10000.00",
			"currency": "IDR"
		},
		"trxDateTime": "2025-01-28T12:36:00+07:00",
		"referenceNo": "24657125601",
		"journalNum": "",
		"paymentType": "",
		"flagAdvise": "N",
		"subCompany": "00000",
		"billDetails": "",
		"freeTexts": "",
		"additionalInfo": {}
	}`
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

	ctx := context.Background()
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	s, err := bca_service.NewBCAService(
		request.NewBCAEgress(security, bCfg, cfg),
		request.NewBCAIngress(security),
		cfg,
		bCfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)
	if err != nil {
		t.Errorf("Error creating BCA Service: %v", err)
	}
	// Generate the mock signature
	// timeStamp := time.Now().Format(time.RFC3339)
	timeStamp := "2025-01-13T22:58:39+07:00"
	mockSignature, err := security.CreateAsymmetricSignature(ctx, timeStamp)
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
	mockRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.AppHost+bCfg.RequestedEndpoints.AuthURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Errorf("Error creating mock request: %v", err)
	}

	mockRequest.Header.Set("Content-Type", "application/json")
	mockRequest.Header.Set("X-TIMESTAMP", timeStamp)
	mockRequest.Header.Set("X-CLIENT-KEY", bCfg.BankRequestedCredentials.ClientID)
	// mockRequest.Header.Set("X-SIGNATURE", mockSignature)
	mockRequest.Header.Set("X-SIGNATURE", "TPlezaM75RMEr1L8Sr80v0NuPnb3BCRbtB9RbMqfcsquSYW8R0IZI375Ua/ej53vpQgMbPD55/fSXbN74r8NehCTd9s8MFnUzSLvaJ+to/udikB3AYa8aTU3VHu7UeChcYlyNFxLjnRM6d5DeIbBEP4NMi5t+9oSxhhwKWY+OpsGJSU7yqZ0bWNXPH7Di+GgYwXWJHdAVJfthI5X55SSs4z40+p1S6IQRI8/kAY+3R+Pucp9zucJkHmhX0MKOAFvhQH1sz812Fz04E5a26f1SdY++F4/1QzpSVd7dmnBknEaH/8OYETxeIJCEwZ5GDkfUnVpqvqKtJThvHZEWJ6yLw==")

	// Call the Generate Access Token function
	data, err := s.GenerateAccessToken(ctx, mockRequest)
	if err != nil {
		t.Errorf("Error generating access token: %v", err)
	}

	slog.Debug("timestamp", "data", timeStamp)
	slog.Debug("signature", "data", mockSignature)
	marshalled, _ := json.Marshal(data)
	slog.Debug("response", "data", marshalled)
}

func TestValidateAccessToken(t *testing.T) {
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	s, err := bca_service.NewBCAService(
		request.NewBCAEgress(security, bCfg, cfg),
		request.NewBCAIngress(security),
		cfg,
		bCfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)
	if err != nil {
		t.Errorf("Error creating BCA Service: %v", err)
	}

	result, err := s.Ingress.ValidateAccessToken(context.Background(), biStorage.GetRedisInstance(), "QyAuKj2Ph0dYkwZ-zozRTg85FC86nfd43qFPqj_dwAKnCIrKg1I4TxSxOeFiZt1F")
	if err != nil {
		t.Errorf("Error validating access token: %v", err)
	}

	slog.Debug("response", "data", fmt.Sprintf("%+v", result))
}

func TestCreateVA(t *testing.T) {
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	s, err := bca_service.NewBCAService(
		request.NewBCAEgress(security, bCfg, cfg),
		request.NewBCAIngress(security),
		cfg,
		bCfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)
	if err != nil {
		t.Errorf("Error creating BCA Service: %v", err)
	}

	// Generate CreateVA object
	data := biModels.CreateVAReq{
		NamaUser:         "Joko Dono",
		IdJenisPembelian: 1,
		JumlahPembayaran: 100000,
	}

	err = s.CreateVA(context.Background(), &data)
	if err != nil {
		t.Errorf("Error bill presentment: %v", err)
	}
}

func TestCreateVAV2(t *testing.T) {
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	s, err := bca_service.NewBCAService(
		request.NewBCAEgress(security, bCfg, cfg),
		request.NewBCAIngress(security),
		cfg,
		bCfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)
	if err != nil {
		t.Errorf("Error creating BCA Service: %v", err)
	}

	// Generate CreateVA object
	data := biModels.CreatePaymentVARequestV2{
		IDWallet:      6,
		IDTransaction: 100,
		IDBank:        3,
		AccountName:   "Joko Dono",
		TotalAmount:   "10000",
		IDService:     1,
		CustomerNo:    "1234567890",
	}

	err = s.CreateVAV2(context.Background(), &data)
	if err != nil {
		t.Errorf("Error bill presentment: %v", err)
	}
}

func TestGetWatchedTransaction(t *testing.T) {
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	s, err := bca_service.NewBCAService(
		request.NewBCAEgress(security, bCfg, cfg),
		request.NewBCAIngress(security),
		cfg,
		bCfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)
	if err != nil {
		t.Errorf("Error creating BCA Service: %v", err)
	}

	data := s.GetWatchedTransaction(context.Background())
	for _, watcher := range data {
		slog.Debug("watcher", "data", watcher)
	}

	time.Sleep(60 * time.Minute)
}

func TestTransferIntraBank(t *testing.T) {
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	s, err := bca_service.NewBCAService(
		request.NewBCAEgress(security, bCfg, cfg),
		request.NewBCAIngress(security),
		cfg,
		bCfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)
	if err != nil {
		t.Errorf("Error creating BCA Service: %v", err)
	}

	res, err := s.TransferIntraBank(context.Background(), &biModels.BCATransferIntraBankReq{
		PartnerReferenceNumber: uuid.New().String(),
		Amount: biModels.Amount{
			Value:    "10000.00",
			Currency: "IDR",
		},
		BeneficiaryAccountNo: "0611115813",
		SourceAccountNo:      "1234567890",
	})
	if err != nil {
		t.Errorf("Error transfer intra bank: %v", err)
	}

	slog.Debug("response", "data", res)
}

func TestTransferInterBank(t *testing.T) {
	security, err := setup()
	if err != nil {
		t.Fatalf("error setting up bca security instance: %v", err)
	}

	s, err := bca_service.NewBCAService(
		request.NewBCAEgress(security, bCfg, cfg),
		request.NewBCAIngress(security),
		cfg,
		bCfg,
		biStorage.GetDBConnection(),
		biStorage.GetRedisInstance(),
	)
	if err != nil {
		t.Errorf("Error creating BCA Service: %v", err)
	}

	res, err := s.TransferInterBank(context.Background(), &biModels.BCATransferInterBankRequest{
		PartnerReferenceNo: uuid.New().String(),
		Amount: biModels.Amount{
			Value:    "10000",
			Currency: "IDR",
		},
		BeneficiaryAccountNo:   "888801000157508",
		BeneficiaryAccountName: "Yories Yolanda",
		BeneficiaryBankCode:    "789",
	})
	if err != nil {
		t.Errorf("Error transfer intra bank: %v", err)
	}

	slog.Debug("response", "data", res)
}
