package watcher

import (
	"database/sql"
	"log/slog"

	biModel "github.com/voxtmault/bank-integration/models"
	biConst "github.com/voxtmault/bank-integration/utils"
)

type TransactionWatcher struct {
	WatchedList map[uint]*biModel.TransactionWatcher
}

func NewTransactionWatcher() *TransactionWatcher {
	return &TransactionWatcher{
		WatchedList: make(map[uint]*biModel.TransactionWatcher),
	}
}

func (s *TransactionWatcher) AddWatcher(watcher *biModel.TransactionWatcher) {
	s.WatchedList[watcher.IDTransaction] = watcher
}

func (s *TransactionWatcher) RemoveWatcher(id uint) {
	delete(s.WatchedList, id)
}

func (s *TransactionWatcher) GetWatcher(id uint) *biModel.TransactionWatcher {
	return s.WatchedList[id]
}

func (s *TransactionWatcher) GetWatchers() map[uint]*biModel.TransactionWatcher {
	return s.WatchedList
}

// expireFunc is only called when the ticker / timer of the watcher has expired, updating the said transaction status from
// waiting to expired, before updating watcher will also check if the transaction has been completed or not, if it is
// then do nothing.
func (s *TransactionWatcher) expireFunc(obj *biModel.TransactionWatcher) error {

	// Check if the transaction has been completed
	var transactionStatus uint
	statement := `
	SELECT id_va_status
	FROM va_request
	WHERE id = ?
	`
	if err := obj.Con.QueryRow(statement, obj.IDTransaction).Scan(&transactionStatus); err != nil {
		if err == sql.ErrNoRows {
			slog.Info("transaction not found, killing watcher")
			return nil
		} else {
			return err
		}
	}

	if transactionStatus == uint(biConst.VAStatusPaid) {
		slog.Info("transaction already paid, killing watcher")
		return nil
	}

	// Transaction is still on waiting, update the status to expired
	statement = `
	UPDATE va_reqest SET id_va_status = ?
	WHERE id = ?
	`
	tx, err := obj.Con.Begin()
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
