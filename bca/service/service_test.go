package bca_service

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	request "github.com/voxtmault/bank-integration/bca/request"
	security "github.com/voxtmault/bank-integration/bca/security"
	"github.com/voxtmault/bank-integration/config"
	"github.com/voxtmault/bank-integration/models"
	"github.com/voxtmault/bank-integration/storage"
	"github.com/voxtmault/bank-integration/utils"
)

func TestGetAccessToken(t *testing.T) {
	cfg := config.New("/home/andy/go-projects/github.com/voxtmault/bank-integration/.env")
	utils.InitValidator()
	storage.InitMariaDB(&cfg.MariaConfig)
	storage.InitRedis(&cfg.RedisConfig)

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
		storage.GetRedisInstance(),
		utils.GetValidator(),
	)

	if err := service.GetAccessToken(context.Background()); err != nil {
		t.Errorf("Error getting access token: %v", err)
	}
}

func TestBalanceInquiry(t *testing.T) {
	cfg := config.New("/home/andy/go-projects/github.com/voxtmault/bank-integration/.env")
	utils.InitValidator()
	storage.InitMariaDB(&cfg.MariaConfig)
	storage.InitRedis(&cfg.RedisConfig)

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
		storage.GetRedisInstance(),
		utils.GetValidator(),
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

func TestGenerateAccessToken(t *testing.T) {
	// Logic
	// 1. Load the configs
	// 2. Initiate required services (DB, Validations, etc...)
	// 3. Initiate the BCA Service
	// 4. Generate a mock signature using the BCA Requested Client ID and Secret
	// 5. Generate a mock http request using the data received from (4)
	// 6. Call the Generate Access Token function

	cfg := config.New("/home/andy/go-projects/github.com/voxtmault/bank-integration/.env")
	utils.InitValidator()
	storage.InitMariaDB(&cfg.MariaConfig)
	storage.InitRedis(&cfg.RedisConfig)

	// Load Registered Banks

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
		storage.GetRedisInstance(),
		utils.GetValidator(),
	)

	mockSecurity := security.NewBCASecurity(
		&cfg.BCAConfig, &cfg.Keys,
	)
	mockSecurity.ClientID = "c3e7fe0d-379c-4ce2-ad85-372fea661aa0"
	mockSecurity.ClientSecret = "3fd9d63c-f4f1-4c26-8886-fecca45b1053"

	// Generate the mock signature
	timeStamp := time.Now().Format(time.RFC3339)
	mockSignature, err := mockSecurity.CreateAsymmetricSignature(context.Background(), timeStamp)
	if err != nil {
		t.Errorf("Error generating mock signature: %v", err)
	}

	// Generate the mock http request
	body := models.GrantType{
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
	mockRequest.Header.Set("X-SIGNATURE", mockSignature)

	// Call the Generate Access Token function
	data, err := service.GenerateAccessToken(context.Background(), mockRequest)
	if err != nil {
		t.Errorf("Error generating access token: %v", err)
	}

	slog.Debug("response", "data", data)

}

func TestValidateAccessToken(t *testing.T) {
	cfg := config.New("/home/andy/go-projects/github.com/voxtmault/bank-integration/.env")
	utils.InitValidator()
	storage.InitMariaDB(&cfg.MariaConfig)
	storage.InitRedis(&cfg.RedisConfig)

	// Load Registered Banks

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
		storage.GetRedisInstance(),
		utils.GetValidator(),
	)

	result, err := service.ValidateAccessToken(context.Background(), "LHbfo2wp6iPrVwvvcZgKhkeXmyEOKA1rMBY2-VfNfmDUQ0fsvZTxI8Aeh0RmYLNc")
	if err != nil {
		t.Errorf("Error validating access token: %v", err)
	}

	slog.Debug("response", "data", result)
}
