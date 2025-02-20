package bank_integration_utils

// Stored in redis as a hash set with the key being client-id and the value being the client-secret
var ClientCredentialsRedis = "client-credentials"
var AuthenticatedBankNameRedis = "authenticated-bank-name"
var VendorsLogoRedis = "vendors_logo"

// Reworked from the original code
var PartneredBanks = "partnered-banks"
var PartneredBanksCredentialsMapping = "partnered-banks-credentials-mapping"

// Format stored in redis is access-tokens:{token} as the key and the value is the client secret
// Structure is a regular key-value pair
var AccessTokenRedis = "access-tokens"

// Bank API Access Tokens
const (
	BCAAccessToken = "bca-access-token"
)

// RedisConsts
const (
	BCABalanceInquiry = "bca-balance-inquiry"
	BCABankStatement  = "bca-bank-statement"
)

var UniqueExternalIDRedis = "unique-external-id"

const (
	BankCodeBCA = "bca"
)

type VAPaymentStatus uint

const (
	VAStatusPending   VAPaymentStatus = 1
	VAStatusPaid      VAPaymentStatus = 2
	VAStatusExpired   VAPaymentStatus = 3
	VAStatusCancelled VAPaymentStatus = 4
)

type TransactionWatcherStatus uint

const (
	WatcherSuccess   TransactionWatcherStatus = 1
	WatcherFailed    TransactionWatcherStatus = 2
	WatcherCancelled TransactionWatcherStatus = 3
)

// Message Broker Constants
const (
	HelperUpdate = "helper-update"
)

// Bank Feature Constants
const (
	FeatureOAuth = iota + 1
	FeatureBillPresentment
	FeaturePaymentFlag
	FeaturePaymentStatus
	FeatureBalanceInquiry
	FeatureExternalAccountInquiry
	FeatureInternalAccountInquiry
	FeatureIntrabankTransfer
	FeatureInterbankTransfer
	FeatureBankStatement
	FeatureTransactionStatusInquiry
)

// Bank Integration Management Constants
var StartupHelper = []string{
	"bank_features", "bank_feature_types", "payment_methods",
}

// BCA Account Balance Infos Hint Text English
const (
	EN_BCABalanceTypeAmountHintText = "Net amount of the transaction"
	EN_BCABalanceTypeFloatHintText  = "Amount of deposit that is not effective yet (due to holiday, etc)"
	EN_BCABalanceTypeHoldHintText   = "Hold amount that cannot be used"
	EN_BCABalanceTypeAvailableHint  = "Account balance that can be used for financial transaction"
)

// BCA Account Balance Infos Hint Text Bahasa Indonesia
const (
	ID_BCABalanceTypeAmountHintText = "Jumlah bersih dari transaksi"
	ID_BCABalanceTypeFloatHintText  = "Jumlah deposit yang belum efektif (karena hari libur, dll)"
	ID_BCABalanceTypeHoldHintText   = "Jumlah yang tidak dapat digunakan"
	ID_BCABalanceTypeAvailableHint  = "Saldo akun yang dapat digunakan untuk transaksi keuangan"
)
