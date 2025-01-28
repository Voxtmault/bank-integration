package watcher

import (
	"database/sql"
	"log/slog"
	"sync"
	"time"

	biConfig "github.com/voxtmault/bank-integration/config"
	biModel "github.com/voxtmault/bank-integration/models"
	biStorage "github.com/voxtmault/bank-integration/storage"
	biConst "github.com/voxtmault/bank-integration/utils"
)

type TransactionWatcher struct {
	con         *sql.DB
	WatchedList map[uint]*biModel.TransactionWatcher
	sync.RWMutex
}

func NewTransactionWatcher() *TransactionWatcher {
	return &TransactionWatcher{
		con:         biStorage.GetDBConnection(),
		WatchedList: make(map[uint]*biModel.TransactionWatcher),
	}
}

func NewWatcher() *biModel.TransactionWatcher {
	return &biModel.TransactionWatcher{
		MaxRetry:      biConfig.GetConfig().TransactionWatcherConfig.MaxRetry,
		ExpireAt:      time.Now().Add(biConfig.GetConfig().TransactionWatcherConfig.DefaultExpireTime),
		PaymentStatus: make(chan biConst.VAPaymentStatus),
	}
}

func (s *TransactionWatcher) AddWatcher(watcher *biModel.TransactionWatcher) {
	s.Lock()

	// Check for existing transaction ID
	if _, ok := s.WatchedList[watcher.IDTransaction]; ok {
		// Transaction already exists, skipping
		slog.Warn("skipping transaction since it already exists in watcher list", "transaction id", watcher.IDTransaction)
		return
	}
	s.WatchedList[watcher.IDTransaction] = watcher
	s.Unlock()

	// Calculate the time remaining
	remainingTimer := time.Until(watcher.ExpireAt)
	slog.Debug(remainingTimer.String())
	if remainingTimer < 0 {
		slog.Warn("remaining time is less than 0, setting to 10 seconds into the future", "remaining time", remainingTimer.String())
		remainingTimer = time.Until(time.Now().Add(time.Second * 10))
	}

	watcher.Timer = time.NewTimer(remainingTimer)

	go func(w *biModel.TransactionWatcher) {
		select {
		// Wait till the timer expires
		case <-w.Timer.C:
			if err := s.expireFunc(w); err != nil {
				slog.Error("error while expiring transaction", "error", err)
				watcher.ExpireAt = watcher.ExpireAt.Add(biConfig.GetConfig().TransactionWatcherConfig.DefaultRetryInterval)

				// Log error
				s.logWatcher(w, biConst.WatcherFailed, err.Error())

				// Remove the current watcher
				s.RemoveWatcher(w.IDTransaction)

				// Add another attempt
				s.AddWatcher(w)
				return
			} else {
				slog.Debug("successfully expired transaction")
				// No errors, add log to watcher table and remove the watcher
				s.logWatcher(w, biConst.WatcherSuccess, "watcher successfully run")

				// Make sure this is thread safe
				s.RemoveWatcher(w.IDTransaction)
				return
			}
		case paymentStatus := <-w.PaymentStatus:
			slog.Info("payment status received", "status", paymentStatus)
			// Stop the timer
			if !w.Timer.Stop() {
				<-w.Timer.C // If an error occurs, drain the channel
			}

			// Notify external channel
			w.ExternalChannel <- uint(paymentStatus)

			if paymentStatus == biConst.VAStatusPaid {
				s.logWatcher(w, biConst.WatcherCancelled, "transaction has been paid")

			} else {
				// This else is for when the transaction is cancelled
				s.logWatcher(w, biConst.WatcherCancelled, "transaction has been cancelled")
			}

			s.RemoveWatcher(w.IDTransaction)
			return
		}
	}(watcher)

}

func (s *TransactionWatcher) RemoveWatcher(id uint) {
	s.Lock()
	defer s.Unlock()
	delete(s.WatchedList, id)
}

func (s *TransactionWatcher) GetWatcher(id uint) *biModel.TransactionWatcherPublic {
	return s.WatchedList[id].ToPublic()
}

func (s *TransactionWatcher) GetWatchers() []*biModel.TransactionWatcherPublic {
	s.RLock()
	defer s.RUnlock()
	var watchers []*biModel.TransactionWatcherPublic
	for _, watcher := range s.WatchedList {
		watchers = append(watchers, watcher.ToPublic())
	}
	return watchers
}

func (s *TransactionWatcher) TransactionPaid(idTransaction uint) {
	s.Lock()
	defer s.Unlock()
	if watcher, exists := s.WatchedList[idTransaction]; exists {
		slog.Info("transaction has been paid", "transaction id", idTransaction)
		watcher.PaymentStatus <- biConst.VAStatusPaid
	}
}

// expireFunc is only called when the ticker / timer of the watcher has expired, updating the said transaction status from
// waiting to expired, before updating watcher will also check if the transaction has been completed or not, if it is
// then do nothing.
func (s *TransactionWatcher) expireFunc(obj *biModel.TransactionWatcher) error {
	// Increment the attempt
	obj.Attempts++

	// Remove from the watcher list
	s.RemoveWatcher(obj.IDTransaction)

	// Check if the transaction has been completed
	var transactionStatus uint
	statement := `
	SELECT id_va_status
	FROM va_request
	WHERE id = ?
	`
	if err := s.con.QueryRow(statement, obj.IDTransaction).Scan(&transactionStatus); err != nil {
		if err == sql.ErrNoRows {
			slog.Info("transaction not found, killing watcher")
			return nil
		} else {
			return err
		}
	}

	if transactionStatus == uint(biConst.VAStatusPaid) || transactionStatus == uint(biConst.VAStatusCancelled) {
		slog.Info("current transaction is either already paid or cancelled, killing watcher", "current status", transactionStatus)
		return nil
	}

	// Transaction is still on waiting, update the status to expired
	statement = `
	UPDATE va_request SET id_va_status = ?
	WHERE id = ?
	`
	tx, err := s.con.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	if _, err = tx.Exec(statement, biConst.VAStatusExpired, obj.IDTransaction); err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *TransactionWatcher) logWatcher(watcher *biModel.TransactionWatcher, status biConst.TransactionWatcherStatus, message string) error {

	tx, err := s.con.Begin()
	if err != nil {
		tx.Rollback()
		slog.Error("error while logging watcher", "starting transaction", err)
		return err
	}

	statement := `
	INSERT INTO transaction_watcher_log (id_transaction, id_watcher_status, message, attempts, max_attempts)
	VALUES (?, ?, ?, ?, ?)
	`
	if _, err := tx.Exec(statement, watcher.IDTransaction, status, message, watcher.Attempts, watcher.MaxRetry); err != nil {
		tx.Rollback()
		slog.Error("error while logging watcher", "exec statement", err)
		return err
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		slog.Error("error while logging watcher", "commit transaction", err)
		return err
	}

	return nil
}
