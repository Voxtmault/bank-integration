package timerexpired

import (
	"sync"
	"time"

	"github.com/rotisserie/eris"
	biModel "github.com/voxtmault/bank-integration/models"
	biStorage "github.com/voxtmault/bank-integration/storage"
)

var transactions = make(map[int]*biModel.TimerPayment)
var mutex = sync.Mutex{}

func expireTransaction(id int) error {
	mutex.Lock()
	defer mutex.Unlock()
	con := biStorage.GetDBConnection()
	if txn, exists := transactions[id]; exists {
		query := `UPDATE va_request SET id_va_status = 3 WHERE id = ?;`
		_, err := con.Exec(query, txn.Id)
		if err != nil {
			return eris.Wrap(err, "Update Expired Status")
		}
		delete(transactions, id)
	}
	return nil
}

func SetTimer(transaction biModel.TimerPayment) {
	// Menambahkan transaksi ke peta transaksi
	mutex.Lock()
	transactions[transaction.Id] = &transaction
	mutex.Unlock()

	// Durasi kadaluarsa transaksi
	duration := time.Until(transaction.ExpiredAt)
	timer := time.NewTimer(duration) // Timer berdasarkan waktu kadaluarsa

	// Goroutine untuk menangani timer
	go func() {
		<-timer.C
		err := expireTransaction(transaction.Id)
		if err != nil {
			for err == nil {
				err = expireTransaction(transaction.Id)
			}
		}
	}()
	// fmt.Printf("Transaction %d will expire in %v.\n", transaction.ID, duration)
}
