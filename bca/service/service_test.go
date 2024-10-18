package bca_service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"testing"

	request "github.com/voxtmault/bank-integration/bca/request"
	security "github.com/voxtmault/bank-integration/bca/security"
	"github.com/voxtmault/bank-integration/config"
	"github.com/voxtmault/bank-integration/models"
	"github.com/voxtmault/bank-integration/storage"
	"github.com/voxtmault/bank-integration/utils"
)

func TestGetAccessToken(t *testing.T) {
	cfg := config.New("../../.env")
	utils.InitValidator()
	storage.InitMariaDB(&cfg.MariaConfig)

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	service := NewBCAService(
		request.NewBCARequest(
			security.NewBCASecurity(
				&cfg.BCAConfig,
				&cfg.Keys,
			),
			utils.GetValidator(),
		),
		cfg,
		storage.GetDBConnection(),
	)

	if err := service.GetAccessToken(context.Background()); err != nil {
		t.Errorf("Error getting access token: %v", err)
	}
}

func TestBalanceInquiry(t *testing.T) {
	cfg := config.New("/home/andy/go-projects/github.com/voxtmault/bank-integration/.env")
	storage.InitMariaDB(&cfg.MariaConfig)
	service := NewBCAService(
		request.NewBCARequest(
			security.NewBCASecurity(
				&cfg.BCAConfig,
				&cfg.Keys,
			),
			utils.GetValidator(),
		),
		cfg,
		storage.GetDBConnection(),
	)

	data, err := service.BalanceInquiry(context.Background(), &models.BCABalanceInquiry{})
	if err != nil {
		t.Errorf("Error getting access token: %v", err)
	}

	log.Println(data)
}

func TestBalanceInquiryUnmarshall(t *testing.T) {

	// sample := `
	// {"responseCode":"4001102","responseMessage":"Invalid Mandatory Field accountNo","referenceNo":"","partnerReferenceNo":"2020102900000000000001","accountNo":"1234567890","name":"","accountInfos":{"balanceType":"","amount":{"value":"","currency":""},"floatAmount":{"value":"","currency":""},"holdAmount":{"value":"","currency":""},"availableBalance":{"value":"","currency":""},"ledgerBalance":{"value":"","currency":""},"currentMultilateralLimit":{"value":"","currency":""},"registrationStatusCode":"","status":""}}
	// `
	// sample := `
	// {"responseCode":"5001100","responseMessage":"General error","referenceNo":"","partnerReferenceNo":"2020102900000000000001","accountNo":"1234567890","name":"","accountInfos":{"balanceType":"","amount":{"value":"","currency":""},"floatAmount":{"value":"","currency":""},"holdAmount":{"value":"","currency":""},"availableBalance":{"value":"","currency":""},"ledgerBalance":{"value":"","currency":""},"currentMultilateralLimit":{"value":"","currency":""},"registrationStatusCode":"","status":""}}
	// `
	sample := `
	{"responseCode":"2001100","responseMessage":"Successful","referenceNo":"2020102977770000000009","partnerReferenceNo":"2020102900000000000001","accountNo":"1234567890","name":"ANDHIKA","accountInfos":{"balanceType":"Cash","amount":{"value":"100000.00","currency":"IDR"},"floatAmount":{"value":"500000.00","currency":"IDR"},"holdAmount":{"value":"200000.00","currency":"IDR"},"availableBalance":{"value":"200000.00","currency":"IDR"},"ledgerBalance":{"value":"200000.00","currency":"IDR"},"currentMultilateralLimit":{"value":"200000.00","currency":"IDR"},"registrationStatusCode":"0001","status":"0001"}}
	`

	var obj models.BCAAccountBalance
	if err := json.Unmarshal([]byte(sample), &obj); err != nil {
		t.Errorf("Error unmarshalling: %v", err)
	}

	log.Printf("%+v\n", obj)
}

func TestBillPresement(t *testing.T) {
	cfg := config.New("../../.env")
	utils.InitValidator()
	storage.InitMariaDB(&cfg.MariaConfig)

	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	s := BCAService{DB: storage.GetDBConnection()}
	bodyReq := `{
 "partnerServiceId": " 11223",
 "customerNo": "1234567890123456",
 "virtualAccountNo": " 112231234567890123456",
 "inquiryRequestId": "202410180000000000001"
}`
	var obj models.BCAVARequestPayload
	if err := json.Unmarshal([]byte(bodyReq), &obj); err != nil {
		t.Errorf("Error unmarshalling: %v", err)
	}
	log.Printf("%+v\n", obj)
	res, err := s.BillPresentment(context.Background(), obj)
	if err != nil {
		t.Errorf("Error From cuntion Bill: %v", err)
	}
	result, err := json.Marshal(res)
	if err != nil {
		t.Errorf("Error From Marshal: %v", err)
	}
	fmt.Println(string(result))
}

func TestInquiryVA(t *testing.T) {
	cfg := config.New("../../.env")
	utils.InitValidator()
	storage.InitMariaDB(&cfg.MariaConfig)
	if strings.Contains(strings.ToLower(cfg.Mode), "debug") {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	s := BCAService{DB: storage.GetDBConnection()}
	bodyReq := `{
 "partnerServiceId": " 11223",
 "customerNo": "1234567890123456",
 "virtualAccountNo": " 112231234567890123456",
 "virtualAccountName": "Test Va",
 "paymentRequestId": "202410180000000000001",
 "paidAmount": {
 "value": "100000.00",
 "currency": "IDR"
 },
 "totalAmount": {
 "value": "100000.00",
 "currency": "IDR"
 }
}`
	var obj models.BCAInquiryRequest
	if err := json.Unmarshal([]byte(bodyReq), &obj); err != nil {
		t.Errorf("Error unmarshalling: %v", err)
	}
	log.Printf("%+v\n", obj)
	res, err := s.InquiryVA(context.Background(), obj)
	if err != nil {
		t.Errorf("Error From cuntion Bill: %v", err)
	}
	result, err := json.Marshal(res)
	if err != nil {
		t.Errorf("Error From Marshal: %v", err)
	}
	fmt.Println(string(result))
}
