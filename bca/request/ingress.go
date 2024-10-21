package bca_request

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/voxtmault/bank-integration/bca"
	"github.com/voxtmault/bank-integration/interfaces"
	"github.com/voxtmault/bank-integration/models"
	"github.com/voxtmault/bank-integration/storage"
	"github.com/voxtmault/bank-integration/utils"
)

type BCAIngress struct {
	// Security is mainly used to generate signatures for request headers
	Security interfaces.Security
}

var _ interfaces.RequestIngress = &BCAIngress{}

func NewBCAIngress(security interfaces.Security) *BCAIngress {
	return &BCAIngress{
		Security: security,
	}
}

func (s *BCAIngress) VerifyAsymmetricSignature(ctx context.Context, request *http.Request, redis *storage.RedisInstance) (bool, *models.BCAResponse, string) {

	// Parse the request header
	timeStamp := request.Header.Get("X-TIMESTAMP")
	clientKey := request.Header.Get("X-CLIENT-KEY")
	signature := request.Header.Get("X-SIGNATURE")

	// Validate parsed header
	if clientKey == "" {
		slog.Debug("clientKey is empty")

		response := bca.BCAAUthInvalidMandatoryField
		response.ResponseMessage = response.ResponseMessage + "[X-CLIENT-KEY]"

		return false, &response, ""
	} else if timeStamp == "" {
		slog.Debug("timeStamp is empty")

		response := bca.BCAAUthInvalidMandatoryField
		response.ResponseMessage = response.ResponseMessage + "[X-TIMESTAMP]"

		return false, &response, ""
	}

	// Validate the timestamp format
	if _, err := time.Parse(time.RFC3339, timeStamp); err != nil {
		slog.Debug("invalid timestamp format")
		return false, &bca.BCAAuthInvalidFieldFormatTimestamp, ""
	}

	// Retrieve the client secret from redis
	clientSecret, err := redis.GetIndividualValueRedisHash(ctx, utils.ClientCredentialsRedis, clientKey)
	if err != nil {
		slog.Debug("error getting client secret", "error", err)
		return false, &bca.BCAAuthGeneralError, ""
	}

	if clientSecret == "" {
		slog.Debug("clientId is not registered")
		return false, &bca.BCAAuthUnauthorizedUnknownClient, ""
	}

	result, err := s.Security.VerifyAsymmetricSignature(ctx, timeStamp, clientKey, signature)
	if err != nil {
		slog.Debug("error verifying signature", "error", err)

		return false, &bca.BCAAuthGeneralError, ""
	}

	return result, nil, clientSecret
}

func (s *BCAIngress) VerifySymmetricSignature(ctx context.Context, request *http.Request) (bool, *models.BCAResponse) {
	// return s.Security.VerifySymmetricSignature(ctx, obj, clientSecret, signature)
	return false, nil
}
