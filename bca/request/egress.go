package bca_request

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/rotisserie/eris"
	"github.com/voxtmault/bank-integration/config"
	"github.com/voxtmault/bank-integration/interfaces"
	"github.com/voxtmault/bank-integration/models"
	"github.com/voxtmault/bank-integration/utils"
)

type BCAEgress struct {
	// Security is mainly used to generate signatures for request headers
	Security interfaces.Security
}

var _ interfaces.RequestEgress = &BCAEgress{}

func NewBCAEgress(security interfaces.Security) *BCAEgress {
	return &BCAEgress{
		Security: security,
	}
}

func (s *BCAEgress) GenerateAccessRequestHeader(ctx context.Context, request *http.Request, cfg *config.BankingConfig) error {

	// Checks for problematic configurations
	if err := utils.ValidateStruct(ctx, cfg.BCAConfig); err != nil {
		return eris.Wrap(err, "invalid BCA configuration")
	}

	timeStamp := time.Now().Format(time.RFC3339)

	slog.Debug("Creating asymetric signature")
	signature, err := s.Security.CreateAsymmetricSignature(ctx, timeStamp)
	if err != nil {
		slog.Debug("error creating signature", "error", err)
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

func (s *BCAEgress) GenerateGeneralRequestHeader(ctx context.Context, request *http.Request, cfg *config.BankingConfig, body any, relativeURL, accessToken string) error {
	timeStamp := time.Now().Format(time.RFC3339)

	slog.Debug("Creating symetric signature")
	signature, err := s.Security.CreateSymmetricSignature(ctx, &models.SymmetricSignatureRequirement{
		HTTPMethod:  request.Method,
		AccessToken: accessToken,
		Timestamp:   timeStamp,
		RequestBody: body,
		RelativeURL: relativeURL,
	})
	if err != nil {
		slog.Debug("error creating signature", "error", err)
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
