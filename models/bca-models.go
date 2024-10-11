package models

type GrantType struct {
	GrantType string `json:"grantType"`
}

type SymetricSignatureRequirement struct {
	HTTPMethod  string // GET, POST, PUT, PATCH, DELETE. Only accepts uppercase values
	AccessToken string // Access Token from BCA API
	Timestamp   string // Timestamp in RFC3339 format

	// Request body will be hashed with SHA-256 Algorithm. It will be converted into MinifyJSON (remove
	// whitespace except for the key or value json)
	//
	// If the request body is empty / nil then it will be set to empty string
	RequestBody any

	// The Relative URL will be URI-encoded according to the following rules:
	//
	// 1. Do not URI-encode forward slash ( / ) if it was used as path component.
	//
	// 2. Do not URI-encode question mark ( ? ), equals sign ( = ), and ampersand ( & ) if they were
	// used as query string component: as separator between the path and query string,
	// between query parameter and its value, and between each query parameter and value
	// pairs.
	//
	// 3. Do not URI-encode these characters: A-Z, a-z, 0-9, hyphen ( - ), underscore ( _ ), period ( .
	// ), and tilde ( ~ ) which are defined as unreserved characters in RFC 3986.
	// As for RFC 3986, means that comma that appear in the value of query params or path
	// params should be encoded too when generating Signature.
	//
	// 4. Percent-encode all other characters not meeting the above conditions using the format:
	// %XY, where X and Y are hexadecimal characters (0-9 and uppercase A-F).
	// For example, the space character must be encoded as %20 (not using '+', as some
	// encoding schemes do) and extended UTF-8 characters must be in the form %XY%ZA%BC.
	//
	// The query string parameters must be re-ordered according to the following rules:
	//
	// 1. Sorted by parameter name lexicographically
	//
	// 2. If there are two or more parameters with the same name, sort them by parameter values.
	RelativeURL string
}

type BCAResponse struct {
	HTTPStatusCode  int    `json:"httpStatusCode"`  // HTTP Status Code, Custom for Shifter Wallet
	ResponseCode    string `json:"responseCode"`    // BCA Unique Status Code
	ResponseMessage string `json:"responseMessage"` // BCA Message regarding the request
}

type AccessTokenResponse struct {
	BCAResponse
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   string `json:"expiresIn"`
}

type BCABalance struct {
	// If it's IDR then value includes 2 decimal digits. BCA will send response amount with format value
	// Numeric (13.2)
	Value string `json:"value"`
	// Currency of the account / balance. Defined in ISO4217
	Currency string `json:"currency"`
}
type BCAAccountInfo struct {
	// Account type name
	BalanceType string `json:"balanceType"`

	// Net ammount of the transaction
	Amount BCABalance `json:"amount"`

	// Amount of deposit that is not yet effective yet (due to holiday, etc...)
	FloatAmount BCABalance `json:"floatAmount"`

	// Hold amount that cannot be used
	HoldAmount BCABalance `json:"holdAmount"`

	// Account balance that can be used for financial transactions
	AvailableBalance BCABalance `json:"availableBalance"`

	// Account balance at the beginning of each day
	LedgerBalance BCABalance `json:"ledgerBalance"`

	// Credit limit of the account / plafon
	CurrentMultilateralLimit BCABalance `json:"currentMultilateralLimit"`

	// Customer registration status
	RegistrationStatusCode string `json:"registrationStatusCode"`

	// Account Status;
	//
	// 1. Active Account
	//
	// 2. Closed Account
	//
	// 4. New Account
	//
	// 6. Restricted Account
	//
	// 7. Frozen Account
	//
	// 9. Dormant Account
	Status string `json:"status"`
}
type BCAAccountBalance struct {
	BCAResponse
	// Transaction identifier on service provider system. Unique each day from BCA
	//
	// Must be filled upon successful transaction
	ReferenceNumber string `json:"referenceNo"`

	// Transaction identifier on service customer system
	PartnerReferenceNumber string `json:"partnerReferenceNo"`

	// Registered account number in KlikBCA Bisnis. For BCA, length account number is 10 digit
	AccountNumber string `json:"accountNo"`

	// Customer account name
	AccountName string `json:"name"`

	// Information regarding the account
	AccountInfos BCAAccountInfo `json:"accountInfos"`
}

// BCABalanceInquiry is a struct that is used to query a BCA Account Balance
type BCABalanceInquiry struct {

	// Transaction identifier on service provider system.
	PartnerReferenceNumber string `json:"partnerReferenceNo" validate:"required,max=64"`

	//Bank account number using registered number in KlikBCA Bisnis. For BCA, length account number is 10 digit
	AccountNumber string `json:"accountNo" validate:"required,max=16"`

	// Only filled if Account Number is null and Authorization-Customer is null
	BankCardToken string `json:"bankCardToken" validate:"required_without=AccountNumber,max=128"`

	// Balance Types of this parameter does not exists., its mean to inquiry all Balance Type on the account
	BalanceTypes []string `json:"balanceTypes"`
}
