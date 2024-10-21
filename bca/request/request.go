package bca_request

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/rotisserie/eris"
	"github.com/voxtmault/bank-integration/config"
	"github.com/voxtmault/bank-integration/interfaces"
	"github.com/voxtmault/bank-integration/models"
	"github.com/voxtmault/bank-integration/utils"
)

type BCARequest struct {
	// Security is mainly used to generate signatures for request headers
	Security interfaces.Security
}

var _ interfaces.Request = &BCARequest{}

func NewBCARequest(security interfaces.Security) *BCARequest {
	return &BCARequest{
		Security: security,
	}
}

func (s *BCARequest) AccessTokenRequestHeader(ctx context.Context, request *http.Request, cfg *config.BankingConfig) error {

	// Checks for problematic configurations
	if err := utils.ValidateStruct(ctx, cfg.BCAConfig); err != nil {
		return eris.Wrap(err, "invalid BCA configuration")
	}

	timeStamp := time.Now().Format(time.RFC3339)

	slog.Debug("Creating asymetric signature")
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

func (s *BCARequest) RequestHeader(ctx context.Context, request *http.Request, cfg *config.BankingConfig, body any, relativeURL, accessToken string) error {
	timeStamp := time.Now().Format(time.RFC3339)

	slog.Debug("Creating symetric signature")
	// Create the signature
	signature, err := s.Security.CreateSymmetricSignature(ctx, &models.SymmetricSignatureRequirement{
		HTTPMethod:  request.Method,
		AccessToken: accessToken,
		Timestamp:   timeStamp,
		RequestBody: body,
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

func (s *BCARequest) VerifyAsymmetricSignature(ctx context.Context, timeStamp, clientKey, signature string) (bool, error) {
	return s.Security.VerifyAsymmetricSignature(ctx, timeStamp, clientKey, signature)
}

func (s *BCARequest) VerifySymmetricSignature(ctx context.Context, obj *models.SymmetricSignatureRequirement, clientSecret, signature string) (bool, error) {
	return s.Security.VerifySymmetricSignature(ctx, obj, clientSecret, signature)
}
