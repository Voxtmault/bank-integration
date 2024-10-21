package interfaces

import (
	"context"
	"net/http"

	"github.com/voxtmault/bank-integration/config"
	models "github.com/voxtmault/bank-integration/models"
	"github.com/voxtmault/bank-integration/storage"
)

// Deprecated: Use RequestEgress and RequestIngress instead.
type Request interface {
	// AccessTokenRequestHeader is ONLY used to set the headers for the request to get the access token.
	AccessTokenRequestHeader(ctx context.Context, request *http.Request, config *config.BankingConfig) error

	// RequestHeader is used to set the headers for all other requests.
	RequestHeader(ctx context.Context, request *http.Request, cfg *config.BankingConfig, body any, relativeURL, accessToken string) error

	RequestHandler(ctx context.Context, request *http.Request) (string, error)

	VerifyAsymmetricSignature(ctx context.Context, timeStamp, clientKey, signature string) (bool, error)
	VerifySymmetricSignature(ctx context.Context, obj *models.SymmetricSignatureRequirement, clientSecret, signature string) (bool, error)
}

// RequestEgress is an interface that defines the methods that are used to send requests to banks.
type RequestEgress interface {
	// GenerateAccessRequestHeader is ONLY used to generate the headers for the request to get the access token.
	GenerateAccessRequestHeader(ctx context.Context, request *http.Request, cfg *config.BankingConfig) error

	// GenerateGeneralRequestHeader is used to generate the headers for all other requests.
	GenerateGeneralRequestHeader(ctx context.Context, request *http.Request, cfg *config.BankingConfig, body any, relativeURL, accessToken string) error
}

// RequestIngress is an interface that defines the methods that are used to receive requests from banks.
type RequestIngress interface {
	// VerifyAsymmetricSignature verifies the request headers ONLY for access-token related http requests.
	VerifyAsymmetricSignature(ctx context.Context, request *http.Request, redis *storage.RedisInstance) (bool, *models.BCAResponse, string)

	// VerifySymmetricSignature verifies the request headers for non access-token related http requests.
	VerifySymmetricSignature(ctx context.Context, request *http.Request, redis *storage.RedisInstance) (bool, *models.BCAResponse)

	ValidateAccessToken(ctx context.Context, redis *storage.RedisInstance, accessToken string) (string, error)
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
	CreateSymmetricSignature(ctx context.Context, obj *models.SymmetricSignatureRequirement) (string, error)

	// VerifySymmetricSignature verifies the request headers for non access-token related http requests.
	// It will compares the received HMAC with the calculated HMAC based on the received public key from banks.
	//
	// This function will return a boolean value signifying the results of comparison and an error regarding the internal process
	VerifySymmetricSignature(ctx context.Context, obj *models.SymmetricSignatureRequirement, clientSecret, signature string) (bool, error)
}

type SNAP interface {
	// Generally used to get access token from banks, but can also be used to renew existing tokens.
	GetAccessToken(ctx context.Context) error

	// GenerateAccessToken is used to generate the access token as a response form banks trying to authenticate
	// with our wallet API
	GenerateAccessToken(ctx context.Context, request *http.Request) (*models.AccessTokenResponse, error)

	// Used to get the information regarding the account balance and other informations.
	BalanceInquiry(ctx context.Context, payload *models.BCABalanceInquiry) (*models.BCAAccountBalance, error)

	BillPresentment(ctx context.Context, request *http.Request) (*models.VAResponsePayload, error)

	// CreateVA is used in tandem with order creation when VA Payment is chosen as the payment method.
	CreateVA(ctx context.Context, payload *models.CreateVAReq) error
}
