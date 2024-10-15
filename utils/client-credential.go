package utils

import (
	"github.com/google/uuid"
)

// This file is mainly used to generate client id and client secret.
//
// Client in this context are banks that needs to interact with our API;
// A few things they will do are:
//
// 1. Get access token (preferable through OAuth2.0)
//
// 2. Bill Presentment, i have no idea what this do
//
// 3. Payment Flag, probably for banks to update transaction statuses on our end

type ClientCredential struct {
}

// GenerateClientCredential returns a client id and client secret. The generated client credentials
// are simple UUIDv4
func (s *ClientCredential) GenerateClientCredential() (string, string) {
	return uuid.New().String(), uuid.New().String()
}
