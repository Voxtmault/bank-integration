package bank_integration_interfaces

import (
	"context"
	"net/http"

	biConfig "github.com/voxtmault/bank-integration/config"
	biModel "github.com/voxtmault/bank-integration/models"
	biStorage "github.com/voxtmault/bank-integration/storage"
)

// Deprecated: Use RequestEgress and RequestIngress instead.
type Request interface {
	// AccessTokenRequestHeader is ONLY used to set the headers for the request to get the access token.
	AccessTokenRequestHeader(ctx context.Context, request *http.Request, config *biConfig.BankingConfig) error

	// RequestHeader is used to set the headers for all other requests.
	RequestHeader(ctx context.Context, request *http.Request, cfg *biConfig.BankingConfig, body any, relativeURL, accessToken string) error

	RequestHandler(ctx context.Context, request *http.Request) (string, error)

	VerifyAsymmetricSignature(ctx context.Context, timeStamp, clientKey, signature string) (bool, error)
	VerifySymmetricSignature(ctx context.Context, obj *biModel.SymmetricSignatureRequirement, clientSecret, signature string) (bool, error)
}

// RequestEgress is an interface that defines the methods that are used to send requests to banks.
type RequestEgress interface {
	// GenerateAccessRequestHeader is ONLY used to generate the headers for the request to get the access token.
	GenerateAccessRequestHeader(ctx context.Context, request *http.Request, cfg *biConfig.BankingConfig) error

	// GenerateGeneralRequestHeader is used to generate the headers for all other requests.
	GenerateGeneralRequestHeader(ctx context.Context, request *http.Request, cfg *biConfig.BankingConfig, body any, relativeURL, accessToken string) error
}

// RequestIngress is an interface that defines the methods that are used to receive requests from banks.
type RequestIngress interface {
	// VerifyAsymmetricSignature verifies the request headers ONLY for access-token related http requests.
	VerifyAsymmetricSignature(ctx context.Context, request *http.Request, redis *biStorage.RedisInstance) (bool, *biModel.BCAResponse, string)

	// VerifySymmetricSignature verifies the request headers for non access-token related http requests.
	VerifySymmetricSignature(ctx context.Context, request *http.Request, redis *biStorage.RedisInstance) (bool, *biModel.BCAResponse)

	ValidateAccessToken(ctx context.Context, redis *biStorage.RedisInstance, accessToken string) (string, error)
}

type Security interface {
	//CreateAsymmetricSignature returns a base64 encoded signature. Based on SHA256-RSA algorithm.
	CreateAsymmetricSignature(ctx context.Context, timeStamp string) (string, error)

	// VerifyAsymmetricSignature verifies the request headers ONLY for access-token related http requests.
	// It will compares the received HMAC with the calculated HMAC based on the received public key from banks.
	//
	// This function will return a boolean value signifying the results of comparison and an error regarding the internal process
	VerifyAsymmetricSignature(ctx context.Context, timeStamp, clientKey, signature string) (bool, error)

	// CreateSymmetricSignature returns a base64 encoded signature. Based on SHA512-HMAC algorithm.
	CreateSymmetricSignature(ctx context.Context, obj *biModel.SymmetricSignatureRequirement) (string, error)

	// VerifySymmetricSignature verifies the request headers for non access-token related http requests.
	// It will compares the received HMAC with the calculated HMAC based on the received public key from banks.
	//
	// This function will return a boolean value signifying the results of comparison and an error regarding the internal process
	VerifySymmetricSignature(ctx context.Context, obj *biModel.SymmetricSignatureRequirement, clientSecret, signature string) (bool, error)
}

type SNAP interface {
	// Generally used to get access token from banks, but can also be used to renew existing tokens.
	GetAccessToken(ctx context.Context) error

	// GenerateAccessToken is used to generate the access token as a response form banks trying to authenticate
	// with our wallet API
	GenerateAccessToken(ctx context.Context, request *http.Request) (*biModel.AccessTokenResponse, error)

	// Used to get the information regarding the account balance and other informations.
	BalanceInquiry(ctx context.Context, payload *biModel.BCABalanceInquiry) (*biModel.BCAAccountBalance, error)

	BillPresentment(ctx context.Context, data []byte) (*biModel.VAResponsePayload, error)
	VerifyAdditionalBillPresentmentRequiredHeader(ctx context.Context, request *http.Request) (*biModel.BCAResponse, []byte, error)

	InquiryVA(ctx context.Context, data []byte) (*biModel.BCAInquiryVAResponse, error)
	VerifyAdditionalInquiryVARequiredHeader(ctx context.Context, request *http.Request) (*biModel.BCAResponse, []byte, error)

	// CreateVA is used in tandem with order creation when VA Payment is chosen as the payment method.
	CreateVA(ctx context.Context, payload *biModel.CreateVAReq) error

	// Middleware is used to verify the ingress request from banks. It will return the auth response and any payload inside the request body
	Middleware(ctx context.Context, request *http.Request) (*biModel.BCAResponse, []byte, error)
}

type Management interface {
	GetAuthenticatedBanks(ctx context.Context) ([]*biModel.AuthenticatedBank, error)

	RegisterBank(ctx context.Context, bankName string) (*biModel.BankClientCredential, error)

	UpdateRegisteredBank(ctx context.Context) error

	RevokeRegisteredBank(ctx context.Context) error
}
