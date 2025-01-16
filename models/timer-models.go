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
	IDTransaction   uint                         // Identifier
	IDBank          uint                         // Bank who owns the transaction
	BankName        string                       // Bank name
	Location        *time.Location               // Timezone
	ExpireAt        time.Time                    // Time of expiration
	PaymentStatus   chan biConst.VAPaymentStatus // Controlls the watcher behavior
	ExternalChannel chan uint                    // Inject other channel, probably from importer
	MaxRetry        uint                         // Maximum number of retries
	Attempts        uint                         // Current retry count
	Timer           *time.Timer                  // Timer for the watcher
}

type TransactionWatcherPublic struct {
	IDTransaction uint   `json:"id_transaction"`
	IDBank        uint   `json:"id_bank"`
	BankName      string `json:"bank_name"`
	Attempts      uint   `json:"attempts"`
	MaxRetry      uint   `json:"max_retry"`
	RemainingTime string `json:"remaining_time"`
}

func (w *TransactionWatcher) ToPublic() *TransactionWatcherPublic {
	data := TransactionWatcherPublic{}

	data.IDTransaction = w.IDTransaction
	data.IDBank = w.IDBank
	data.BankName = w.BankName
	data.Attempts = w.Attempts
	data.MaxRetry = w.MaxRetry
	data.RemainingTime = time.Until(w.ExpireAt.Local()).String()

	return &data
}
