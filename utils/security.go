package utils

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log/slog"

	"github.com/rotisserie/eris"
)

// This file is used for general security like generating access tokens and such
// The reason i chose this way is becaused each bank does not specify how to generate the access tokens
// they're going to use, so i made a general way to generate access tokens.

// Unless the bank specifies how to generate the access token, this is the way to go.

type GeneralSecurity struct {
}

// GenerateAccessToken will return a 64 character long string as an access token
func (s *GeneralSecurity) GenerateAccessToken(ctx context.Context) (string, error) {

	bytes := make([]byte, 64)
	if _, err := rand.Read(bytes); err != nil {
		slog.Debug("error generating random bytes", "error", err)
		return "", eris.Wrap(err, "generating random bytes")
	}

	token := base64.URLEncoding.EncodeToString(bytes)

	return token[:64], nil
}
