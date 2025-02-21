package bank_integration_interfaces

import (
	"context"
	"net/http"

	biModel "github.com/voxtmault/bank-integration/models"
	biStorage "github.com/voxtmault/bank-integration/storage"
	"github.com/voxtmault/bank-integration/watcher"
)

// RequestEgress is an interface that defines the methods that are used to send requests to banks.
type RequestEgress interface {
	// GenerateAccessRequestHeader is ONLY used to generate the headers for the request to get the access token.
	GenerateAccessRequestHeader(ctx context.Context, request *http.Request) error

	// GenerateGeneralRequestHeader is used to generate the headers for all other requests.
	GenerateGeneralRequestHeader(ctx context.Context, request *http.Request, relativeURL, accessToken string) error
}

// RequestIngress is an interface that defines the methods that are used to receive requests from banks.
type RequestIngress interface {
	// VerifyAsymmetricSignature verifies the request headers ONLY for access-token related http requests.
	VerifyAsymmetricSignature(ctx context.Context, request *http.Request, redis *biStorage.RedisInstance) (bool, *biModel.BCAResponse, string)

	// VerifySymmetricSignature verifies the request headers for non access-token related http requests.
	VerifySymmetricSignature(ctx context.Context, request *http.Request, redis *biStorage.RedisInstance, payload []byte) (bool, *biModel.BCAResponse)

	ValidateAccessToken(ctx context.Context, redis *biStorage.RedisInstance, accessToken string) (string, error)

	ValidateUniqueExternalID(ctx context.Context, rdb *biStorage.RedisInstance, externalId string) (bool, error)
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
	GetVAPaymentStatus(ctx context.Context, vaNum string) (*biModel.VAPaymentStatusResponse, error)
	CreateVAV2(ctx context.Context, payload *biModel.CreatePaymentVARequestV2) error

	// Ingress

	// Called by each partnered bank to generate / get access-token to be used for the next requests
	GenerateAccessToken(ctx context.Context, request *http.Request) (*biModel.AccessTokenResponse, error)
	// Called by each partnered bank to get the virtual account billing information from our system
	BillPresentment(ctx context.Context, request *http.Request) (*biModel.VAResponsePayload, error)
	// Called by each partnered bank to update the billing status of a virtual account in our system
	InquiryVA(ctx context.Context, request *http.Request) (*biModel.BCAInquiryVAResponse, error)

	// Egress

	GetAccessToken(ctx context.Context) error
	BalanceInquiry(ctx context.Context) (*biModel.BankAccountBalance, error)
	BankStatement(ctx context.Context, fromDateTime, toDateTime string) (*biModel.BankStatement, error)
	TransferIntraBank(ctx context.Context, payload *biModel.BCATransferIntraBankReq) (*biModel.BankTransferResponse, error)
	TransferInterBank(ctx context.Context, payload *biModel.BCATransferInterBankRequest) (*biModel.BankTransferResponse, error)

	// GetWatchedTransaction returns list of transaction that is currently being watched by the bank service implementation
	GetWatchedTransaction(ctx context.Context) []*biModel.TransactionWatcherPublic
	GetWatcher() *watcher.TransactionWatcher
	GetAllVAWaitingPayment(ctx context.Context) error
}

type Management interface {
	// Partnered Banks

	GetPartneredBanks(ctx context.Context) ([]*biModel.PartneredBank, error)
	RegisterPartneredBank(ctx context.Context, obj *biModel.PartneredBankAdd) (*biModel.PartneredBank, error)
	UpdatePartneredBanks(ctx context.Context, obj *biModel.PartneredBank) error
	DeletePartneredBank(ctx context.Context, idBank uint) error

	// Partnered Banks Integrated Features

	GetBankIntegratedFeatures(ctx context.Context, idBank uint) ([]*biModel.IntegratedFeature, error)
	EditBankIntegratedFeatures(ctx context.Context, arr []*biModel.IntegratedFeatureAdd) error
	DeleteBankIntegratedFeature(ctx context.Context, idBank, idBIF uint) error

	// Partnered Banks Payment Method

	GetBankPaymentMethods(ctx context.Context, idBank uint) ([]*biModel.PaymentMethod, error)
	EditBankPamentMethods(ctx context.Context, arr []*biModel.PaymentMethodAdd) error
	DeleteBankPaymentMethod(ctx context.Context, idBank, idPM uint) error

	// Partnered Banks Client Credentials

	GetBankClientCredentials(ctx context.Context, idBank uint) ([]*biModel.BankClientCredential, error)
	AddBankClientCredential(ctx context.Context, idBank uint, note string) (*biModel.BankClientCredential, error)
	EditBankClientCredential(ctx context.Context, obj *biModel.BankClientCredential) error
	DeleteBankClientCredential(ctx context.Context, idBank, idCC uint) error
}

type Internal interface {
	GetOrderVAInformation(ctx context.Context, idOrder uint) (*biModel.InternalVAInformation, error)
	GetTopUpVAInformation(ctx context.Context, trxId uint) (*biModel.InternalVAInformation, error)
}
