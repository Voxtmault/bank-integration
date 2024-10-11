package bca

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/rotisserie/eris"
	"github.com/voxtmault/bank-integration/config"
	"github.com/voxtmault/bank-integration/interfaces"
	"github.com/voxtmault/bank-integration/models"
)

type BCARequest struct {
	// Security is mainly used to generate signatures for request headers
	Security  interfaces.Security
	Validator *validator.Validate
}

var _ interfaces.Request = &BCARequest{}

func NewBCARequest(security interfaces.Security) *BCARequest {
	return &BCARequest{
		Security: security,
	}
}

func (s *BCARequest) AccessTokenRequestHeader(ctx context.Context, request *http.Request) error {
	cfg := config.GetConfig()

	// Checks for problematic configurations
	if err := s.Validator.Struct(cfg.BCAConfig); err != nil {
		return eris.Wrap(err, "invalid BCA configuration")
	}

	timeStamp := time.Now().Format(time.RFC3339)

	// Create the signature
	signature, err := s.Security.CreateAsymetricSignature(ctx, timeStamp)
	if err != nil {
		return eris.Wrap(err, "creating signature")
	}

	// Checks if the caller has set a content-type already
	if request.Header.Get("Content-Type") == "" {
		// Default to application/json if not set
		request.Header.Set("Content-Type", "application/json")
	}

	log.Println("Timestamp: ", timeStamp)
	log.Println("Client ID: ", cfg.BCAConfig.ClientID)
	log.Println("Signature: ", signature)

	// Add custom headers required by BCA
	request.Header.Set("X-TIMESTAMP", timeStamp)
	request.Header.Set("X-CLIENT-KEY", cfg.BCAConfig.ClientID)
	request.Header.Set("X-SIGNATURE", signature)
	request.Header.Set("Host", cfg.AppHost)

	return nil
}

func (s *BCARequest) RequestHeader(ctx context.Context, request *http.Request, body any, relativeURL, accessToken string) error {

	cfg := config.GetConfig()

	timeStamp := time.Now().Format(time.RFC3339)

	// Create the signature
	signature, err := s.Security.CreateSymetricSignature(ctx, &models.SymetricSignatureRequirement{
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
