package bca_request

import (
	"context"
	"net/http"

	"github.com/voxtmault/bank-integration/interfaces"
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

func (s *BCAIngress) VerifyAsymmetricSignature(ctx context.Context, request *http.Request, clientSecret string) (bool, error) {
	// return s.Security.VerifyAsymmetricSignature(ctx, timeStamp, clientKey, signature)
	return false, nil
}

func (s *BCAIngress) VerifySymmetricSignature(ctx context.Context, request *http.Request) (bool, error) {
	// return s.Security.VerifySymmetricSignature(ctx, obj, clientSecret, signature)
	return false, nil
}
