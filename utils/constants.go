package bank_integration_utils

// Stored in redis as a hash set with the key being client-id and the value being the client-secret
var ClientCredentialsRedis = "client-credentials"

// Format stored in redis is access-tokens:{token} as the key and the value is the client secret
// Structure is a regular key-value pair
var AccessTokenRedis = "access-tokens"

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
