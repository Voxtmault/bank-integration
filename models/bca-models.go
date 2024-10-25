package bank_integration_models

type BCARequestHeader struct {
	Timestamp     string `validate:"required,timezone"`
	ContentType   string `validate:"required"`
	Signature     string `validate:"required"`
	ClientKey     string `validate:"required_without=Authorization"`
	Authorization string `validate:"required_without=ClientKey"`
	Origin        string `validate:"omitempty"`
	ExternalID    string `validate:"required_without=ClientKey,numeric,max=36"`
	ChannelID     string `validate:"required_with=PartnerID,max=5"`
	PartnerID     string `validate:"required_with=ChannelID,max=32"` // Company Code / ID
}

type GrantType struct {
	GrantType string `json:"grantType" validate:"required,eq=client_credentials"`
}

type SymmetricSignatureRequirement struct {
	HTTPMethod  string // GET, POST, PUT, PATCH, DELETE. Only accepts uppercase values
	AccessToken string // Access Token from BCA API
	Timestamp   string // Timestamp in RFC3339 format

	// Request body will be hashed with SHA-256 Algorithm. It will be converted into MinifyJSON (remove
	// whitespace except for the key or value json)
	//
	// If the request body is empty / nil then it will be set to empty string
	RequestBody []byte

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
	HTTPStatusCode  int    `json:"-"`
	ResponseCode    string `json:"responseCode"`    // BCA Unique Status Code
	ResponseMessage string `json:"responseMessage"` // BCA Message regarding the request
}

func (r BCAResponse) Data() (int, string, string) {
	return r.HTTPStatusCode, r.ResponseCode, r.ResponseMessage
}

type AccessTokenResponse struct {
	*BCAResponse
	AccessToken string `json:"accessToken,omitempty"`
	TokenType   string `json:"tokenType,omitempty"`
	ExpiresIn   string `json:"expiresIn,omitempty"`
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
type BCAVARequestPayload struct {
	PartnerServiceID string `json:"partnerServiceId" validate:"required,max=8"`   // Derived from X-PARTNER-ID
	CustomerNo       string `json:"customerNo" validate:"required,max=20"`        // Unique customer number
	VirtualAccountNo string `json:"virtualAccountNo" validate:"required,max=28"`  // Combined PartnerServiceID and CustomerNo
	InquiryRequestID string `json:"inquiryRequestId" validate:"required,max=128"` // Unique inquiry ID (generated by BCA)
}

type BillDescription struct {
	English   string `json:"english"`   // Deskripsi tagihan dalam bahasa Inggris
	Indonesia string `json:"indonesia"` // Deskripsi tagihan dalam bahasa Indonesia
}
type Amount struct {
	Value    string `json:"value" validate:"required"`    // Transaction amount
	Currency string `json:"currency" validate:"required"` // Currency code (e.g., IDR)
}

type VAResponsePayload struct {
	BCAResponse
	VirtualAccountData *VABCAResponseData `json:"virtualAccountData,omitempty"` // Virtual account data object
}

type VABCAResponseData struct {
	InquiryStatus         string                 `json:"inquiryStatus"`                          // Inquiry status
	InquiryReason         InquiryReason          `json:"inquiryReason"`                          // Reason for inquiry status
	PartnerServiceID      string                 `json:"partnerServiceId" validate:"required"`   // Derived from X-PARTNER-ID
	CustomerNo            string                 `json:"customerNo" validate:"required"`         // Customer number
	VirtualAccountNo      string                 `json:"virtualAccountNo" validate:"required"`   // Virtual account number
	VirtualAccountName    string                 `json:"virtualAccountName" validate:"required"` // Customer name
	VirtualAccountEmail   string                 `json:"virtualAccountEmail"`                    // Customer email (optional)
	VirtualAccountPhone   string                 `json:"virtualAccountPhone"`                    // Customer's phone number (optional)
	InquiryRequestID      string                 `json:"inquiryRequestId" validate:"required"`   // Inquiry request ID
	TotalAmount           Amount                 `json:"totalAmount"`                            // Total transaction amount
	SubCompany            string                 `json:"subCompany"`                             // Sub company code (optional)
	BillDetails           []BillInfo             `json:"billDetails"`                            // Bill details (optional for multi-settlement)
	FreeTexts             []FreeText             `json:"freeTexts"`                              // Optional free text displayed in channel
	FeeAmount             *Amount                `json:"feeAmount"`
	VirtualAccountTrxType string                 `json:"virtualAccountTrxType"` // Type of virtual account (optional)
	AdditionalInfo        map[string]interface{} `json:"additionalInfo"`        // Optional additional information
}

func (r VABCAResponseData) Default() *VABCAResponseData {
	data := &VABCAResponseData{
		InquiryReason:  InquiryReason{},
		BillDetails:    []BillInfo{},
		FreeTexts:      []FreeText{},
		TotalAmount:    Amount{},
		FeeAmount:      &Amount{},
		AdditionalInfo: map[string]interface{}{},
	}

	return data
}

type InquiryReason struct {
	English   string `json:"english"`   // Reason in English
	Indonesia string `json:"indonesia"` // Reason in Indonesian
}

type BillDetail struct {
	BillCode        string                 `json:"billCode"`        // Bill code
	BillNo          string                 `json:"billNo"`          // Bill number
	BillName        string                 `json:"billName"`        // Bill name
	BillShortName   string                 `json:"billShortName"`   // Short bill name
	BillDescription BillDescription        `json:"billDescription"` // Bill description
	BillSubCompany  string                 `json:"billSubCompany"`  // Bill sub company code (optional)
	BillAmount      Amount                 `json:"billAmount"`      // Amount for each bill
	BillStatus      string                 `json:"status"`
	Reason          Reason                 `json:"reason"`
	AdditionalInfo  map[string]interface{} `json:"additionalInfo"`
}
type BillInfo struct {
	BillCode        string                 `json:"billCode"`        // Kode tagihan, opsional
	BillNo          string                 `json:"billNo"`          // Nomor tagihan pelanggan, opsional
	BillName        string                 `json:"billName"`        // Nama tagihan, opsional
	BillShortName   string                 `json:"billShortName"`   // Nama singkat tagihan, opsional
	BillDescription BillDescription        `json:"billDescription"` // Deskripsi tagihan, bisa dalam dua bahasa
	BillSubCompany  string                 `json:"billSubCompany"`  // Kode sub-perusahaan, opsional
	BillAmount      Amount                 `json:"billAmount"`      // Jumlah tagihan dan mata uang
	BillAmountLabel string                 `json:"billAmountLabel"` // Label jumlah tagihan, opsional
	BillAmountValue string                 `json:"billAmountValue"` // Nilai yang ditampilkan untuk jumlah tagihan, opsional
	AdditionalInfo  map[string]interface{} `json:"additionalInfo"`  // Informasi tambahan, opsional
}

type FreeText struct {
	English   string `json:"english"`   // Free text in English
	Indonesia string `json:"indonesia"` // Free text in Indonesian
}

type BCAInquiryRequest struct {
	PartnerServiceID        string                 `json:"partnerServiceId" validate:"required"` // Partner ID (Company Code VA)
	CustomerNo              string                 `json:"customerNo" validate:"required"`       // Unique customer number
	VirtualAccountNo        string                 `json:"virtualAccountNo" validate:"required"` // Combination of partnerServiceId and customerNo
	VirtualAccountName      string                 `json:"virtualAccountName,omitempty"`         // Customer name (optional)
	VirtualAccountEmail     string                 `json:"virtualAccountEmail,omitempty"`        // Customer email (optional)
	VirtualAccountPhone     string                 `json:"virtualAccountPhone,omitempty"`        // Customer phone number (optional)
	TrxID                   string                 `json:"trxId,omitempty"`                      // Transaction ID, optional if the payment is not from Create VA Request
	PaymentRequestID        string                 `json:"paymentRequestId" validate:"required"` // Unique ID generated by BCA, must match inquiryRequestId
	ChannelCode             int                    `json:"channelCode,omitempty"`                // Channel code based on ISO 18245 (optional)
	HashedSourceAccountNo   string                 `json:"hashedSourceAccountNo,omitempty"`      // Source account number in hash (optional)
	SourceBankCode          string                 `json:"sourceBankCode,omitempty"`             // Source account bank code (optional)
	PaidAmount              Amount                 `json:"paidAmount" validate:"required"`       // Transaction amount (mandatory)
	CumulativePaymentAmount *Amount                `json:"cumulativePaymentAmount,omitempty"`    // Cumulative transaction amount (optional)
	PaidBills               string                 `json:"paidBills,omitempty"`                  // Flag of paid bills (optional)
	TotalAmount             Amount                 `json:"totalAmount" validate:"required"`      // Total transaction amount (mandatory)
	TrxDateTime             string                 `json:"trxDateTime,omitempty"`                // BCA system datetime with timezone in ISO-8601 format (optional)
	ReferenceNo             string                 `json:"referenceNo,omitempty"`                // Payment authorization code generated by BCA (optional)
	JournalNum              string                 `json:"journalNum,omitempty"`                 // Sequence journal number (optional)
	PaymentType             string                 `json:"paymentType,omitempty"`                // Type of payment (optional)
	FlagAdvise              string                 `json:"flagAdvise,omitempty"`                 // Retry flag status (optional, default 'N')
	SubCompany              string                 `json:"subCompany,omitempty"`                 // Sub company code (optional)
	BillDetails             []BillDetail           `json:"billDetails,omitempty"`                // Array of bill details (optional)
	FreeTexts               []FreeText             `json:"freeTexts,omitempty"`                  // Optional array of free text (optional)
	AdditionalInfo          map[string]interface{} `json:"additionalInfo,omitempty"`             // Additional information for custom use (optional)
}

type InquiryResponse struct {
	ResponseCode       string                 `json:"responseCode" validate:"required"`       // Response code from partner
	ResponseMessage    string                 `json:"responseMessage" validate:"required"`    // Response message from partner
	VirtualAccountData VirtualAccountData     `json:"virtualAccountData" validate:"required"` // Data related to virtual account
	AdditionalInfo     map[string]interface{} `json:"additionalInfo"`                         // Additional information (optional)
}

type VirtualAccountData struct {
	PaymentFlagReason   Reason       `json:"paymentFlagReason,omitempty"`            // Reason for payment status
	PartnerServiceID    string       `json:"partnerServiceId" validate:"required"`   // Partner ID
	CustomerNo          string       `json:"customerNo" validate:"required"`         // Customer number
	VirtualAccountNo    string       `json:"virtualAccountNo" validate:"required"`   // Virtual account number
	VirtualAccountName  string       `json:"virtualAccountName" validate:"required"` // Customer name
	VirtualAccountEmail string       `json:"virtualAccountEmail,omitempty"`          // Customer email (optional)
	VirtualAccountPhone string       `json:"virtualAccountPhone,omitempty"`          // Customer phone number (optional)
	TrxID               string       `json:"trxId,omitempty"`                        // Transaction ID
	PaymentRequestID    string       `json:"paymentRequestId" validate:"required"`   // Payment request ID
	PaidAmount          Amount       `json:"paidAmount" validate:"required"`         // Paid amount
	PaidBills           string       `json:"paidBills,omitempty"`                    // Flag of paid bills (optional)
	TotalAmount         Amount       `json:"totalAmount" validate:"required"`        // Total transaction amount
	TrxDateTime         string       `json:"trxDateTime,omitempty"`                  // Transaction datetime in ISO-8601 format
	ReferenceNo         string       `json:"referenceNo,omitempty"`                  // Payment reference number
	JournalNum          string       `json:"journalNum,omitempty"`                   // Journal number
	PaymentType         string       `json:"paymentType,omitempty"`                  // Type of payment
	FlagAdvise          string       `json:"flagAdvise,omitempty"`                   // Retry flag status
	PaymentFlagStatus   string       `json:"paymentFlagStatus,omitempty"`            // Status for payment flag
	BillDetails         []BillDetail `json:"billDetail"`                             // Array of bill details (optional)
	FreeTexts           []FreeText   `json:"freeTexts"`                              // Array of free texts (optional)
}

type Reason struct {
	English   string `json:"english"`   // Reason in English
	Indonesia string `json:"indonesia"` // Reason in Indonesian
}

type BCAInquiryVAResponse struct {
	BCAResponse
	VirtualAccountData *VirtualAccountDataInquiry `json:"virtualAccountData,omitempty"` // Data related to virtual account
	AdditionalInfo     map[string]interface{}     `json:"additionalInfo,omitempty"`     // Additional information (optional)
}

type VirtualAccountDataInquiry struct {
	PaymentFlagReason   Reason       `json:"paymentFlagReason"`                      // Reason for payment status
	PartnerServiceID    string       `json:"partnerServiceId" validate:"required"`   // Partner ID
	CustomerNo          string       `json:"customerNo" validate:"required"`         // Customer number
	VirtualAccountNo    string       `json:"virtualAccountNo" validate:"required"`   // Virtual account number
	VirtualAccountName  string       `json:"virtualAccountName" validate:"required"` // Customer name
	VirtualAccountEmail string       `json:"virtualAccountEmail"`                    // Customer email (optional)
	VirtualAccountPhone string       `json:"virtualAccountPhone"`                    // Customer phone number (optional)
	TrxID               string       `json:"trxId"`                                  // Transaction ID
	PaymentRequestID    string       `json:"paymentRequestId" validate:"required"`   // Payment request ID
	PaidAmount          Amount       `json:"paidAmount" validate:"required"`         // Paid amount
	PaidBills           string       `json:"paidBills"`                              // Flag of paid bills (optional)
	TotalAmount         Amount       `json:"totalAmount" validate:"required"`        // Total transaction amount
	TrxDateTime         string       `json:"trxDateTime"`                            // Transaction datetime in ISO-8601 format
	ReferenceNo         string       `json:"referenceNo"`                            // Payment reference number
	JournalNum          string       `json:"journalNum"`                             // Journal number
	PaymentType         string       `json:"paymentType"`                            // Type of payment
	FlagAdvise          string       `json:"flagAdvise"`                             // Retry flag status
	PaymentFlagStatus   string       `json:"paymentFlagStatus"`                      // Status for payment flag
	BillDetails         []BillDetail `json:"billDetails"`                            // Array of bill details (optional)
	FreeTexts           []FreeText   `json:"freeTexts"`                              // Array of free texts (optional)
}

func (r VirtualAccountDataInquiry) Default() *VirtualAccountDataInquiry {
	data := &VirtualAccountDataInquiry{
		PaymentFlagReason: Reason{},
		BillDetails:       []BillDetail{},
		FreeTexts:         []FreeText{},
		PaidAmount:        Amount{},
		TotalAmount:       Amount{},
	}

	return data
}

type AdditionalInfo struct {
	Label Reason `json:"label"` // Label for additional information
	Value Reason `json:"value"` // Value for additional information
}

type CreateVAReq struct {
	IdUser           int    `json:"id_user"`
	NamaUser         string `json:"nama_user"`
	IdJenisPembelian int    `json:"id_jenis_pembelian"`
	IdJenisUser      int    `json:"id_jenis_user"`
	JumlahPembayaran int    `json:"jumlah_pembayaran"`
	// TotalAmount      Amount `json:"total_amaount"`
}
