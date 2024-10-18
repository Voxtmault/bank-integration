package bca_service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/rotisserie/eris"
	"github.com/voxtmault/bank-integration/config"
	"github.com/voxtmault/bank-integration/interfaces"
	"github.com/voxtmault/bank-integration/models"
)

type BCAService struct {
	Request   interfaces.Request
	Config    *config.BankingConfig
	Validator *validator.Validate

	AccessToken          string
	AccessTokenExpiresAt int64

	DB *sql.DB
}

var _ interfaces.SNAP = &BCAService{}

func NewBCAService(request interfaces.Request, config *config.BankingConfig, db *sql.DB) *BCAService {
	return &BCAService{
		Request: request,
		Config:  config,
		DB:      db,
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

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return eris.Wrap(err, "marshalling body")
	}

	slog.Debug("Getting Access Token from BCA", "URL", baseUrl)

	req, err := http.NewRequestWithContext(ctx, method, baseUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return eris.Wrap(err, "creating request")
	}

	slog.Debug("Building Request Header")
	// Before sending the request, customize the header
	if err = s.Request.AccessTokenRequestHeader(ctx, req, s.Config); err != nil {
		return eris.Wrap(err, "access token request header")
	}

	slog.Debug("Sending Request")
	// Send the request
	response, err := s.Request.RequestHandler(ctx, req)
	if err != nil {
		if response != "" {
			return eris.Wrap(eris.New(response), "sending request")
		} else {
			return eris.Wrap(err, "sending request")
		}
	}

	slog.Debug("Response from BCA", "Response: ", response)

	var atObj models.AccessTokenResponse
	if err = json.Unmarshal([]byte(response), &atObj); err != nil {
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
	// 3. Verify Asymmetric Signature
	// 4. Generate Access Token
	// 5. Save the Access Token along with client secret to redis

	// Parse the request body
	var body models.GrantType
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		return nil, eris.Wrap(err, "decoding request body")
	}

	if err := s.Validator.StructCtx(ctx, body); err != nil {
		return nil, eris.Wrap(err, "validating request body")
	}

	// Parse the request header
	timeStamp := request.Header.Get("X-TIMESTAMP")
	clientKey := request.Header.Get("X-CLIENT-KEY")
	// signature := request.Header.Get("X-SIGNATURE")

	// Validate parsed header
	if clientKey == "" {
		return &models.AccessTokenResponse{
			BCAResponse: models.BCAResponse{
				HTTPStatusCode:  http.StatusBadRequest,
				ResponseCode:    "4007301",
				ResponseMessage: "Invalid field format [clientId/clientSecret/grantType]",
			},
		}, nil
	} else if timeStamp == "" {
		return &models.AccessTokenResponse{
			BCAResponse: models.BCAResponse{
				HTTPStatusCode:  http.StatusBadRequest,
				ResponseCode:    "4007301",
				ResponseMessage: "Invalid field format [X-TIMESTAMP]",
			},
		}, nil
	}

	return nil, nil
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

	if err = s.Request.RequestHeader(ctx, request, s.Config, payload, s.Config.BCAURLEndpoints.BalanceInquiryURL, s.AccessToken); err != nil {
		return nil, eris.Wrap(err, "constructing request header")
	}

	response, err := s.Request.RequestHandler(ctx, request)
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

func (s *BCAService) BillPresentment(ctx context.Context, payload models.BCAVARequestPayload) (any, error) {
	var obj models.VAResponsePayload
	// log.Println(payload)
	statement := `SELECT 
    partnerServiceId,
    customerNo,
    virtualAccountNo,
    virtualAccountName,
    totalAmountValue,
    totalAmountCurrency,
    feeAmountValue,
    feeAmountCurrency
	FROM va_request WHERE virtualAccountNo='` + payload.VirtualAccountNo + `' AND paidAmountValue = '0';
`
	// log.Println(statement)
	err := s.DB.QueryRowContext(ctx, statement).Scan(&obj.VirtualAccountData.PartnerServiceID,
		&obj.VirtualAccountData.CustomerNo,
		&obj.VirtualAccountData.VirtualAccountNo,
		&obj.VirtualAccountData.VirtualAccountName,
		&obj.VirtualAccountData.TotalAmount.Value,
		&obj.VirtualAccountData.TotalAmount.Currency,
		&obj.VirtualAccountData.FeeAmount.Value,
		&obj.VirtualAccountData.FeeAmount.Currency)
	if err != nil {
		return nil, eris.Wrap(err, "querying va_table")
	}

	updateQuery := "UPDATE va_request SET inqueryRequestId = '" + payload.InquiryRequestID + "'  WHERE virtualAccountNo='" + payload.VirtualAccountNo + "' AND paidAmountValue = '0' "
	_, err = s.DB.QueryContext(ctx, updateQuery)
	if err != nil {
		return nil, eris.Wrap(err, "querying update va_table")
	}
	obj.ResponseCode = "2002400"
	obj.ResponseMessage = "Success"
	obj.VirtualAccountData.InquiryRequestID = payload.InquiryRequestID
	return obj, nil
}

// func (s *BCAService) CreateVA(ctx context.Context, payload models.CreateVAReq) (any, error) {
// 	query := "INSERT INTO va_request (totalAmountValue,virtualAccountNo,virtualAccountName)"
// 	err := s.DB.QueryRowContext(ctx, query).Scan(&amount.Value, &amount.Currency)
// 	if err != nil {
// 		return amount, eris.Wrap(err, "querying va_table")
// 	}
// 	return nil, nil
// }

func (s *BCAService) GetVirtualAccountByInqueryRequestId(ctx context.Context, inqueryRequestId string) (models.Amount, error) {
	var amount models.Amount
	query := "SELECT totalAmountValue,totalAmountCurrency FROM va_request"
	err := s.DB.QueryRowContext(ctx, query).Scan(&amount.Value, &amount.Currency)
	if err != nil {
		return amount, eris.Wrap(err, "querying va_table")
	}
	return amount, nil
}

func (s *BCAService) InquiryVA(ctx context.Context, payload models.BCAInquiryRequest) (any, error) {
	var obj models.BCAInquiryVAResponse
	amount, err := s.GetVirtualAccountByInqueryRequestId(ctx, payload.PaymentRequestID)
	if err != nil {
		obj.ResponseCode = "5002400"
		obj.ResponseMessage = "General Error"
		return obj, eris.Wrap(err, "querying va_table")
	}
	if amount.Value > payload.PaidAmount.Value {
		obj.ResponseCode = "5002400"
		obj.ResponseMessage = "General Error"
		return obj, nil
	}
	updateQuery := "UPDATE va_request SET paidAmountValue = '" + payload.PaidAmount.Value + "', paidAmountCurrency = '" + payload.PaidAmount.Currency + "', id_va_status = 2   WHERE inqueryRequestId = '" + payload.PaymentRequestID + "'"
	_, err = s.DB.QueryContext(ctx, updateQuery)
	if err != nil {
		obj.ResponseCode = "5002400"
		obj.ResponseMessage = "General Error"
		return obj, eris.Wrap(err, "querying va_table")
	}

	statement := `SELECT 
    partnerServiceId,
    customerNo,
    virtualAccountNo,
    virtualAccountName,
    totalAmountValue,
    totalAmountCurrency
	FROM va_request WHERE inqueryRequestId='` + payload.PaymentRequestID + `';
`
	log.Println(statement)
	rows, err := s.DB.QueryContext(ctx, statement)
	if err != nil {
		obj.ResponseCode = "5002400"
		obj.ResponseMessage = "General Error"
		return obj, eris.Wrap(err, "querying va_table")
	}
	for rows.Next() {
		err = rows.Scan(&obj.VirtualAccountData.PartnerServiceID,
			&obj.VirtualAccountData.CustomerNo,
			&obj.VirtualAccountData.VirtualAccountNo,
			&obj.VirtualAccountData.VirtualAccountName,
			&obj.VirtualAccountData.TotalAmount.Value,
			&obj.VirtualAccountData.TotalAmount.Currency)
		if err != nil {
			obj.ResponseCode = "5002400"
			obj.ResponseMessage = "General Error"
			return obj, eris.Wrap(err, "querying va_table")
		}
	}
	obj.ResponseCode = "2002400"
	obj.ResponseMessage = "Success"
	obj.VirtualAccountData.PaymentRequestID = payload.PaymentRequestID
	obj.VirtualAccountData.PaidAmount.Value = payload.PaidAmount.Value
	obj.VirtualAccountData.PaidAmount.Currency = payload.PaidAmount.Currency
	return obj, nil
}
