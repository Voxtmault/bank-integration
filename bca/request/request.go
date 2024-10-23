package bca_request

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/rotisserie/eris"
	biConfig "github.com/voxtmault/bank-integration/config"
	biInterfaces "github.com/voxtmault/bank-integration/interfaces"
	biModels "github.com/voxtmault/bank-integration/models"
	biUtil "github.com/voxtmault/bank-integration/utils"
)

type BCARequest struct {
	// Security is mainly used to generate signatures for request headers
	Security biInterfaces.Security
}

var _ biInterfaces.Request = &BCARequest{}

func NewBCARequest(security biInterfaces.Security) *BCARequest {
	return &BCARequest{
		Security: security,
	}
}

func (s *BCARequest) AccessTokenRequestHeader(ctx context.Context, request *http.Request, cfg *biConfig.BankingConfig) error {

	// Checks for problematic configurations
	if err := biUtil.ValidateStruct(ctx, cfg.BCAConfig); err != nil {
		return eris.Wrap(err, "invalid BCA configuration")
	}

	timeStamp := time.Now().Format(time.RFC3339)

	slog.Debug("Creating asymmetric signature")
	// Create the signature
	signature, err := s.Security.CreateAsymmetricSignature(ctx, timeStamp)
	if err != nil {
		return eris.Wrap(err, "creating signature")
	}

	// Checks if the caller has set a content-type already
	if request.Header.Get("Content-Type") == "" {
		// Default to application/json if not set
		request.Header.Set("Content-Type", "application/json")
	}

	slog.Debug("Request Header Debug", "Timestamp: ", timeStamp)
	slog.Debug("Request Header Debug", "Client ID: ", cfg.BCAConfig.ClientID)
	slog.Debug("Request Header Debug", "Signature: ", signature)

	// Add custom headers required by BCA
	request.Header.Set("X-TIMESTAMP", timeStamp)
	request.Header.Set("X-CLIENT-KEY", cfg.BCAConfig.ClientID)
	request.Header.Set("X-SIGNATURE", signature)
	request.Header.Set("Host", cfg.AppHost)

	return nil
}

func (s *BCARequest) RequestHeader(ctx context.Context, request *http.Request, cfg *biConfig.BankingConfig, body any, relativeURL, accessToken string) error {
	timeStamp := time.Now().Format(time.RFC3339)

	// Read the body and convert to string
	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		return eris.Wrap(err, "reading request body")
	}
	request.Body.Close()                                    // Close the body to prevent resource leaks
	request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset the body for further use

	slog.Debug("Creating symetric signature")
	// Create the signature
	signature, err := s.Security.CreateSymmetricSignature(ctx, &biModels.SymmetricSignatureRequirement{
		HTTPMethod:  request.Method,
		AccessToken: accessToken,
		Timestamp:   timeStamp,
		RequestBody: bodyBytes,
		RelativeURL: relativeURL,
	})
	if err != nil {
		return eris.Wrap(err, "creating signature")
	}

	// Checks if the caller has set a content-type already
	if request.Header.Get("Content-Type") == "" {
		// Default to application/json if not set
		request.Header.Set("Content-Type", "application/json")
	}

	// Add custom headers required by BCA
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("X-TIMESTAMP", timeStamp)
	request.Header.Set("X-CLIENT-KEY", cfg.BCAConfig.ClientID)
	request.Header.Set("X-SIGNATURE", signature)
	request.Header.Set("ORIGIN", cfg.AppHost)
	request.Header.Set("X-EXTERNAL-ID", "")

	return nil
}

func (s *BCARequest) RequestHandler(ctx context.Context, request *http.Request) (string, error) {

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

func (s *BCARequest) VerifyAsymmetricSignature(ctx context.Context, timeStamp, clientKey, signature string) (bool, error) {
	return s.Security.VerifyAsymmetricSignature(ctx, timeStamp, clientKey, signature)
}

func (s *BCARequest) VerifySymmetricSignature(ctx context.Context, obj *biModels.SymmetricSignatureRequirement, clientSecret, signature string) (bool, error) {
	return s.Security.VerifySymmetricSignature(ctx, obj, clientSecret, signature)
}
