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
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rotisserie/eris"
	"github.com/voxtmault/bank-integration/bca"
	"github.com/voxtmault/bank-integration/config"
	"github.com/voxtmault/bank-integration/interfaces"
	"github.com/voxtmault/bank-integration/models"
	"github.com/voxtmault/bank-integration/storage"
	"github.com/voxtmault/bank-integration/utils"
)

type BCAService struct {

	// Dependency Injection
	Egress          interfaces.RequestEgress
	Ingress         interfaces.RequestIngress
	GeneralSecurity utils.GeneralSecurity

	// Configs
	Config *config.BankingConfig

	// Runtime Access Tokens
	AccessToken          string
	AccessTokenExpiresAt int64

	// DB Connections
	DB  *sql.DB
	RDB *storage.RedisInstance
}

var _ interfaces.SNAP = &BCAService{}

func NewBCAService(egress interfaces.RequestEgress, ingress interfaces.RequestIngress, config *config.BankingConfig, db *sql.DB, rdb *storage.RedisInstance) *BCAService {
	return &BCAService{
		Egress:  egress,
		Ingress: ingress,
		Config:  config,
		DB:      db,
		RDB:     rdb,
	}
}

func (s *BCAService) GetAccessToken(ctx context.Context) error {

	// Logic
	// 1. Customize the header of the request (including creating the signature)
	// 2. Send the request
	// 3. Parse the response

	baseUrl := s.Config.BCAConfig.BaseURL + s.Config.BCAURLEndpoints.AccessTokenURL
	method := http.MethodPost
	body := models.GrantType{
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

	var atObj models.AccessTokenResponse
	if err = json.Unmarshal([]byte(response), &atObj); err != nil {
		slog.Debug("error unmarshalling response", "error", err)
		return eris.Wrap(err, "unmarshalling response")
	}

	s.AccessToken = atObj.AccessToken

	// Create internal counter for when the access token expires
	s.AccessTokenExpiresAt = time.Now().Add(time.Second * 900).Unix()

	return nil
}

func (s *BCAService) GenerateAccessToken(ctx context.Context, request *http.Request) (*models.AccessTokenResponse, error) {
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
	var body models.GrantType
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		return nil, eris.Wrap(err, "decoding request body")
	}

	// Validate the received struct
	if err := utils.ValidateStruct(ctx, body); err != nil {
		return nil, eris.Wrap(err, "validating request body")
	}

	// Parse the request header
	timeStamp := request.Header.Get("X-TIMESTAMP")
	clientKey := request.Header.Get("X-CLIENT-KEY")
	// signature := request.Header.Get("X-SIGNATURE")

	// Validate parsed header
	if clientKey == "" {
		slog.Debug("clientKey is empty")
		return &models.AccessTokenResponse{
			BCAResponse: models.BCAResponse{
				HTTPStatusCode:  http.StatusBadRequest,
				ResponseCode:    "4007301",
				ResponseMessage: "Invalid field format [clientId/clientSecret/grantType]",
			},
		}, nil
	} else if timeStamp == "" {
		slog.Debug("timeStamp is empty")
		return &models.AccessTokenResponse{
			BCAResponse: models.BCAResponse{
				HTTPStatusCode:  http.StatusBadRequest,
				ResponseCode:    "4007301",
				ResponseMessage: "Invalid field format [X-TIMESTAMP]",
			},
		}, nil
	}

	// Validate the timestamp format
	if _, err := time.Parse(time.RFC3339, timeStamp); err != nil {
		slog.Debug("invalid timestamp format")
		return &models.AccessTokenResponse{
			BCAResponse: models.BCAResponse{
				HTTPStatusCode:  http.StatusBadRequest,
				ResponseCode:    "4007301",
				ResponseMessage: "Invalid field format [X-TIMESTAMP]",
			},
		}, nil
	}

	// Retrieve the client secret from redis
	clientSecret, err := s.RDB.GetIndividualValueRedisHash(ctx, utils.ClientCredentialsRedis, clientKey)
	if err != nil {
		return nil, eris.Wrap(err, "getting client secret")
	}

	if clientSecret == "" {
		slog.Debug("clientId is not registered")
		return &models.AccessTokenResponse{
			BCAResponse: models.BCAResponse{
				HTTPStatusCode:  http.StatusUnauthorized,
				ResponseCode:    "4007301",
				ResponseMessage: "Unauthorized. [Unknown client]",
			},
		}, nil
	}

	// Verify Asymmetric Signature
	result, err := s.Ingress.VerifyAsymmetricSignature(ctx, request, clientSecret)
	if err != nil {
		return nil, eris.Wrap(err, "verifying asymmetric signature")
	}

	if !result {
		return &models.AccessTokenResponse{
			BCAResponse: models.BCAResponse{
				HTTPStatusCode:  http.StatusUnauthorized,
				ResponseCode:    "4017300",
				ResponseMessage: "Unauthorized. [Signature]",
			},
		}, nil
	}

	// Generate the access token
	token, err := s.GeneralSecurity.GenerateAccessToken(ctx)
	if err != nil {
		slog.Debug("error generating access token", "error", err)
		return nil, eris.Wrap(err, "generating access token")
	}
	slog.Debug("generated token", "token", token)

	// Save the access token to redis along with the configured client secret & expiration time
	key := fmt.Sprintf("%s:%s", utils.AccessTokenRedis, token)
	if err := s.RDB.RDB.Set(ctx, key, clientSecret, time.Second*900).Err(); err != nil {
		return nil, eris.Wrap(err, "saving access token to redis")
	}

	return &models.AccessTokenResponse{
		AccessToken: token,
		TokenType:   "bearer",
		ExpiresIn:   "900",
		BCAResponse: models.BCAResponse{
			HTTPStatusCode:  http.StatusOK,
			ResponseCode:    "2007300",
			ResponseMessage: "Successful",
		},
	}, nil
}

func (s *BCAService) ValidateAccessToken(ctx context.Context, accessToken string) (bool, error) {
	// Logic
	// 1. Get the access token from Redis
	// 2. If redis return nil then return false to the caller
	// 3. if redis returns a value then return true to the caller

	data, err := s.RDB.RDB.Get(ctx, fmt.Sprintf("%s:%s", utils.AccessTokenRedis, accessToken)).Result()
	if err != nil {
		if err == redis.Nil {
			slog.Debug("token not found in redis, possibly expired or nonexistent")
			return false, nil
		} else {
			slog.Debug("error getting data from redis", "error", err)
			return false, eris.Wrap(err, "getting data from redis")
		}
	}

	slog.Debug("token found in redis", "client secret", data)

	return true, nil
}

func (s *BCAService) BalanceInquiry(ctx context.Context, payload *models.BCABalanceInquiry) (*models.BCAAccountBalance, error) {

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

	var obj models.BCAAccountBalance
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

func (s *BCAService) BillPresentment(ctx context.Context, payload *models.BCAVARequestPayload) (*models.VAResponsePayload, error) {
	var obj models.VAResponsePayload

	amount, err := s.GetVirtualAccountByInqueryRequestId(ctx, payload.InquiryRequestID)
	if err != nil {
		slog.Debug("error querying va_request", "error", err)

		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCABillInquiryResponseGeneralError.Data()
		return &obj, eris.Wrap(err, "Error Find VA")
	}

	if amount.Value != "0" {
		slog.Debug("va has been paid")

		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCABillInquiryResponseVAPaid.Data()
		return &obj, eris.Wrap(err, "va has been paid")
	}

	// if time.Now().After(expired) {
	// 	obj.ResponseCode = "4042419"
	// 	obj.ResponseMessage = "Invalid Bill/Virtual Account"
	// 	return obj, eris.Wrap(err, "VA has been expired")
	// }

	obj.VirtualAccountData = &models.VABCAResponseData{}
	slog.Debug("received data", "payload", payload)

	// Start of logic
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Debug("error beginning transaction", "error", err)
		tx.Rollback()

		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCABillInquiryResponseVANotFound.Data()
		return &obj, eris.Wrap(err, "begin tx")
	}

	statement := `
	SELECT partnerServiceId, customerNo, virtualAccountNo, virtualAccountName, totalAmountValue,
		   totalAmountCurrency, feeAmountValue, feeAmountCurrency
	FROM va_request 
	WHERE virtualAccountNo = ? AND paidAmountValue = '0'
	`
	if err := tx.QueryRowContext(ctx, statement, payload.VirtualAccountNo).Scan(
		&obj.VirtualAccountData.PartnerServiceID,
		&obj.VirtualAccountData.CustomerNo,
		&obj.VirtualAccountData.VirtualAccountNo,
		&obj.VirtualAccountData.VirtualAccountName,
		&obj.VirtualAccountData.TotalAmount.Value,
		&obj.VirtualAccountData.TotalAmount.Currency,
		&obj.VirtualAccountData.FeeAmount.Value,
		&obj.VirtualAccountData.FeeAmount.Currency,
	); err == sql.ErrNoRows {
		slog.Debug("va not found")
		tx.Rollback()

		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCABillInquiryResponseVANotFound.Data()
		return &obj, eris.Wrap(err, "va Not Found")
	} else if err != nil {
		slog.Debug("error querying va_request", "error", err)
		tx.Rollback()

		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCABillInquiryResponseGeneralError.Data()
		return &obj, eris.Wrap(err, "querying va_request")
	}

	statement = `
	UPDATE va_request SET inqueryRequestId = ? 
	WHERE virtualAccountNo = ? AND paidAmountValue = '0'
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

	if err = tx.Commit(); err != nil {
		slog.Debug("error committing transaction", "error", err)
		tx.Rollback()

		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCABillInquiryResponseGeneralError.Data()
		return &obj, eris.Wrap(err, "commit tx")
	}

	return &obj, nil
}

// func (s *BCAService) randomNumberString(length int) string {}
func (s *BCAService) BuildNumVA(ctx context.Context, idUser, idJenis int, partnerId string) (string, string) {

	partnerId += "0" + strconv.Itoa(idJenis)
	nIdU := strconv.Itoa(idUser)
	customerNo := ""
	for i := 0; i < 10-len(nIdU); i++ {
		customerNo += "0"
	}
	customerNo += nIdU
	return partnerId + customerNo, customerNo
}

func (s *BCAService) CreateVA(ctx context.Context, payload *models.CreateVAReq) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		tx.Rollback()
		return eris.Wrap(err, "begin tx")
	}

	partnerId := s.Config.BCAPartnerInformation.BCAPartnerId
	query := `
	INSERT INTO va_request (partnerServiceId, customerNo, virtualAccountNo, totalAmountValue, 
				   			virtualAccountName, id_user, owner_table)
	VALUES(?,?,?,?,?,?,?)
	`
	numVA, customerNo := s.BuildNumVA(ctx, payload.IdUser, payload.IdJenisUser, partnerId)
	if _, err = tx.ExecContext(ctx, query, partnerId, customerNo, numVA, payload.JumlahPembayaran,
		payload.NamaUser, payload.IdUser, payload.IdJenisUser); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "inserting new va into va_table")
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "commit tx")
	}

	return nil
}

func (s *BCAService) GetVirtualAccountByInqueryRequestId(ctx context.Context, inquiryRequestId string) (*models.Amount, error) {
	var amount models.Amount
	query := `
	SELECT totalAmountValue, totalAmountCurrency 
	FROM va_request
	WHERE inqueryRequestId = ?
	`
	if err := s.DB.QueryRowContext(ctx, query, inquiryRequestId).Scan(
		&amount.Value, &amount.Currency,
	); err != nil {
		slog.Debug("error querying va_request", "error", err)
		if err == sql.ErrNoRows {
			return &amount, nil
		} else {
			return nil, eris.Wrap(err, "querying va_request")
		}
	}
	return &amount, nil
}

func (s *BCAService) InquiryVA(ctx context.Context, payload *models.BCAInquiryRequest) (*models.BCAInquiryVAResponse, error) {
	var obj models.BCAInquiryVAResponse

	amount, err := s.GetVirtualAccountByInqueryRequestId(ctx, payload.PaymentRequestID)

	if err == sql.ErrNoRows {
		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseVANotFound.Data()
		return &obj, eris.Wrap(err, "querying va_table")
	} else if err != nil {
		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseGeneralError.Data()
		return &obj, eris.Wrap(err, "querying va_table")
	}

	if amount.Value > payload.PaidAmount.Value {
		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseGeneralError.Data()
		return &obj, nil
	}

	// Start of logic
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Debug("error beginning transaction", "error", err)
		tx.Rollback()

		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseGeneralError.Data()
		return &obj, eris.Wrap(err, "begin tx")
	}

	statement := `
	UPDATE va_request SET paidAmountValue = ?, 
						  paidAmountCurrency = ?, 
						  id_va_status = 2   
	WHERE inqueryRequestId = ?
	`
	_, err = tx.QueryContext(ctx, statement, payload.PaidAmount.Value, payload.PaidAmount.Currency,
		payload.PaymentRequestID)
	if err != nil {
		slog.Debug("error updating va_request", "error", err)
		tx.Rollback()

		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseGeneralError.Data()
		return &obj, eris.Wrap(err, "updating va_request")
	}

	obj.VirtualAccountData = &models.VirtualAccountDataInqury{}
	statement = `
	SELECT  partnerServiceId, customerNo, virtualAccountNo, virtualAccountName, totalAmountValue,
			totalAmountCurrency
	FROM va_request 
	WHERE inqueryRequestId = ?
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
		tx.Rollback()

		if err == sql.ErrNoRows {
			obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseVANotFound.Data()
			return &obj, nil
		} else {
			obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseGeneralError.Data()
			return &obj, eris.Wrap(err, "querying va_request")
		}
	}

	obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseSuccess.Data()

	obj.VirtualAccountData.PaymentRequestID = payload.PaymentRequestID
	obj.VirtualAccountData.PaidAmount.Value = payload.PaidAmount.Value
	obj.VirtualAccountData.PaidAmount.Currency = payload.PaidAmount.Currency

	if err = tx.Commit(); err != nil {
		slog.Debug("error committing transaction", "error", err)
		tx.Rollback()

		obj.HTTPStatusCode, obj.ResponseCode, obj.ResponseMessage = bca.BCAPaymentFlagResponseGeneralError.Data()
		return &obj, eris.Wrap(err, "commit tx")
	}

	return &obj, nil
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
		var obj models.BCAResponse

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
