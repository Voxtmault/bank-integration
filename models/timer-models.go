package bank_integration_models

import (
	"database/sql"
	"time"
)

type TimerPayment struct {
	Id        int
	NumVA     string
	IdBank    int
	ExpiredAt time.Time
}

type TransactionWatcher struct {
	IDTransaction uint      // Identifier
	ExpireAt      time.Time // Time of expiration
	PaymentStatus chan bool // Controlls the watcher behavior
	MaxRetry      uint      // Maximum number of retries
	RetryCount    uint      // Current retry count
	Con           *sql.DB   // Database connection
}
