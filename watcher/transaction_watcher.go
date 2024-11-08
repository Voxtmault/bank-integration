package watcher

import (
	biModel "github.com/voxtmault/bank-integration/models"
)

type TransactionWatcher struct {
	WatchedList map[uint]*biModel.TransactionWatcher
}

func NewTransactionWatcher() *TransactionWatcher {
	return &TransactionWatcher{
		WatchedList: make(map[uint]*biModel.TransactionWatcher),
	}
}
