package bca_security

import (
	"bytes"
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/rotisserie/eris"
	"github.com/tdewolff/minify/v2"
	mjson "github.com/tdewolff/minify/v2/json"
	"github.com/voxtmault/bank-integration/config"
	"github.com/voxtmault/bank-integration/interfaces"
	"github.com/voxtmault/bank-integration/models"
)

type BCASecurity struct {
	PrivateKeyPath string
	PublicKeyPath  string
	ClientID       string // Given by BCA
	ClientSecret   string // Given by BCA
	privateKey     *rsa.PrivateKey
	publicKey      *rsa.PublicKey
}

var _ interfaces.Security = &BCASecurity{}

func NewBCASecurity(bcaConfig *config.BCAConfig, keys *config.Keys) *BCASecurity {
	return &BCASecurity{
		PrivateKeyPath: keys.PrivateKeyPath,
		PublicKeyPath:  keys.PublicKeyPath,
		ClientID:       bcaConfig.ClientID,
		ClientSecret:   bcaConfig.ClientSecret,
	}
}

func (s *BCASecurity) CreateAsymmetricSignature(ctx context.Context, timeStamp string) (string, error) {
	var err error

	// Checks if the private key is already loaded
	if s.privateKey == nil {
		// Load the Private Key from file
		s.privateKey, err = loadPrivateKey(s.PrivateKeyPath)
		if err != nil {
			return "", eris.Wrap(err, "loading private key")
		}
	}

	// Hash the String To Sign
	hashed := sha256.Sum256([]byte(fmt.Sprintf("%s|%s", s.ClientID, timeStamp)))

	// Sign the hashed string
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", eris.Wrap(err, "signing string")
	}

	// Return the signature as a base64 encoded string
	return base64.StdEncoding.EncodeToString(signature), nil
}

func (s *BCASecurity) VerifyAsymmetricSignature(ctx context.Context, timeStamp, clientKey, signature string) (bool, error) {

	if s.publicKey == nil {
		var err error
		s.publicKey, err = loadPublicKey(s.PublicKeyPath)
		if err != nil {
			return false, eris.Wrap(err, "loading public key")
		}
	}

	// Decode the received signature
	decodedSignature, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, eris.Wrap(err, "decoding signature")
	}

	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%s|%s", clientKey, timeStamp)))
	hashed := hash.Sum(nil)

	// Verify the signature
	if err = rsa.VerifyPKCS1v15(s.publicKey, crypto.SHA256, hashed, decodedSignature); err != nil {
		return false, eris.Wrap(err, "verifying signature")
	} else {
		return true, nil
	}
}

func (s *BCASecurity) CreateSymmetricSignature(ctx context.Context, obj *models.SymetricSignatureRequirement) (string, error) {

	// Encode the Relative URL
	relativeURL, err := processRelativeURL(obj.RelativeURL)
	if err != nil {
		return "", eris.Wrap(err, "processing relative url")
	}

	// Generate the hash value of Request Body
	requestBody, err := processRequestBody(obj.RequestBody)
	if err != nil {
		return "", eris.Wrap(err, "processing request body")
	}
	stringToSign := obj.HTTPMethod + ":" + relativeURL + ":" + obj.AccessToken + ":" + requestBody + ":" + obj.Timestamp

	// Generate Signature using SHA512-HMAC Algorithm
	h := hmac.New(sha512.New, []byte(s.ClientSecret))
	h.Write([]byte(stringToSign))
	signature := h.Sum(nil)

	return base64.StdEncoding.EncodeToString(signature), nil
}

func (s *BCASecurity) VerifySymmetricSignature(ctx context.Context, obj *models.SymetricSignatureRequirement, clientSecret, signature string) (bool, error) {
	// Encode the Relative URL
	relativeURL, err := processRelativeURL(obj.RelativeURL)
	if err != nil {
		return false, eris.Wrap(err, "processing relative url")
	}

	// Generate the hash value of Request Body
	requestBody, err := processRequestBody(obj.RequestBody)
	if err != nil {
		return false, eris.Wrap(err, "processing request body")
	}
	stringToSign := obj.HTTPMethod + ":" + relativeURL + ":" + obj.AccessToken + ":" + requestBody + ":" + obj.Timestamp

	h := hmac.New(sha512.New, []byte(clientSecret))
	h.Write([]byte(stringToSign))

	calculatedSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(calculatedSignature)), nil
}

// Helper Functions

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	// Read the private key file

	slog.Info("Loading Private Key", "Path: ", path)

	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, eris.Wrap(err, "reading file")
	}

	// Decode the PEM Block
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, eris.New("failed to decode PEM block containing private key")
	}

	// Check the type of the PEM block
	var privateKey *rsa.PrivateKey
	if block.Type == "RSA PRIVATE KEY" {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, eris.Wrap(err, "failed to parse PKCS1 private key")
		}
	} else if block.Type == "PRIVATE KEY" {
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, eris.Wrap(err, "failed to parse PKCS8 private key")
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, eris.New("not an RSA private key")
		}
	} else {
		return nil, eris.New("unsupported key type " + block.Type)
	}

	return privateKey, nil
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, fmt.Errorf("unexpected type of public key")
	}
}

// processRequestBody is a helper function that returns a lowercase hex encoded SHA256 hash of the minified request body
func processRequestBody(obj any) (string, error) {
	// MinifyJSON the Request Body
	minifiedBody, err := minifyJSON(obj)
	if err != nil {
		return "", eris.Wrap(err, "minifying json")
	}

	// Hash the minifiedBody with SHA256
	hashed := sha256.Sum256([]byte(minifiedBody))

	// Hex Encode the hash value since it is returning a binary stream
	hexHashed := hex.EncodeToString(hashed[:])

	// Convert the hex value into lower case
	lowerHexHashed := strings.ToLower(hexHashed)

	return lowerHexHashed, nil
}

func minifyJSON(obj any) (string, error) {
	// Logic
	// 1. Marshall the obj
	// 2. Minify the marshalled obj result

	jsonData, err := json.Marshal(obj)
	if err != nil {
		return "", eris.Wrap(err, "marshalling object")
	}

	m := minify.New()
	m.AddFunc("application/json", mjson.Minify)
	var buf bytes.Buffer
	if err := m.Minify("application/json", &buf, bytes.NewReader(jsonData)); err != nil {
		return "", eris.Wrap(err, "minifying json")
	}

	return buf.String(), nil
}

func processRelativeURL(rawUrl string) (string, error) {

	// Parse the URL
	parsedURL, err := url.Parse(rawUrl)
	if err != nil {
		return "", eris.Wrap(err, "parsing url")
	}

	// Encode the path
	encodedPath := encodePath(parsedURL.Path)

	// Encode the query
	encodedQuery := encodeQuery(parsedURL.RawQuery)

	encodedURL := encodedPath
	if encodedQuery != "" {
		encodedURL += "?" + encodedQuery
	}

	return encodedURL, nil
}

func encodePath(path string) string {
	var encodedPath strings.Builder
	for _, c := range path {
		if isUnreserved(c) || c == '/' {
			encodedPath.WriteRune(c)
		} else {
			encodedPath.WriteString(percentEncode(c))
		}
	}

	return encodedPath.String()
}

func isUnreserved(c rune) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' || c == '~'
}

func percentEncode(c rune) string {
	return "%" + strings.ToUpper(hex.EncodeToString([]byte(string(c))))
}

// encodeQuery encodes the query component of the URL
func encodeQuery(rawQuery string) string {
	// Parse the query parameters
	queryParams, err := url.ParseQuery(rawQuery)
	if err != nil {
		return ""
	}

	// Sort the query parameters by name and value
	var sortedParams []string
	for name, values := range queryParams {
		sort.Strings(values)
		for _, value := range values {
			sortedParams = append(sortedParams, name+"="+value)
		}
	}
	sort.Strings(sortedParams)

	// Encode the query parameters
	var encodedQuery strings.Builder
	for i, param := range sortedParams {
		if i > 0 {
			encodedQuery.WriteString("&")
		}
		for _, c := range param {
			if isUnreserved(c) || c == '?' || c == '=' || c == '&' {
				encodedQuery.WriteRune(c)
			} else {
				encodedQuery.WriteString(percentEncode(c))
			}
		}
	}
	return encodedQuery.String()
}