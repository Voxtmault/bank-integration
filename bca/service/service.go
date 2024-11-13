package bca_service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/rotisserie/eris"
	"github.com/voxtmault/bank-integration/bca"
	biConfig "github.com/voxtmault/bank-integration/config"
	biInterfaces "github.com/voxtmault/bank-integration/interfaces"
	biModels "github.com/voxtmault/bank-integration/models"
	biStorage "github.com/voxtmault/bank-integration/storage"

	// timerexpired "github.com/voxtmault/bank-integration/timer_expired"
	biUtil "github.com/voxtmault/bank-integration/utils"
	watcher "github.com/voxtmault/bank-integration/watcher"
)

type BCAService struct {

	// Dependency Injection
	Egress          biInterfaces.RequestEgress
	Ingress         biInterfaces.RequestIngress
	GeneralSecurity biUtil.GeneralSecurity

	// Adding Watcher here instead of using a centralized Watcher is an intentional design choice
	Watcher *watcher.TransactionWatcher

	// Configs
	Config *biConfig.BankingConfig

	// Runtime Access Tokens
	AccessToken          string
	AccessTokenExpiresAt int64

	// DB Connections
	DB  *sql.DB
	RDB *biStorage.RedisInstance
}

var _ biInterfaces.SNAP = &BCAService{}

func NewBCAService(egress biInterfaces.RequestEgress, ingress biInterfaces.RequestIngress, config *biConfig.BankingConfig, db *sql.DB, rdb *biStorage.RedisInstance) (*BCAService, error) {

	service := BCAService{
		Egress:  egress,
		Ingress: ingress,
		Config:  config,
		DB:      db,
		RDB:     rdb,
		Watcher: watcher.NewTransactionWatcher(),
	}
	// Get current loaded BCAService internal bank id and bank name
	if err := service.getInternalBankInfo(); err != nil {
		slog.Error("error getting internal bank id", "error", err)
		return nil, err
	}

	// Get VA created by the loaded bank id that is still waiting for payment and add it to the watcher
	if err := service.GetAllVAWaitingPayment(context.Background()); err != nil {
		slog.Error("error getting all va waiting payment", "error", err)
		return nil, err
	}

	return &service, nil
}

// Egress
func (s *BCAService) GetAccessToken(ctx context.Context) error {

	// Logic
	// 1. Customize the header of the request (including creating the signature)
	// 2. Send the request
	// 3. Parse the response

	baseUrl := s.Config.BCAConfig.BaseURL + s.Config.BCAURLEndpoints.AccessTokenURL
	method := http.MethodPost
	body := biModels.GrantType{
		GrantType: "client_credentials",
	}

	slog.Debug("Marshalling body")
	jsonBody, err := json.Marshal(body)
	if err != nil {
		slog.Debug("error marshalling body", "error", err)
		return eris.Wrap(err, "marshalling body")
	}

	req, err := http.NewRequestWithContext(ctx, method, baseUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return eris.Wrap(err, "creating request")
	}

	slog.Debug("Building Request Header")
	// Before sending the request, customize the header
	if err = s.Egress.GenerateAccessRequestHeader(ctx, req, s.Config); err != nil {
		slog.Debug("error generating access token request header", "error", err)
		return eris.Wrap(err, "access token request header")
	}

	slog.Debug("Sending Request")
	response, err := s.RequestHandler(ctx, req)
	if err != nil {
		slog.Debug("error sending request", "error", err)
		if response != "" {
			return eris.Wrap(eris.New(response), "sending request")
		} else {
			return eris.Wrap(err, "sending request")
		}
	}
	slog.Debug("Response from BCA", "Response: ", response)

	var atObj biModels.AccessTokenResponse
	if err = json.Unmarshal([]byte(response), &atObj); err != nil {
		slog.Debug("error unmarshalling response", "error", err)
		return eris.Wrap(err, "unmarshalling response")
	}

	s.AccessToken = atObj.AccessToken

	// Create internal counter for when the access token expires
	s.AccessTokenExpiresAt = time.Now().Add(time.Second * 900).Unix()

	return nil
}

func (s *BCAService) BalanceInquiry(ctx context.Context, payload *biModels.BCABalanceInquiry) (*biModels.BCAAccountBalance, error) {

	// Checks if the access token is empty, if yes then get a new one
	if err := s.CheckAccessToken(ctx); err != nil {
		return nil, eris.Wrap(err, "checking access token")
	}

	baseUrl := s.Config.BCAConfig.BaseURL + s.Config.BCAURLEndpoints.BalanceInquiryURL
	method := http.MethodPost
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, eris.Wrap(err, "marshalling payload")
	}

	request, err := http.NewRequestWithContext(ctx, method, baseUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, eris.Wrap(err, "creating request")
	}

	if err = s.Egress.GenerateGeneralRequestHeader(ctx, request, s.Config, payload, s.Config.BCAURLEndpoints.BalanceInquiryURL, s.AccessToken); err != nil {
		return nil, eris.Wrap(err, "constructing request header")
	}

	response, err := s.RequestHandler(ctx, request)
	if err != nil {
		if response != "" {
			return nil, eris.Wrap(eris.New(response), "sending request")
		} else {
			return nil, eris.Wrap(err, "sending request")
		}
	}

	var obj biModels.BCAAccountBalance
	if err = json.Unmarshal([]byte(response), &obj); err != nil {
		return nil, eris.Wrap(err, "unmarshalling balance inquiry response")
	}

	// Checks for erronous response
	if obj.ResponseCode != "2001100" {
		return nil, eris.New(obj.ResponseMessage)
	}

	return &obj, nil
}

func (s *BCAService) CreateVA(ctx context.Context, payload *biModels.CreateVAReq) error {
	partnerId := "   " + s.Config.BCAPartnerInformation.BCAPartnerId
	query := `
	INSERT INTO va_request (partnerServiceId, customerNo, virtualAccountNo, totalAmountValue, 
				   			virtualAccountName, id_user, owner_table,expired_date, id_bank)
	VALUES(?,?,?,?,?,?,?,?,?)
	`
	var id int64
	expiredTime := time.Now().Add(time.Hour * time.Duration(s.Config.BCAConfig.BCAVAExpireTime))
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Debug("error beginning transaction", "error", err)
		tx.Rollback()
		return eris.Wrap(err, "beginning transaction")
	}

	numVA, customerNo := s.BuildNumVA(payload.IdUser, payload.IdJenisUser, partnerId)
	cekpaid, err := s.CheckVAPaid(ctx, numVA, tx)
	if err != nil {
		return eris.Wrap(err, "querying va_table")
	}
	if cekpaid {
		// Meaning all previous VA with the same VA Number has been paid or cancelled
		result, err := tx.ExecContext(ctx, query, partnerId, customerNo, numVA, payload.JumlahPembayaran,
			payload.NamaUser, payload.IdUser, payload.IdJenisUser, expiredTime.Format(time.DateTime),
			s.Config.BCAConfig.InternalBankID)
		if err != nil {
			tx.Rollback()
			return eris.Wrap(err, "querying va_table")
		}
		id, _ = result.LastInsertId()
	} else {
		// Meaning there is still a VA with the same VA Number that is still waiting for payment
		tx.Rollback()
		return eris.Wrap(err, "Va Not Paid")
	}

	if err = tx.Commit(); err != nil {
		slog.Debug("error committing transaction", "error", err)
		tx.Rollback()
		return eris.Wrap(err, "committing transaction")
	}

	// Create Transaction Watcher after successfull transaction commit
	watchedTransaction := watcher.NewWatcher()
	watchedTransaction.IDTransaction = uint(id)
	watchedTransaction.ExpireAt = expiredTime.Local()

	s.Watcher.AddWatcher(watchedTransaction)

	// timerexpired.SetTimer(biModels.TimerPayment{Id: int(id), NumVA: numVA, IdBank: 0, ExpiredAt: expiredTime})
	return nil
}

// ChecksAccessToken is an exclusive function to renew the access token if it is expired or if it's empty.
func (s *BCAService) CheckAccessToken(ctx context.Context) error {
	if s.AccessToken == "" {
		// Access token is empty, get a new one
		slog.Debug("Access Token is empty, getting a new one")
		if err := s.GetAccessToken(ctx); err != nil {
			return eris.Wrap(err, "getting access token")
		}
	} else if time.Now().Unix() > s.AccessTokenExpiresAt {
		// Access token is expired, get a new one
		if err := s.GetAccessToken(ctx); err != nil {
			return eris.Wrap(err, "renewing access token")
		}
	}

	return nil
}

// Ingress
func (s *BCAService) GenerateAccessToken(ctx context.Context, request *http.Request) (*biModels.AccessTokenResponse, error) {
	// Logic
	// 1. Parse the request body
	// 2. Parse the request header
	// 3. Validate body and header
	// 4. Retrieve the client secret from redis
	// 5. Verify Asymmetric Signature
	// 6. Generate Access Token
	// 7. Save the Access Token along with client secret to redis
	// 8. Return to caller

	// Parse the request body
	var body biModels.GrantType
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		return &biModels.AccessTokenResponse{
			BCAResponse: &bca.BCAAuthGeneralError,
		}, eris.Wrap(err, "decoding request body")
	}

	// Validate the received struct
	if err := biUtil.ValidateStruct(ctx, body); err != nil {
		slog.Debug("error validating request body", "error", err)

		return &biModels.AccessTokenResponse{
			BCAResponse: &bca.BCAAuthInvalidFieldFormatClient,
		}, nil
	}

	// Verify Asymmetric Signature
	result, response, clientSecret := s.Ingress.VerifyAsymmetricSignature(ctx, request, s.RDB)
	if response != nil {
		return &biModels.AccessTokenResponse{
			BCAResponse: response,
		}, nil
	}

	if !result {
		return &biModels.AccessTokenResponse{
			BCAResponse: &bca.BCAAuthUnauthorizedSignature,
		}, nil
	}

	slog.Debug("received client secret", "clientSecret", clientSecret)

	// Generate the access token
	token, err := s.GeneralSecurity.GenerateAccessToken(ctx)
	if err != nil {
		slog.Debug("error generating access token", "error", err)
		return nil, eris.Wrap(err, "generating access token")
	}
	slog.Debug("generated token", "token", token)

	// Save the access token to redis along with the configured client secret & expiration time
	key := fmt.Sprintf("%s:%s", biUtil.AccessTokenRedis, token)
	if err := s.RDB.RDB.Set(ctx, key, clientSecret, time.Second*time.Duration(s.Config.BCARequestedClientCredentials.AccessTokenExpirationTime)).Err(); err != nil {
		return &biModels.AccessTokenResponse{
			BCAResponse: &bca.BCAAuthGeneralError,
		}, eris.Wrap(err, "saving access token to redis")
	}

	bcaResponse := bca.BCAAuthResponseSuccess
	bcaResponse.ResponseMessage = "Successful"

	return &biModels.AccessTokenResponse{
		AccessToken: token,
		TokenType:   "bearer",
		ExpiresIn:   strconv.Itoa(int(s.Config.BCARequestedClientCredentials.AccessTokenExpirationTime)),
		BCAResponse: &bcaResponse,
	}, nil
}

func (s *BCAService) BillPresentment(ctx context.Context, request *http.Request) (*biModels.VAResponsePayload, error) {
	var response biModels.VAResponsePayload

	// Validate Channel ID
	if request.Header.Get("CHANNEL-ID") == "" || request.Header.Get("CHANNEL-ID") != s.Config.BCAConfig.ChannelID {

		response.BCAResponse = bca.BCABillInquiryResponseUnauthorizedUnknownClient

		if request.Header.Get("CHANNEL-ID") == "" {
			response.BCAResponse = bca.BCABillInquiryResponseMissingMandatoryField
			response.BCAResponse.ResponseMessage = "Invalid Mandatory Field {CHANNEL-ID}"
		}

		return &response, nil
	}

	// Validate Partner ID
	if request.Header.Get("X-PARTNER-ID") == "" || request.Header.Get("X-PARTNER-ID") != s.Config.BCAPartnerInformation.BCAPartnerId {

		response.BCAResponse = bca.BCABillInquiryResponseUnauthorizedUnknownClient

		if request.Header.Get("X-PARTNER-ID") == "" {
			response.BCAResponse = bca.BCABillInquiryResponseMissingMandatoryField
			response.BCAResponse.ResponseMessage = "Invalid Mandatory Field {X-PARTNER-ID}"
		}

		return &response, nil
	}

	// Validate External ID
	if request.Header.Get("X-EXTERNAL-ID") == "" {
		slog.Debug("externalId is empty")

		response.BCAResponse = bca.BCABillInquiryResponseMissingMandatoryField
		response.ResponseMessage = response.ResponseMessage + " [X-EXTERNAL-ID]"

		return &response, nil
	}

	// Parse the Request Body
	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		slog.Debug("error reading request body", "error", err)

		response.BCAResponse = bca.BCABillInquiryResponseRequestParseError

		return &response, nil
	}
	defer request.Body.Close()

	var payload biModels.BCAVARequestPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		slog.Debug("error un-marshaling request body", "error", err)

		response.BCAResponse = bca.BCABillInquiryResponseRequestParseError

		return &response, nil
	}

	// Validate the received payload
	if err := biUtil.ValidateStruct(ctx, payload); err != nil {
		for _, item := range err.(validator.ValidationErrors) {
			slog.Warn("error validating struct field", "field", item.Field(), "tag", item.Tag())
			if item.Tag() == "required" || item.Tag() == "min" {
				response.BCAResponse = bca.BCABillInquiryResponseMissingMandatoryField
				response.BCAResponse.ResponseMessage = "Invalid Mandatory Field {" + item.Field() + "}"

				return &response, nil
			} else {
				response.BCAResponse = bca.BCABillInquiryResponseInvalidFieldFormat
				response.BCAResponse.ResponseMessage = "Invalid Field Format {" + item.Field() + "}"

				return &response, nil
			}
		}
	}

	// Set the default value of response
	response.VirtualAccountData = biModels.VABCAResponseData{}.Default()

	// Validate Auth related header, this function will also be validating for duplicate X-EXTERNAL-ID
	result, authResponse := s.Ingress.VerifySymmetricSignature(ctx, request, s.RDB, bodyBytes)
	if authResponse != nil {
		slog.Error("verifying symmetric signature failed", "response", authResponse)

		response.BCAResponse = *authResponse
		response.BCAResponse.ResponseCode = (response.BCAResponse.ResponseCode)[:3] + "24" + (response.BCAResponse.ResponseCode)[5:]
		if response.BCAResponse.HTTPStatusCode == http.StatusInternalServerError {
			response.VirtualAccountData = biModels.VABCAResponseData{}.Default()
		} else {
			response.VirtualAccountData = nil
		}

		return &response, nil
	}

	if !result {
		response.BCAResponse = bca.BCABillInquiryResponseUnauthorizedSignature
		response.VirtualAccountData = nil

		return &response, nil
	}

	// Validate unique external id if the request is consistent
	externalUnique, err := s.Ingress.ValidateUniqueExternalID(ctx, s.RDB, request.Header.Get("X-EXTERNAL-ID"))
	if err != nil {
		slog.Debug("error validating externalId", "error", err)

		if eris.Cause(err).Error() == "invalid field format" {
			response.BCAResponse = bca.BCABillInquiryResponseInvalidFieldFormat
			response.ResponseMessage = "Invalid Field Format [X-EXTERNAL-ID]"
		} else {
			response.BCAResponse = bca.BCABillInquiryResponseGeneralError
		}

		return &response, nil
	}

	// Populate the default value
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

	// Call the core function
	if err = s.BillPresentmentCore(ctx, &response, &payload); err != nil {
		slog.Error("error in BillPresentmentCore", "error", err)

		return &response, nil
	}

	if !externalUnique {
		slog.Debug("externalId is not unique")

		response.BCAResponse = bca.BCABillInquiryResponseDuplicateExternalID
		response.VirtualAccountData.InquiryReason.English = "Cannot use the same X-EXTERNAL-ID"
		response.VirtualAccountData.InquiryReason.Indonesia = "Tidak bisa menggunakan X-EXTERNAL-ID yang sama"

		// For conflicting external id, set the inquiry status to failure
		response.VirtualAccountData.InquiryStatus = "01"
	}

	return &response, nil
}
func (s *BCAService) BillPresentmentCore(ctx context.Context, response *biModels.VAResponsePayload, payload *biModels.BCAVARequestPayload) error {

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Debug("error beginning transaction", "error", err)
		tx.Rollback()

		response.BCAResponse = bca.BCABillInquiryResponseGeneralError
		response.VirtualAccountData = biModels.VABCAResponseData{}.Default()

		return eris.Wrap(err, "beginning transaction")
	}
	paidAmount := biModels.Amount{}
	statement := `
	SELECT partnerServiceId, customerNo, virtualAccountNo, virtualAccountName, totalAmountValue, totalAmountCurrency, paidAmountValue, paidAmountCurrency,COALESCE(expired_date, CURRENT_TIMESTAMP()) AS effective_expired_date
	FROM va_request 
	WHERE TRIM(virtualAccountNo) = ?
	ORDER BY created_at DESC
	LIMIT 1
	`
	expDate := ""
	err = tx.QueryRowContext(ctx, statement, strings.ReplaceAll(payload.VirtualAccountNo, " ", "")).Scan(
		&response.VirtualAccountData.PartnerServiceID,
		&response.VirtualAccountData.CustomerNo,
		&response.VirtualAccountData.VirtualAccountNo,
		&response.VirtualAccountData.VirtualAccountName,
		&response.VirtualAccountData.TotalAmount.Value,
		&response.VirtualAccountData.TotalAmount.Currency,
		&paidAmount.Value,
		&paidAmount.Currency, &expDate)
	if err == sql.ErrNoRows {
		slog.Debug("bill presentment core", "error", "va not found")
		tx.Rollback()

		response.BCAResponse = bca.BCABillInquiryResponseVANotFound
		response.VirtualAccountData.InquiryReason.English = "Bill Not Found"
		response.VirtualAccountData.InquiryReason.Indonesia = "Tagihan tidak ditemukan"
		response.VirtualAccountData.InquiryStatus = "01"

		return nil
	} else if err != nil {
		slog.Debug("error querying va_request", "error", err)
		tx.Rollback()

		response.BCAResponse = bca.BCABillInquiryResponseGeneralError
		response.VirtualAccountData = biModels.VABCAResponseData{}.Default()

		return nil
	}
	if paidAmount.Value != "0.00" && paidAmount.Value != "" {
		slog.Debug("va has been paid")
		tx.Rollback()

		response.BCAResponse = bca.BCABillInquiryResponseVAPaid
		response.VirtualAccountData.InquiryReason.English = "Paid Bill"
		response.VirtualAccountData.InquiryReason.Indonesia = "Tagihan Telah Terbayar"
		response.VirtualAccountData.InquiryStatus = "01"
		return nil
	}
	nExpDate, _ := time.Parse(time.DateTime, expDate)
	if time.Now().After(nExpDate) {
		slog.Debug("va has been Expired")
		tx.Rollback()
		response.BCAResponse = bca.BCABillInquiryResponseVAExpired
		response.VirtualAccountData.InquiryReason.English = "Bill Expired"
		response.VirtualAccountData.InquiryReason.Indonesia = "Tagihan sudah kadarluasa"
		response.VirtualAccountData.InquiryStatus = "01"

		return nil
	}
	statement = `
	UPDATE va_request SET inquiryRequestId = ? 
	WHERE virtualAccountNo = ? AND paidAmountValue = '0.00'
	`
	_, err = tx.QueryContext(ctx, statement, payload.InquiryRequestID, payload.VirtualAccountNo)
	if err != nil {
		slog.Debug("error updating va_request", "error", err)
		tx.Rollback()

		response.BCAResponse = bca.BCABillInquiryResponseGeneralError
		response.VirtualAccountData = biModels.VABCAResponseData{}.Default()

		return eris.Wrap(err, "updating va_request")
	}

	response.BCAResponse = bca.BCABillInquiryResponseSuccess
	response.VirtualAccountData.InquiryStatus = "00"
	response.VirtualAccountData.InquiryReason.Indonesia = "Sukses"
	response.VirtualAccountData.InquiryReason.English = "Success"

	if err = tx.Commit(); err != nil {
		slog.Debug("bill presentment", "error committing transaction", err)
		tx.Rollback()

		response.BCAResponse = bca.BCABillInquiryResponseGeneralError
		response.VirtualAccountData = biModels.VABCAResponseData{}.Default()

		return eris.Wrap(err, "committing transaction")
	}

	return nil
}

func (s *BCAService) InquiryVA(ctx context.Context, request *http.Request) (*biModels.BCAInquiryVAResponse, error) {

	var response biModels.BCAInquiryVAResponse

	// Validate Channel ID
	if request.Header.Get("CHANNEL-ID") == "" || request.Header.Get("CHANNEL-ID") != s.Config.BCAConfig.ChannelID {

		response.BCAResponse = bca.BCAPaymentFlagResponseUnauthorizedUnknownClient

		if request.Header.Get("CHANNEL-ID") == "" {
			response.BCAResponse = bca.BCAPaymentFlagResponseMissingMandatoryField
			response.BCAResponse.ResponseMessage = "Invalid Mandatory Field {CHANNEL-ID}"
		}

		return &response, nil
	}

	// Validate Partner ID
	if request.Header.Get("X-PARTNER-ID") == "" || request.Header.Get("X-PARTNER-ID") != s.Config.BCAPartnerInformation.BCAPartnerId {

		response.BCAResponse = bca.BCAPaymentFlagResponseUnauthorizedUnknownClient

		if request.Header.Get("X-PARTNER-ID") == "" {
			response.BCAResponse = bca.BCAPaymentFlagResponseMissingMandatoryField
			response.BCAResponse.ResponseMessage = "Invalid Mandatory Field {X-PARTNER-ID}"
		}

		return &response, nil
	}

	// Validate External ID
	if request.Header.Get("X-EXTERNAL-ID") == "" {
		slog.Debug("externalId is empty")

		response.BCAResponse = bca.BCAPaymentFlagResponseMissingMandatoryField
		response.ResponseMessage = response.ResponseMessage + " [X-EXTERNAL-ID]"

		return &response, nil
	}

	// Parse the Request Body
	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		slog.Debug("error reading request body", "error", err)

		response.BCAResponse = bca.BCAPaymentFlagResponseRequestParseError

		return &response, nil
	}
	defer request.Body.Close()

	var payload biModels.BCAInquiryRequest
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		slog.Debug("error un-marshaling request body", "error", err)

		response.BCAResponse = bca.BCAPaymentFlagResponseRequestParseError

		return &response, nil
	}

	// Validate the received payload
	if err := biUtil.ValidateStruct(ctx, payload); err != nil {
		for _, item := range err.(validator.ValidationErrors) {
			slog.Warn("error validating struct field", "field", item.Field(), "tag", item.Tag())
			if item.Tag() == "required" || item.Tag() == "min" {
				response.BCAResponse = bca.BCAPaymentFlagResponseMissingMandatoryField
				response.BCAResponse.ResponseMessage = "Invalid Mandatory Field {" + item.Field() + "}"

				return &response, nil
			} else {
				response.BCAResponse = bca.BCAPaymentFlagResponseInvalidFieldFormat
				response.BCAResponse.ResponseMessage = "Invalid Field Format {" + item.Field() + "}"

				return &response, nil
			}
		}
	}

	// Validate Auth related header
	result, authResponse := s.Ingress.VerifySymmetricSignature(ctx, request, s.RDB, bodyBytes)
	if authResponse != nil {
		slog.Error("verifying symmetric signature failed", "response", authResponse)

		response.BCAResponse = *authResponse
		response.BCAResponse.ResponseCode = (response.BCAResponse.ResponseCode)[:3] + "25" + (response.BCAResponse.ResponseCode)[5:]
		if response.BCAResponse.HTTPStatusCode == http.StatusInternalServerError {
			response.VirtualAccountData = biModels.VirtualAccountDataInquiry{}.Default()
			response.VirtualAccountData.BillDetails = payload.BillDetails
			response.VirtualAccountData.FreeTexts = payload.FreeTexts
			response.AdditionalInfo = payload.AdditionalInfo
		} else {
			response.VirtualAccountData = nil
		}

		return &response, nil
	}

	if !result {
		response.BCAResponse = bca.BCAPaymentFlagResponseUnauthorizedSignature
		response.VirtualAccountData = nil

		return &response, nil
	}

	// Set the default value of response
	response.VirtualAccountData = biModels.VirtualAccountDataInquiry{}.Default()
	response.VirtualAccountData.BillDetails = payload.BillDetails
	response.VirtualAccountData.FreeTexts = payload.FreeTexts
	response.AdditionalInfo = payload.AdditionalInfo

	// Validate X-EXTERNAL-ID and paymentRequestID is not already stored in redis
	key := request.Header.Get("X-EXTERNAL-ID") + ":" + payload.PaymentRequestID
	val, err := s.RDB.RDB.Get(ctx, key).Result()
	if err == nil && len(val) > 0 {
		// Meaning that a system error has occurred at BCA side causing double flagging request with the same X-EXTERNAL-ID and paymentRequestId
		uncompressedResponse, err := biUtil.DecompressData([]byte(val))
		if err != nil {
			slog.Error("error decompressing response", "error", err)
			response.BCAResponse = bca.BCAPaymentFlagResponseGeneralError
			response.VirtualAccountData = biModels.VirtualAccountDataInquiry{}.Default()
			response.AdditionalInfo = payload.AdditionalInfo

			return &response, nil
		}

		var storedResponse biModels.BCAInquiryVAResponse
		if err := json.Unmarshal(uncompressedResponse, &storedResponse); err != nil {
			slog.Error("error unmarshalling stored response", "error", err)

			response.BCAResponse = bca.BCAPaymentFlagResponseGeneralError
			response.VirtualAccountData = biModels.VirtualAccountDataInquiry{}.Default()
			response.AdditionalInfo = payload.AdditionalInfo

			return &response, nil
		}

		response.VirtualAccountData = storedResponse.VirtualAccountData
		response.BCAResponse = bca.BCAPaymentFlagResponseDuplicateExternalIDAndPaymentRequestID

		return &response, nil
	} else if err != nil && err != redis.Nil {
		slog.Error("error checking X-EXTERNAL-ID and paymentRequestID in redis", "error", err)
		response.BCAResponse = bca.BCAPaymentFlagResponseGeneralError
		response.VirtualAccountData = biModels.VirtualAccountDataInquiry{}.Default()
		response.AdditionalInfo = payload.AdditionalInfo

		return &response, nil
	}

	// Validate unique external id if the request is consistent
	externalUnique, err := s.Ingress.ValidateUniqueExternalID(ctx, s.RDB, request.Header.Get("X-EXTERNAL-ID"))
	if err != nil {
		slog.Debug("error validating externalId", "error", err)

		if eris.Cause(err).Error() == "invalid field format" {
			response.BCAResponse = bca.BCAPaymentFlagResponseInvalidFieldFormat
			response.ResponseMessage = "Invalid Field Format [X-EXTERNAL-ID]"
		} else {
			response.BCAResponse = bca.BCAPaymentFlagResponseGeneralError
		}

		return &response, nil
	}

	// Populate the default value
	response.VirtualAccountData = &biModels.VirtualAccountDataInquiry{
		BillDetails:        payload.BillDetails,
		FreeTexts:          payload.FreeTexts,
		PaymentRequestID:   payload.PaymentRequestID,
		ReferenceNo:        payload.ReferenceNo,
		CustomerNo:         payload.CustomerNo,
		VirtualAccountNo:   payload.VirtualAccountNo,
		VirtualAccountName: payload.VirtualAccountName,
		TrxDateTime:        payload.TrxDateTime,
		PartnerServiceID:   payload.PartnerServiceID,
		PaidAmount:         biModels.Amount{},
		TotalAmount:        biModels.Amount{},
		PaymentFlagReason:  biModels.Reason{},
		FlagAdvise:         "N",
	}
	response.AdditionalInfo = payload.AdditionalInfo

	// Call the core function
	if err = s.InquiryVACore(ctx, &response, &payload); err != nil {
		slog.Error("error in InquiryVACore", "error", eris.Cause(err))
	}

	// Save the response to redis along with the key consisting of X-EXTERNAL-ID and paymentRequestId
	responseBytes, err := json.Marshal(response)
	if err != nil {
		slog.Error("error marshalling response", "error", err)
		response.BCAResponse = bca.BCAPaymentFlagResponseGeneralError
		response.VirtualAccountData = biModels.VirtualAccountDataInquiry{}.Default()
		response.VirtualAccountData.BillDetails = payload.BillDetails
		response.VirtualAccountData.FreeTexts = payload.FreeTexts
		response.AdditionalInfo = payload.AdditionalInfo

		return &response, nil
	}

	compressedResponse, err := biUtil.CompressData(responseBytes)
	if err != nil {
		slog.Error("error compressing response", "error", err)
		response.BCAResponse = bca.BCAPaymentFlagResponseGeneralError
		response.VirtualAccountData = biModels.VirtualAccountDataInquiry{}.Default()
		response.VirtualAccountData.BillDetails = payload.BillDetails
		response.VirtualAccountData.FreeTexts = payload.FreeTexts
		response.AdditionalInfo = payload.AdditionalInfo

		return &response, nil
	}

	if err := s.RDB.RDB.Set(ctx, key, string(compressedResponse), 0).Err(); err != nil {
		slog.Error("error saving response to redis", "error", err)
		response.BCAResponse = bca.BCAPaymentFlagResponseGeneralError
		response.VirtualAccountData = biModels.VirtualAccountDataInquiry{}.Default()
		response.VirtualAccountData.BillDetails = payload.BillDetails
		response.VirtualAccountData.FreeTexts = payload.FreeTexts
		response.AdditionalInfo = payload.AdditionalInfo

		return &response, nil
	}

	if !externalUnique {
		slog.Debug("externalId is not unique")

		response.BCAResponse = bca.BCAPaymentFlagResponseDuplicateExternalID
		response.VirtualAccountData.PaymentFlagReason.English = "Cannot use the same X-EXTERNAL-ID"
		response.VirtualAccountData.PaymentFlagReason.Indonesia = "Tidak bisa menggunakan X-EXTERNAL-ID yang sama"
	}

	return &response, nil
}
func (s *BCAService) InquiryVACore(ctx context.Context, response *biModels.BCAInquiryVAResponse, payload *biModels.BCAInquiryRequest) error {

	amountPaid, amountTotal, expDate, err := s.GetVirtualAccountPaidTotalAmountByInquiryRequestId(ctx, strings.ReplaceAll(payload.VirtualAccountNo, " ", ""))
	if eris.Cause(err) == sql.ErrNoRows {

		slog.Debug("va not found in database")
		response.BCAResponse = bca.BCAPaymentFlagResponseVANotFound
		response.VirtualAccountData.PaymentFlagReason.English = "Bill Not Found"
		response.VirtualAccountData.PaymentFlagReason.Indonesia = "Tagihan Tidak Ditemukan"
		response.VirtualAccountData.PaymentFlagStatus = "01"
		return eris.Wrap(err, "va not found")
	} else if err != nil {
		slog.Error("error getting virtual account paid amount by inquiry request id", "error", eris.Cause(err))
		response.BCAResponse = bca.BCAPaymentFlagResponseGeneralError
		response.VirtualAccountData = biModels.VirtualAccountDataInquiry{}.Default()
		response.AdditionalInfo = map[string]interface{}{}
		return eris.Wrap(err, "get virtual account paid total amount by inquiry request id")
	}

	if payload.PaidAmount.Value == "" || payload.PaidAmount.Value == "0.00" || payload.TotalAmount.Value == "" || payload.TotalAmount.Value == "0.00" {
		slog.Debug("va has been paid")
		response.BCAResponse = bca.BCAPaymentFlagResponseInvalidAmount
		response.VirtualAccountData.PaymentFlagReason.English = "Invalid Amount at Paid Amount or Total Amount"
		response.VirtualAccountData.PaymentFlagReason.Indonesia = "Jumlah Tidak Valid pada Jumlah Bayar atau Jumlah Total"
		response.VirtualAccountData.PaymentFlagStatus = "01"
		response.VirtualAccountData.PaidAmount = biModels.Amount{}
		response.VirtualAccountData.TotalAmount = biModels.Amount{}
		return nil
	}

	if amountPaid.Value != "" && amountPaid.Value != "0.00" {
		slog.Debug("va has been paid")
		response.BCAResponse = bca.BCAPaymentFlagResponseVAPaid
		response.VirtualAccountData.PaymentFlagReason.English = "Bill has been paid"
		response.VirtualAccountData.PaymentFlagReason.Indonesia = "Tagihan Telah Terbayar"
		response.VirtualAccountData.PaymentFlagStatus = "01"
		return nil
	}
	nExpDate, _ := time.Parse(time.DateTime, expDate)
	if time.Now().After(nExpDate) {
		slog.Debug("va not found in database")

		response.BCAResponse = bca.BCAPaymentFlagResponseVAExpired
		response.VirtualAccountData.PaymentFlagReason.English = "Bill has been expired"
		response.VirtualAccountData.PaymentFlagReason.Indonesia = "Tagihan sudah kadarluasa"
		response.VirtualAccountData.PaymentFlagStatus = "01"

		return eris.Wrap(err, "va not found")
	}

	if amountTotal.Value != payload.PaidAmount.Value {
		slog.Debug("paid amount is not equal to total amount")
		response.BCAResponse = bca.BCAPaymentFlagResponseInvalidAmount
		response.VirtualAccountData.PaymentFlagReason.English = "Invalid Amount"
		response.VirtualAccountData.PaymentFlagReason.Indonesia = "Jumlah yang dibayarkan tidak sesuai"
		response.VirtualAccountData.PaymentFlagStatus = "01"
		response.VirtualAccountData.PaidAmount = biModels.Amount{}
		response.VirtualAccountData.TotalAmount = biModels.Amount{}
		return nil
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Debug("error beginning transaction", "error", err)
		tx.Rollback()
		response.BCAResponse = bca.BCAPaymentFlagResponseGeneralError
		response.VirtualAccountData = biModels.VirtualAccountDataInquiry{}.Default()
		response.AdditionalInfo = map[string]interface{}{}
		return eris.Wrap(err, "beginning transaction")
	}

	statement := `
	UPDATE va_request SET paidAmountValue = ?, 
						  paidAmountCurrency = ?, 
						  id_va_status = ?   
	WHERE inquiryRequestId = ?
	`
	_, err = tx.ExecContext(ctx, statement, payload.PaidAmount.Value, payload.PaidAmount.Currency, biUtil.VAStatusPaid,
		payload.PaymentRequestID)
	if err != nil {
		slog.Error("error updating va_request", "error", eris.Cause(err))
		tx.Rollback()

		response.BCAResponse = bca.BCAPaymentFlagResponseGeneralError
		response.VirtualAccountData = biModels.VirtualAccountDataInquiry{}.Default()
		response.AdditionalInfo = map[string]interface{}{}

		return eris.Wrap(err, "updating va_request")
	}

	statement = `
	SELECT  partnerServiceId, customerNo, virtualAccountNo, virtualAccountName, totalAmountValue,
			totalAmountCurrency
	FROM va_request 
	WHERE inquiryRequestId = ?
	LIMIT 1
	`
	if err := tx.QueryRowContext(ctx, statement, payload.PaymentRequestID).Scan(
		&response.VirtualAccountData.PartnerServiceID,
		&response.VirtualAccountData.CustomerNo,
		&response.VirtualAccountData.VirtualAccountNo,
		&response.VirtualAccountData.VirtualAccountName,
		&response.VirtualAccountData.TotalAmount.Value,
		&response.VirtualAccountData.TotalAmount.Currency,
	); err != nil {
		if err == sql.ErrNoRows {
			// fmt.Println("masuk sini")
			// TODO : Decide if rollback in this step is necessary or not
			tx.Rollback()

			response.BCAResponse = bca.BCAPaymentFlagResponseVANotFound
			response.VirtualAccountData.PaymentFlagReason.English = "Bill Not Found"
			response.VirtualAccountData.PaymentFlagReason.Indonesia = "Tagihan Tidak Ditemukan"
			response.VirtualAccountData.PaymentFlagStatus = "01"

			return nil
		} else {
			slog.Error("error querying va_request", "error", eris.Cause(err))
			tx.Rollback()

			response.BCAResponse = bca.BCAPaymentFlagResponseGeneralError
			response.VirtualAccountData = biModels.VirtualAccountDataInquiry{}.Default()
			response.AdditionalInfo = map[string]interface{}{}

			return eris.Wrap(err, "querying va_request")
		}
	}

	response.BCAResponse = bca.BCAPaymentFlagResponseSuccess
	response.VirtualAccountData.PaymentFlagReason.English = "Success"
	response.VirtualAccountData.PaymentFlagReason.Indonesia = "Sukses"

	response.VirtualAccountData.PaidAmount = payload.PaidAmount
	response.VirtualAccountData.TotalAmount = payload.TotalAmount
	response.VirtualAccountData.PaymentFlagStatus = "00"

	if err = tx.Commit(); err != nil {
		slog.Debug("InquiryVACore", "error committing transaction", eris.Cause(err))
		tx.Rollback()

		response.BCAResponse = bca.BCAPaymentFlagResponseGeneralError
		response.VirtualAccountData = biModels.VirtualAccountDataInquiry{}.Default()
		response.AdditionalInfo = map[string]interface{}{}

		return eris.Wrap(err, "committing transaction")
	}

	return nil
}

// Transaction Watcher
func (s *BCAService) GetWatchedTransaction(ctx context.Context) []*biModels.TransactionWatcherPublic {
	return s.Watcher.GetWatchers()
}

// Service Utils
func (s *BCAService) RequestHandler(ctx context.Context, request *http.Request) (string, error) {

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return "", eris.Wrap(err, "sending request")
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", eris.Wrap(err, "reading response body")
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		slog.Debug("Non-200 status code", "status", response.StatusCode)
		var obj biModels.BCAResponse

		if err := json.Unmarshal(body, &obj); err != nil {
			return "", eris.Wrap(err, "unmarshalling error response")
		}

		obj.HTTPStatusCode = response.StatusCode

		response, err := json.Marshal(obj)
		if err != nil {
			return "", eris.Wrap(err, "marshalling error response")
		}

		return string(response), eris.New("non-200 status code")
	}

	return string(body), nil
}

func (s *BCAService) BuildNumVA(idUser, idJenis int, partnerId string) (string, string) {

	partnerId += "0" + strconv.Itoa(idJenis)
	nIdU := strconv.Itoa(idUser)
	customerNo := ""
	for i := 0; i < 10-len(nIdU); i++ {
		customerNo += "0"
	}
	customerNo += nIdU
	return partnerId + customerNo, customerNo
}

// GenerateVANumber is called to generate a new virtual account number or return existing va number (this is the case if the requester is a user)
func (s *BCAService) GenerateVANumber() (string, error) {
	return "", nil
}

func (s *BCAService) CheckVAPaid(ctx context.Context, virtualAccountNum string, tx *sql.Tx) (bool, error) {
	// partnerId := s.Config.BCAPartnerId.BCAPartnerId
	query := `
	SELECT paidAmountValue,paidAmountCurrency,expired_date 
	FROM va_request 
	WHERE TRIM(virtualAccountNo) = ? AND paidAmountValue = '0.00' AND id_va_status = ?
	`
	var amount biModels.Amount
	expDate := ""
	err := tx.QueryRowContext(ctx, query, strings.ReplaceAll(virtualAccountNum, " ", ""), biUtil.VAStatusPending).Scan(&amount.Value, &amount.Currency, &expDate)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		tx.Rollback()
		return false, eris.Wrap(err, "querying va_table")
	}
	nExp, _ := time.Parse(time.DateTime, expDate)
	if time.Now().After(nExp) {
		return true, nil
	}
	return false, nil
}
func (s *BCAService) GetVirtualAccountPaidAmountByInquiryRequestId(ctx context.Context, inquiryRequestId string) (*biModels.Amount, error) {
	var amount biModels.Amount
	query := `
	SELECT totalAmountValue, totalAmountCurrency 
	FROM va_request
	WHERE inquiryRequestId = ?
	`
	if err := s.DB.QueryRowContext(ctx, query, inquiryRequestId).Scan(
		&amount.Value, &amount.Currency,
	); err != nil {
		slog.Debug("error querying va_request", "error", err)
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, eris.Wrap(err, "querying va_request")
		}
	}
	return &amount, nil
}

func (s *BCAService) GetVirtualAccountPaidTotalAmountByInquiryRequestId(ctx context.Context, inquiryRequestId string) (*biModels.Amount, *biModels.Amount, string, error) {
	var amountPaid biModels.Amount
	var amountTotal biModels.Amount
	var expDate string
	query := `
	SELECT paidAmountValue, paidAmountCurrency,totalAmountValue, totalAmountCurrency, COALESCE(expired_date, CURRENT_TIMESTAMP()) AS effective_expired_date  
	FROM va_request 
	WHERE TRIM(virtualAccountNo) = ?  ORDER BY created_at DESC
	LIMIT 1
	`
	err := s.DB.QueryRowContext(ctx, query, inquiryRequestId).Scan(&amountPaid.Value, &amountPaid.Currency, &amountTotal.Value, &amountTotal.Currency, &expDate)
	if err != nil {
		return &amountPaid, &amountTotal, "", eris.Wrap(err, "querying va_request")
	}
	return &amountPaid, &amountTotal, expDate, nil
}

func (s *BCAService) GetVirtualAccountPaidByInquiryRequestId(ctx context.Context, vaNum string) (*biModels.Amount, *biModels.Amount, error) {
	var amountPaid biModels.Amount
	var amountTotal biModels.Amount
	query := `
	SELECT paidAmountValue, paidAmountCurrency,totalAmountValue, totalAmountCurrency  
	FROM va_request 
	WHERE virtualAccountNo = ? 
	ORDER BY created_at DESC
	LIMIT 1
	`
	err := s.DB.QueryRowContext(ctx, query, vaNum).Scan(&amountPaid.Value, &amountPaid.Currency, &amountTotal.Value, &amountTotal.Currency)
	if err != nil {
		return &amountPaid, &amountTotal, eris.Wrap(err, "querying va_request")
	}
	return &amountPaid, &amountTotal, nil
}

func (s *BCAService) VerifyAdditionalBillPresentmentRequiredHeader(ctx context.Context, request *http.Request) (*biModels.BCAResponse, error) {

	// For bill presentment, we need to verify the header
	// 1. channel id
	// 2. partner id

	// Parse the request header
	channelID := request.Header.Get("CHANNEL-ID")
	partnerID := request.Header.Get("X-PARTNER-ID")

	if channelID == "" {
		response := bca.BCABillInquiryResponseMissingMandatoryField
		response.ResponseMessage = "Invalid Mandatory Field {CHANNEL-ID}"

		return &response, nil
	}
	if channelID != s.Config.BCAConfig.ChannelID {
		response := bca.BCABillInquiryResponseUnauthorizedUnknownClient

		return &response, nil
	}

	if partnerID == "" {
		response := bca.BCABillInquiryResponseMissingMandatoryField
		response.ResponseMessage = "Invalid Mandatory Field {X-PARTNER-ID}"

		return &response, nil
	}
	if partnerID != s.Config.BCAPartnerInformation.BCAPartnerId {
		response := bca.BCABillInquiryResponseUnauthorizedUnknownClient

		return &response, nil
	}

	return &bca.BCABillInquiryResponseSuccess, nil
}

func (s *BCAService) GetAllVAWaitingPayment(ctx context.Context) error {
	query := `
	SELECT id, expired_date 
	FROM va_request 
	WHERE id_va_status = ? AND id_bank = ?
	`
	rows, err := s.DB.QueryContext(ctx, query, biUtil.VAStatusPending, s.Config.BCAConfig.InternalBankID)
	if err != nil {
		return eris.Wrap(err, "querying va_request")
	}
	defer rows.Close()

	loc, _ := time.LoadLocation(os.Getenv("TZ"))
	for rows.Next() {
		obj := watcher.NewWatcher()
		expireAt := ""
		err = rows.Scan(&obj.IDTransaction, &expireAt)
		if err != nil {
			return eris.Wrap(err, "scan va_table")
		}
		obj.ExpireAt, err = time.ParseInLocation(time.DateTime, expireAt, loc)
		if err != nil {
			slog.Error("skipping transaction due to parsing time error", "error", err, "transaction id", obj.IDTransaction)
			continue
		}
		obj.IDBank = s.Config.InternalBankID
		obj.BankName = s.Config.InternalBankName

		slog.Debug("adding watcher", "id", obj.IDTransaction, "expireAt", obj.ExpireAt)

		s.Watcher.AddWatcher(obj)
	}

	return nil
}

func (s *BCAService) getInternalBankInfo() error {

	statement := `
	SELECT id, bank_name 
	FROM authenticated_banks
	WHERE client_id = ? AND client_secret = ? AND deleted_at IS NULL
	LIMIT 1
	`
	if err := s.DB.QueryRow(statement, s.Config.BCARequestedClientCredentials.ClientID,
		s.Config.BCARequestedClientCredentials.ClientSecret).Scan(
		&s.Config.BCAConfig.InternalBankID, &s.Config.BCAConfig.InternalBankName); err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("unauthorized bank credentials")
			return eris.New("unauthorized")
		}
		slog.Error("error getting internal bank id of BCAService", "error", err)
		return err
	}

	slog.Debug("internal bank info", "id", s.Config.BCAConfig.InternalBankID, "name", s.Config.BCAConfig.InternalBankName)

	return nil
}
