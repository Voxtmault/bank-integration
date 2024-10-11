package interfaces

import (
	"context"
	"net/http"

	models "github.com/voxtmault/bank-integration/models"
)

type Request interface {
	// AccessTokenRequestHeader is ONLY used to set the headers for the request to get the access token.
	AccessTokenRequestHeader(ctx context.Context, request *http.Request) error

	// RequestHeader is used to set the headers for all other requests.
	RequestHeader(ctx context.Context, request *http.Request, body any, relativeURL, accessToken string) error

	RequestHandler(ctx context.Context, request *http.Request) (string, error)
}

type Security interface {
	//CreateAsymetricSignature returns a base64 encoded signature. Based on SHA256-RSA algorithm.
	CreateAsymetricSignature(ctx context.Context, timeStamp string) (string, error)

	// CreateSymetricSignature returns a base64 encoded signature. Based on SHA512-HMAC algorithm.
	CreateSymetricSignature(ctx context.Context, obj *models.SymetricSignatureRequirement) (string, error)
}

type SNAP interface {
	// Generally used to get access token, but can also be used to renew existing tokens.
	GetAccessToken(ctx context.Context) error

	// Used to get the information regarding the account balance and other informations.
	BalanceInquiry(ctx context.Context, payload *models.BCABalanceInquiry) (*models.BCAAccountBalance, error)
}
