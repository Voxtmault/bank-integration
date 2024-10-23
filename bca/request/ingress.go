package bca_request

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rotisserie/eris"
	"github.com/voxtmault/bank-integration/bca"
	biInterfaces "github.com/voxtmault/bank-integration/interfaces"
	biModels "github.com/voxtmault/bank-integration/models"
	biStorage "github.com/voxtmault/bank-integration/storage"
	biUtil "github.com/voxtmault/bank-integration/utils"
)

type BCAIngress struct {
	// Security is mainly used to generate signatures for request headers
	Security biInterfaces.Security
}

var _ biInterfaces.RequestIngress = &BCAIngress{}

func NewBCAIngress(security biInterfaces.Security) *BCAIngress {
	return &BCAIngress{
		Security: security,
	}
}

func (s *BCAIngress) VerifyAsymmetricSignature(ctx context.Context, request *http.Request, redis *biStorage.RedisInstance) (bool, *biModels.BCAResponse, string) {

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
	} else if signature == "" {
		slog.Debug("signature is empty")

		response := bca.BCAAUthInvalidMandatoryField
		response.ResponseMessage = response.ResponseMessage + "[X-SIGNATURE]"
		return false, &response, ""
	}

	// Validate the timestamp format
	if _, err := time.Parse(time.RFC3339, timeStamp); err != nil {
		slog.Debug("invalid timestamp format")
		return false, &bca.BCAAuthInvalidFieldFormatTimestamp, ""
	}

	// Retrieve the client secret from redis
	clientSecret, err := redis.GetIndividualValueRedisHash(ctx, biUtil.ClientCredentialsRedis, clientKey)
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

		if eris.Cause(err).Error() == "verification error" {
			return false, &bca.BCAAuthUnauthorizedSignature, ""
		}

		return false, &bca.BCAAuthGeneralError, ""
	}

	return result, nil, clientSecret
}

func (s *BCAIngress) VerifySymmetricSignature(ctx context.Context, request *http.Request, redis *biStorage.RedisInstance, body any) (bool, *biModels.BCAResponse) {

	var obj biModels.SymmetricSignatureRequirement

	// Validate External ID
	if request.Header.Get("X-EXTERNAL-ID") == "" {
		slog.Debug("externalId is empty")

		response := bca.BCAAUthInvalidMandatoryField
		response.ResponseMessage = response.ResponseMessage + "[X-EXTERNAL-ID]"

		return false, &response
	}

	externalUnique, err := s.ValidateUniqueExternalID(ctx, redis, request.Header.Get("X-EXTERNAL-ID"))
	if err != nil {
		slog.Debug("error validating externalId", "error", err)

		if eris.Cause(err).Error() == "invalid field format" {
			response := bca.BCAAuthInvalidFieldFormatClient

			response.ResponseMessage = "Invalid Field Format [X-EXTERNAL-ID]"
			return false, &response
		}

		return false, &bca.BCAAuthGeneralError
	}

	if !externalUnique {
		slog.Debug("externalId is not unique")
		return false, &bca.BCAAuthConflict
	}

	// Parse the request header
	obj.Timestamp = request.Header.Get("X-TIMESTAMP")
	obj.AccessToken = request.Header.Get("Authorization")
	signature := request.Header.Get("X-SIGNATURE")

	// Validate parsed header
	if obj.Timestamp == "" {
		slog.Debug("timeStamp is empty")

		response := bca.BCAAUthInvalidMandatoryField
		response.ResponseMessage = response.ResponseMessage + "[X-TIMESTAMP]"

		return false, &response
	} else if obj.AccessToken == "" {
		slog.Debug("accessToken is empty")

		response := bca.BCAAUthInvalidMandatoryField
		response.ResponseMessage = response.ResponseMessage + "[Authorization]"

		return false, &response
	} else if signature == "" {
		slog.Debug("signature is empty")

		response := bca.BCAAUthInvalidMandatoryField
		response.ResponseMessage = response.ResponseMessage + "[X-SIGNATURE]"

		return false, &response
	}

	// Validate the timestamp format
	if _, err := time.Parse(time.RFC3339, obj.Timestamp); err != nil {
		slog.Debug("invalid timestamp format")
		return false, &bca.BCAAuthInvalidFieldFormatTimestamp
	}

	// Retrieve the client secret from redis
	clientSecret, err := s.ValidateAccessToken(ctx, redis, obj.AccessToken)
	if err != nil {
		slog.Debug("error getting client secret", "error", err)
		return false, &bca.BCAAuthGeneralError
	}

	if clientSecret == "" {
		slog.Debug("accessToken is not registered")
		return false, &bca.BCAAuthUnauthorizedUnknownClient
	}

	obj.HTTPMethod = request.Method
	obj.RelativeURL = request.URL.Path

	obj.RequestBody = body

	result, err := s.Security.VerifySymmetricSignature(ctx, &obj, clientSecret, signature)
	if err != nil {
		slog.Debug("error verifying signature", "error", err)

		if eris.Cause(err).Error() == "verification error" {
			return false, &bca.BCAAuthUnauthorizedSignature
		}

		return false, &bca.BCAAuthGeneralError
	}

	return result, nil
}

func (s *BCAIngress) ValidateAccessToken(ctx context.Context, rdb *biStorage.RedisInstance, accessToken string) (string, error) {
	// Logic
	// 1. Get the access token from Redis
	// 2. If redis return nil then return false to the caller
	// 3. if redis returns a value then return true to the caller

	data, err := rdb.RDB.Get(ctx, fmt.Sprintf("%s:%s", biUtil.AccessTokenRedis, accessToken)).Result()
	if err != nil {
		if err == redis.Nil {
			slog.Debug("token not found in redis, possibly expired or nonexistent")
			return "", nil
		} else {
			slog.Debug("error getting data from redis", "error", err)
			return "", eris.Wrap(err, "getting data from redis")
		}
	}

	slog.Debug("token found in redis", "client secret", data)

	return data, nil
}

func (s *BCAIngress) ValidateUniqueExternalID(ctx context.Context, rdb *biStorage.RedisInstance, externalId string) (bool, error) {
	// Get only the first 36 characters of the externalId
	if len(externalId) > 36 {
		externalId = externalId[:36]
	}

	// Checks if the external ID is numeric
	if _, err := strconv.ParseInt(externalId, 10, 64); err != nil {
		slog.Debug("externalId is not numeric", "externalId", externalId)
		return false, eris.New("invalid field format")
	}

	data, err := rdb.RDB.SMembers(ctx, fmt.Sprintf("%s:%s", biUtil.UniqueExternalIDRedis, biUtil.BankCodeBCA)).Result()
	if err != nil {
		slog.Debug("error getting data from redis", "error", err)
		return false, eris.Wrap(err, "getting data from redis")
	}

	for _, item := range data {
		if externalId == item {
			slog.Debug("externalId already exists", "externalId", externalId, "cached id", item)
			return false, nil
		}
	}

	// If it's a new one then add to the hash sets
	if err = rdb.RDB.SAdd(ctx, fmt.Sprintf("%s:%s", biUtil.UniqueExternalIDRedis, biUtil.BankCodeBCA), externalId).Err(); err != nil {
		slog.Debug("error saving data to redis cache", "error", err)
		return false, eris.Wrap(err, "saving data to redis cache")
	}

	return true, nil
}
