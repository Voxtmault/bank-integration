package bca_request

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/rotisserie/eris"
	biConfig "github.com/voxtmault/bank-integration/config"
	biInterfaces "github.com/voxtmault/bank-integration/interfaces"
	biModels "github.com/voxtmault/bank-integration/models"
)

type BCAEgress struct {
	bankConfig     *biConfig.BankConfig
	internalConfig *biConfig.InternalConfig

	// Security is mainly used to generate signatures for request headers
	Security biInterfaces.Security
}

var _ biInterfaces.RequestEgress = &BCAEgress{}

func NewBCAEgress(security biInterfaces.Security, bCfg *biConfig.BankConfig, cfg *biConfig.InternalConfig) *BCAEgress {
	return &BCAEgress{
		bankConfig:     bCfg,
		internalConfig: cfg,
		Security:       security,
	}
}

func (s *BCAEgress) GenerateAccessRequestHeader(ctx context.Context, request *http.Request) error {
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
	slog.Debug("Request Header Debug", "Client ID: ", s.bankConfig.BankCredential.ClientID)
	slog.Debug("Request Header Debug", "Signature: ", signature)

	// Add custom headers required by BCA
	request.Header.Set("X-TIMESTAMP", timeStamp)
	request.Header.Set("X-CLIENT-KEY", s.bankConfig.BankCredential.ClientID)
	request.Header.Set("X-SIGNATURE", signature)
	request.Header.Set("Host", s.internalConfig.AppHost)

	return nil
}

func (s *BCAEgress) GenerateGeneralRequestHeader(ctx context.Context, request *http.Request, relativeURL, accessToken string) error {
	timeStamp := time.Now().Format(time.RFC3339)

	// Read the body and convert to string
	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		return eris.Wrap(err, "reading request body")
	}
	request.Body.Close()                                    // Close the body to prevent resource leaks
	request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset the body for further use

	slog.Debug("Creating symetric signature")
	signature, err := s.Security.CreateSymmetricSignature(ctx, &biModels.SymmetricSignatureRequirement{
		HTTPMethod:  request.Method,
		AccessToken: accessToken,
		Timestamp:   timeStamp,
		RequestBody: bodyBytes,
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
	request.Header.Set("X-CLIENT-KEY", s.bankConfig.BankCredential.ClientID)
	request.Header.Set("X-SIGNATURE", signature)
	request.Header.Set("ORIGIN", s.internalConfig.AppHost)
	request.Header.Set("X-EXTERNAL-ID", strconv.Itoa(int(time.Now().Unix())))

	return nil
}
