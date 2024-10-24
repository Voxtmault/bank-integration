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
	"strconv"
	"strings"
	"time"

	"github.com/rotisserie/eris"
	"github.com/voxtmault/bank-integration/bca"
	biConfig "github.com/voxtmault/bank-integration/config"
	biInterfaces "github.com/voxtmault/bank-integration/interfaces"
	biModels "github.com/voxtmault/bank-integration/models"
	biStorage "github.com/voxtmault/bank-integration/storage"
	biUtil "github.com/voxtmault/bank-integration/utils"
)

type BCAService struct {

	// Dependency Injection
	Egress          biInterfaces.RequestEgress
	Ingress         biInterfaces.RequestIngress
	GeneralSecurity biUtil.GeneralSecurity

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

func NewBCAService(egress biInterfaces.RequestEgress, ingress biInterfaces.RequestIngress, config *biConfig.BankingConfig, db *sql.DB, rdb *biStorage.RedisInstance) *BCAService {
	return &BCAService{
		Egress:  egress,
		Ingress: ingress,
		Config:  config,
		DB:      db,
		RDB:     rdb,
	}
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
func (s *BCAService) Middleware(ctx context.Context, request *http.Request) (*biModels.BCAResponse, []byte, error) {

	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		slog.Debug("error reading request body", "error", err)
		return &bca.BCABillInquiryResponseRequestParseError, nil, nil
	}
	defer request.Body.Close()
	fmt.Println("Body Bytes: ", string(bodyBytes))

	// Set the body back to the original state
	request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	result, response := s.Ingress.VerifySymmetricSignature(ctx, request, s.RDB)
	if response != nil {
		slog.Debug("verifying symmetric signature failed", "response", response.ResponseMessage)

		return response, bodyBytes, nil
	}

	if !result {
		return &bca.BCABillInquiryResponseUnauthorizedSignature, bodyBytes, nil
	}

	return &bca.BCAAuthResponseSuccess, bodyBytes, nil
}

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

func (s *BCAService) BillPresentment(ctx context.Context, data []byte) (*biModels.VAResponsePayload, error) {
	var obj biModels.VAResponsePayload
	obj.BCAResponse = &biModels.BCAResponse{}
	inqueryReason := biModels.InquiryReason{}

	obj.VirtualAccountData = &biModels.VABCAResponseData{}
	// fmt.Println("masuk")
	obj.VirtualAccountData.BillDetails = []biModels.BillInfo{}
	// fmt.Println("masuk")
	obj.VirtualAccountData.FreeTexts = []biModels.FreeText{}
	obj.VirtualAccountData.AdditionalInfo = map[string]interface{}{}
	obj.VirtualAccountData.InquiryReason = inqueryReason

	var payload biModels.BCAVARequestPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		slog.Debug("error un marshaling request body", "error", err)
		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCABillInquiryResponseRequestParseError.Data()
		inqueryReason.English = "Error When Unmarshalling Request Body"
		inqueryReason.Indonesia = "Kesalahan saat melakukan unmarshalling pada badan Request Body"
		obj.VirtualAccountData.InquiryStatus = "01"
		obj.VirtualAccountData.InquiryReason = inqueryReason
		// obj.VirtualAccountData.VirtualAccountTrxType = "C"
		// obj.VirtualAccountData.PartnerServiceID =
		return &obj, nil
	}
	obj.VirtualAccountData.SubCompany = "00000"
	obj.VirtualAccountData.VirtualAccountTrxType = "C"

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Debug("error beginning transaction", "error", err)
		tx.Rollback()
		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCABillInquiryResponseGeneralError.Data()
		return &obj, eris.Wrap(err, "beginning transaction")
	}
	paidAmount := biModels.Amount{}
	statement := `
	SELECT partnerServiceId, customerNo, virtualAccountNo, virtualAccountName, totalAmountValue, totalAmountCurrency, paidAmountValue, paidAmountCurrency
	FROM va_request 
	WHERE TRIM(virtualAccountNo) = ?
	ORDER BY created_at DESC
	LIMIT 1
	`
	err = tx.QueryRowContext(ctx, statement, strings.ReplaceAll(payload.VirtualAccountNo, " ", "")).Scan(
		&obj.VirtualAccountData.PartnerServiceID,
		&obj.VirtualAccountData.CustomerNo,
		&obj.VirtualAccountData.VirtualAccountNo,
		&obj.VirtualAccountData.VirtualAccountName,
		&obj.VirtualAccountData.TotalAmount.Value,
		&obj.VirtualAccountData.TotalAmount.Currency,
		&paidAmount.Value,
		&paidAmount.Currency)
	if err == sql.ErrNoRows {
		slog.Debug("bill presentment", "error", "va not found")
		tx.Rollback()
		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCABillInquiryResponseVANotFound.Data()
		inqueryReason.English = "Bill Not Found"
		inqueryReason.Indonesia = "Tagihan tidak ditemukan"
		obj.VirtualAccountData.InquiryStatus = "01"
		obj.VirtualAccountData.InquiryReason = inqueryReason
		obj.VirtualAccountData.PartnerServiceID = payload.PartnerServiceID
		obj.VirtualAccountData.CustomerNo = payload.CustomerNo
		obj.VirtualAccountData.VirtualAccountNo = payload.VirtualAccountNo
		obj.VirtualAccountData.InquiryRequestID = payload.InquiryRequestID
		return &obj, nil
	} else if err != nil {
		slog.Debug("error querying va_request", "error", err)
		tx.Rollback()
		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCABillInquiryResponseGeneralError.Data()
		return &obj, nil
	}
	if paidAmount.Value != "0.00" && paidAmount.Value != "" {
		slog.Debug("va has been paid")
		tx.Rollback()
		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCABillInquiryResponseVAPaid.Data()
		inqueryReason.English = "Paid Bill"
		inqueryReason.Indonesia = "Tagihan Telah Terbayar"
		obj.VirtualAccountData.InquiryStatus = "01"
		obj.VirtualAccountData.InquiryReason = inqueryReason
		obj.VirtualAccountData.PartnerServiceID = payload.PartnerServiceID
		obj.VirtualAccountData.CustomerNo = payload.CustomerNo
		obj.VirtualAccountData.VirtualAccountNo = payload.VirtualAccountNo
		obj.VirtualAccountData.InquiryRequestID = payload.InquiryRequestID
		return &obj, nil
	}
	// fmt.Println("sini 3")

	statement = `
	UPDATE va_request SET inquiryRequestId = ? 
	WHERE virtualAccountNo = ? AND paidAmountValue = '0.00'
	`
	_, err = tx.QueryContext(ctx, statement, payload.InquiryRequestID, payload.VirtualAccountNo)
	if err != nil {
		slog.Debug("error updating va_request", "error", err)
		tx.Rollback()
		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCABillInquiryResponseGeneralError.Data()
		return &obj, eris.Wrap(err, "updating va_request")
	}

	obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCABillInquiryResponseSuccess.Data()
	obj.VirtualAccountData.InquiryRequestID = payload.InquiryRequestID
	obj.VirtualAccountData.InquiryStatus = "00"
	inqueryReason.Indonesia = "Sukses"
	inqueryReason.English = "Success"
	obj.VirtualAccountData.InquiryReason = inqueryReason
	if err = tx.Commit(); err != nil {
		slog.Debug("bill presentment", "error committing transaction", err)
		tx.Rollback()
		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCABillInquiryResponseGeneralError.Data()
		return &obj, eris.Wrap(err, "committing transaction")
	}

	return &obj, nil
}

func (s *BCAService) InquiryVA(ctx context.Context, data []byte) (*biModels.BCAInquiryVAResponse, error) {
	var obj biModels.BCAInquiryVAResponse
	obj.VirtualAccountData = &biModels.VirtualAccountDataInqury{}
	inqueryReason := biModels.Reason{}
	obj.VirtualAccountData.BillDetails = []biModels.BillDetail{}
	obj.VirtualAccountData.FreeTexts = []biModels.FreeText{}
	obj.AdditionalInfo = map[string]interface{}{}
	var payload biModels.BCAInquiryRequest
	if err := json.Unmarshal(data, &payload); err != nil {
		slog.Debug("error un-marshaling request body", "error", err)
		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseRequestParseError.Data()
		inqueryReason.English = "Error When Unmarshalling Request Body"
		inqueryReason.Indonesia = "Kesalahan saat melakukan unmarshalling pada badan Request Body"
		obj.VirtualAccountData.PaymentFlagStatus = "01"
		obj.VirtualAccountData.PaymentFlagReason = inqueryReason
		return &obj, nil
	}
	if payload.PaidAmount.Value == "0.00" || payload.PaidAmount.Value == "0" || payload.PaidAmount.Value == "" {
		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseVANotFound.Data()
		obj.VirtualAccountData.PaymentRequestID = payload.PaymentRequestID
		obj.VirtualAccountData.FlagAdvise = "N"
		inqueryReason.English = "Bill Not Found"
		inqueryReason.Indonesia = "Tagihan Tidak Ditemukan"
		obj.VirtualAccountData.PaymentFlagStatus = "01"
		obj.VirtualAccountData.CustomerNo = payload.CustomerNo
		obj.VirtualAccountData.VirtualAccountNo = payload.VirtualAccountNo
		obj.VirtualAccountData.TrxDateTime = payload.TrxDateTime
		obj.VirtualAccountData.PaidAmount = biModels.Amount{}
		obj.VirtualAccountData.TotalAmount = biModels.Amount{}
		return &obj, nil
	}
	obj.VirtualAccountData.PaymentRequestID = payload.PaymentRequestID
	obj.VirtualAccountData.FlagAdvise = "N"
	obj.VirtualAccountData.ReferenceNo = payload.ReferenceNo
	obj.VirtualAccountData.CustomerNo = payload.CustomerNo
	obj.VirtualAccountData.VirtualAccountNo = payload.VirtualAccountNo
	obj.VirtualAccountData.TrxDateTime = payload.TrxDateTime
	obj.VirtualAccountData.PaidAmount = biModels.Amount{}
	obj.VirtualAccountData.TotalAmount = biModels.Amount{}
	amountPaid, amountTotal, err := s.GetVirtualAccountPaidnTotalAmountByInquiryRequestId(ctx, payload.PaymentRequestID)
	if eris.Cause(err) == sql.ErrNoRows {
		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseVANotFound.Data()
		inqueryReason.English = "Bill Not Found"
		inqueryReason.Indonesia = "Tagihan Tidak Ditemukan"
		obj.VirtualAccountData.PaymentFlagStatus = "01"
		obj.VirtualAccountData.PaymentFlagReason = inqueryReason
		obj.VirtualAccountData.CustomerNo = payload.CustomerNo
		obj.VirtualAccountData.VirtualAccountNo = payload.VirtualAccountNo
		obj.VirtualAccountData.PartnerServiceID = payload.PartnerServiceID
		return &obj, eris.Wrap(err, "querying va_table for")
	} else if err != nil {
		slog.Debug("error getting virtual account paid amount by inquiry request id", "error", err)
		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCABillInquiryResponseGeneralError.Data()
		return &obj, eris.Wrap(err, "Error Find VA")
	}

	if amountPaid.Value != "" && amountPaid.Value != "0.00" {
		slog.Debug("va has been paid")
		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseVAPaid.Data()
		inqueryReason.English = "Paid Bill"
		inqueryReason.Indonesia = "Tagihan Telah Terbayar"
		obj.VirtualAccountData.PaymentFlagStatus = "01"
		obj.VirtualAccountData.PaymentFlagReason = inqueryReason
		obj.VirtualAccountData.PartnerServiceID = payload.PartnerServiceID
		obj.VirtualAccountData.CustomerNo = payload.CustomerNo
		obj.VirtualAccountData.VirtualAccountNo = payload.VirtualAccountNo
		obj.VirtualAccountData.PartnerServiceID = payload.PartnerServiceID
		obj.VirtualAccountData.PaidAmount = payload.PaidAmount
		obj.VirtualAccountData.TotalAmount = payload.TotalAmount
		return &obj, nil
	}

	// amount, err := s.GetVirtualAccountTotalAmountByInquiryRequestId(ctx, payload.PaymentRequestID)
	// if eris.Cause(err) == sql.ErrNoRows {
	// 	obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseVANotFound.Data()
	// 	inqueryReason.English = "Bill Not Found"
	// 	inqueryReason.Indonesia = "Tagihan Tidak Ditemukan"
	// 	obj.VirtualAccountData.PaymentFlagStatus = "01"
	// 	obj.VirtualAccountData.PaymentFlagReason = inqueryReason
	// 	obj.VirtualAccountData.CustomerNo = payload.CustomerNo
	// 	obj.VirtualAccountData.VirtualAccountNo = payload.VirtualAccountNo
	// 	obj.VirtualAccountData.PartnerServiceID = payload.PartnerServiceID
	// 	// fmt.Println("disini 1")
	// 	return &obj, eris.Wrap(err, "querying va_table")
	// } else if err != nil {
	// 	obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseGeneralError.Data()
	// 	return &obj, eris.Wrap(err, "querying va_table")
	// }
	// This could be because the payment request ID is not found in the database
	if amountPaid == nil || amountTotal == nil {
		slog.Debug("payment request ID not found in database")
		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseVANotFound.Data()
		inqueryReason.English = "Bill Not Found"
		inqueryReason.Indonesia = "Tagihan Tidak Ditemukan"
		obj.VirtualAccountData.PaymentFlagStatus = "01"
		obj.VirtualAccountData.PaymentFlagReason = inqueryReason
		obj.VirtualAccountData.CustomerNo = payload.CustomerNo
		obj.VirtualAccountData.VirtualAccountNo = payload.VirtualAccountNo
		obj.VirtualAccountData.PartnerServiceID = payload.PartnerServiceID
		obj.VirtualAccountData.ReferenceNo = payload.ReferenceNo
		obj.VirtualAccountData.TrxDateTime = payload.TrxDateTime
		obj.VirtualAccountData.PaidAmount = biModels.Amount{}
		obj.VirtualAccountData.TotalAmount = biModels.Amount{}

		return &obj, nil
	} else {
		if amountTotal.Value != payload.PaidAmount.Value {
			obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseVANotFound.Data()
			inqueryReason.English = "Bill Not Found"
			inqueryReason.Indonesia = "Tagihan Tidak Ditemukan"
			obj.VirtualAccountData.PaymentFlagStatus = "01"
			obj.VirtualAccountData.PaymentFlagReason = inqueryReason
			obj.VirtualAccountData.CustomerNo = payload.CustomerNo
			obj.VirtualAccountData.VirtualAccountNo = payload.VirtualAccountNo
			obj.VirtualAccountData.PartnerServiceID = payload.PartnerServiceID
			obj.VirtualAccountData.PaidAmount = biModels.Amount{}
			obj.VirtualAccountData.TotalAmount = biModels.Amount{}

			return &obj, nil
		}
		tx, err := s.DB.BeginTx(ctx, nil)
		if err != nil {
			slog.Debug("error beginning transaction", "error", err)
			tx.Rollback()

			obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCABillInquiryResponseGeneralError.Data()
			return &obj, eris.Wrap(err, "beginning transaction")
		}

		statement := `
		UPDATE va_request SET paidAmountValue = ?, 
							  paidAmountCurrency = ?, 
							  id_va_status = 2   
		WHERE inquiryRequestId = ?
		`
		_, err = tx.ExecContext(ctx, statement, payload.PaidAmount.Value, payload.PaidAmount.Currency,
			payload.PaymentRequestID)
		if err != nil {
			slog.Debug("error updating va_request", "error", err)
			obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseGeneralError.Data()
			return &obj, eris.Wrap(err, "updating va_request")
		}

		obj.VirtualAccountData = &biModels.VirtualAccountDataInqury{}
		statement = `
		SELECT  partnerServiceId, customerNo, virtualAccountNo, virtualAccountName, totalAmountValue,
				totalAmountCurrency
		FROM va_request 
		WHERE inquiryRequestId = ?
		LIMIT 1
		`
		if err := tx.QueryRowContext(ctx, statement, payload.PaymentRequestID).Scan(
			&obj.VirtualAccountData.PartnerServiceID,
			&obj.VirtualAccountData.CustomerNo,
			&obj.VirtualAccountData.VirtualAccountNo,
			&obj.VirtualAccountData.VirtualAccountName,
			&obj.VirtualAccountData.TotalAmount.Value,
			&obj.VirtualAccountData.TotalAmount.Currency,
		); err != nil {
			slog.Debug("error querying va_request", "error", err)

			if err == sql.ErrNoRows {
				obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseVANotFound.Data()
				inqueryReason.English = "Bill Not Found"
				inqueryReason.Indonesia = "Tagihan Tidak Ditemukan"
				obj.VirtualAccountData.PaymentFlagStatus = "01"
				obj.VirtualAccountData.PaymentFlagReason = inqueryReason
				obj.VirtualAccountData.CustomerNo = payload.CustomerNo
				obj.VirtualAccountData.VirtualAccountNo = payload.VirtualAccountNo
				obj.VirtualAccountData.PartnerServiceID = payload.PartnerServiceID
				obj.VirtualAccountData.PaidAmount = biModels.Amount{}
				obj.VirtualAccountData.TotalAmount = biModels.Amount{}
				return &obj, nil
			} else {
				obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseGeneralError.Data()
				return &obj, eris.Wrap(err, "querying va_request")
			}
		}
		obj.VirtualAccountData.FlagAdvise = "N"
		obj.VirtualAccountData.ReferenceNo = payload.ReferenceNo
		obj.VirtualAccountData.CustomerNo = payload.CustomerNo
		obj.VirtualAccountData.VirtualAccountNo = payload.VirtualAccountNo
		obj.VirtualAccountData.TrxDateTime = payload.TrxDateTime
		obj.VirtualAccountData.PaidAmount = payload.PaidAmount
		obj.VirtualAccountData.TotalAmount = payload.TotalAmount
		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseSuccess.Data()
		obj.VirtualAccountData.PaidAmount.Value = payload.PaidAmount.Value
		obj.VirtualAccountData.PaidAmount.Currency = payload.PaidAmount.Currency
		obj.VirtualAccountData.PaymentRequestID = payload.PaymentRequestID
		obj.VirtualAccountData.PaymentFlagStatus = "00"
		inqueryReason.Indonesia = "Sukses"
		inqueryReason.English = "Success"
		obj.VirtualAccountData.PaymentFlagReason = inqueryReason
		obj.VirtualAccountData.BillDetails = []biModels.BillDetail{}
		obj.VirtualAccountData.FreeTexts = []biModels.FreeText{}
		obj.AdditionalInfo = map[string]interface{}{}
		if err = tx.Commit(); err != nil {
			slog.Debug("bill presentment", "error committing transaction", err)
			tx.Rollback()

			obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseGeneralError.Data()
			return &obj, eris.Wrap(err, "committing transaction")
		}

		return &obj, nil
	}
}

func (s *BCAService) CreateVA(ctx context.Context, payload *biModels.CreateVAReq) error {
	partnerId := "   " + s.Config.BCAPartnerInformation.BCAPartnerId
	query := `
	INSERT INTO va_request (partnerServiceId, customerNo, virtualAccountNo, totalAmountValue, 
				   			virtualAccountName, id_user, owner_table)
	VALUES(?,?,?,?,?,?,?)
	`
	numVA, customerNo := s.BuildNumVA(payload.IdUser, payload.IdJenisUser, partnerId)

	cekpaid, err := s.CheckVAPaid(ctx, numVA)
	if err != nil {
		return eris.Wrap(err, "querying va_table")
	}
	if cekpaid {
		_, err = s.DB.ExecContext(ctx, query, partnerId, customerNo, numVA, payload.JumlahPembayaran, payload.NamaUser, payload.IdUser, payload.IdJenisUser)
		if err != nil {
			return eris.Wrap(err, "querying va_table")
		}
	} else {
		return eris.Wrap(err, "Va Not Paid")
	}
	return nil
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

func (s *BCAService) CheckVAPaid(ctx context.Context, virtualAccountNum string) (bool, error) {
	// partnerId := s.Config.BCAPartnerId.BCAPartnerId
	query := `
	SELECT paidAmountValue,paidAmountCurrency FROM va_table WHERE TRIM(virtualAccountNo) = ? AND paidAmountValue = '0.00'
	`
	var amount biModels.Amount
	err := s.DB.QueryRowContext(ctx, query, strings.ReplaceAll(virtualAccountNum, " ", "")).Scan(&amount.Value, &amount.Currency)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, eris.Wrap(err, "querying va_table")
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

func (s *BCAService) GetVirtualAccountPaidnTotalAmountByInquiryRequestId(ctx context.Context, inquiryRequestId string) (*biModels.Amount, *biModels.Amount, error) {
	var amountPaid biModels.Amount
	var amountTotal biModels.Amount
	query := `
	SELECT paidAmountValue, paidAmountCurrency,totalAmountValue, totalAmountCurrency  
	FROM va_request 
	WHERE inquiryRequestID = ?  ORDER BY created_at DESC
	LIMIT 1
	`
	err := s.DB.QueryRowContext(ctx, query, inquiryRequestId).Scan(&amountPaid.Value, &amountPaid.Currency, &amountTotal.Value, &amountTotal.Currency)
	if err != nil {
		return &amountPaid, &amountTotal, eris.Wrap(err, "querying va_request")
	}
	return &amountPaid, &amountTotal, nil
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

func (s *BCAService) VerifyAdditionalInquiryVARequiredHeader(ctx context.Context, request *http.Request) (*biModels.BCAResponse, error) {

	// For bill presentment, we need to verify the header
	// 1. channel id
	// 2. partner id

	// Parse the request header
	channelID := request.Header.Get("CHANNEL-ID")
	partnerID := request.Header.Get("X-PARTNER-ID")
	externalID := request.Header.Get("X-EXTERNAL-ID")

	if channelID == "" {
		response := bca.BCAPaymentFlagResponseMissingMandatoryField
		response.ResponseMessage = "Invalid Mandatory Field {CHANNEL-ID}"

		return &response, nil
	}
	if channelID != s.Config.BCAConfig.ChannelID {
		response := bca.BCAPaymentFlagResponseUnauthorizedUnknownClient
		return &response, nil
	}

	if partnerID == "" {
		response := bca.BCAPaymentFlagResponseMissingMandatoryField
		response.ResponseMessage = "Invalid Mandatory Field {X-PARTNER-ID}"

		return &response, nil
	}
	if partnerID != s.Config.BCAPartnerInformation.BCAPartnerId {
		response := bca.BCAPaymentFlagResponseUnauthorizedUnknownClient

		return &response, nil
	}

	// Validate that X-EXTERNAL-ID and paymentRequestId is not the same
	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		slog.Debug("error reading request body", "error", err)
		return &bca.BCABillInquiryResponseRequestParseError, nil
	}
	defer request.Body.Close()

	// Set the body back to the original state
	request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var payload biModels.BCAInquiryRequest
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		slog.Debug("error un-marshaling request body", "error", err)

		return &bca.BCAPaymentFlagResponseRequestParseError, nil
	}

	if externalID == payload.PaymentRequestID {
		slog.Debug("external id and payment request id are the same, inconsistent request")

		return &bca.BCAPaymentFlagResponseDuplicateExternalIDAndPaymentRequestID, nil
	}

	return &bca.BCAPaymentFlagResponseSuccess, nil
}
