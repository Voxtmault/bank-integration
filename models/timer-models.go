package bank_integration_models

import (
	"time"

	biConst "github.com/voxtmault/bank-integration/utils"
)

type TimerPayment struct {
	Id        int
	NumVA     string
	IdBank    int
	ExpiredAt time.Time
}

type TransactionWatcher struct {
	IDTransaction uint                         // Identifier
	ExpireAt      time.Time                    // Time of expiration
	PaymentStatus chan biConst.VAPaymentStatus // Controlls the watcher behavior
	MaxRetry      uint                         // Maximum number of retries
	Attempts      uint                         // Current retry count
	Timer         *time.Timer                  // Timer for the watcher
}
