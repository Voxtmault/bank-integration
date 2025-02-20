package bank_integration_test

import (
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	biModels "github.com/voxtmault/bank-integration/models"
)

func TestInteralBankBalanceFromBCA(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	obj := biModels.BankAccountBalance{}.Default()
	var bcaObj biModels.BCAAccountBalance
	bcaResponseStr := `{"responseCode":"2001100","responseMessage":"Successful","referenceNo":"20250220000000000263115","partnerReferenceNo":"a7591248-1584-428e-b37a-4b61aeb281c6","accountNo":"0611102380","name":"Tahapan 0611102380","accountInfos":[{"amount":{"value":"    944866781490.96","currency":"IDR"},"floatAmount":{"value":"               0.00","currency":"IDR"},"holdAmount":{"value":"          100000.00","currency":"IDR"},"availableBalance":{"value":"    944866681490.96","currency":"IDR"}}]}`

	if err := json.Unmarshal([]byte(bcaResponseStr), &bcaObj); err != nil {
		t.Error("failed to unmarshal bca response", err)
	}

	obj.FromBCAResponse(&bcaObj)

	slog.Debug("bank account balance", "obj", obj)
	time.Sleep(100 * time.Millisecond)
}

func TestInteralBankStatementFromBCA(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	obj := biModels.BankStatement{}.Default()
	var bcaObj biModels.BCABankStatementResponse
	bcaResponseStr := `{"responseCode":"2001400","responseMessage":"Successful","referenceNo":"20250220000000000263127","partnerReferenceNo":"1b6b1538-129d-49c7-b99a-284af8a4d87a","balance":[{"startingBalance":{"value":"944866826990.96","currency":"IDR","dateTime":"2025-02-19T11:57:29"},"endingBalance":{"value":"944866781490.96","currency":"IDR","dateTime":"2025-02-19T17:41:59+07:00"},"amount":{"value":"944866781490.96","currency":"IDR","dateTime":"2025-02-19T17:41:59+07:00"}}],"totalCreditEntries":{"numberOfEntries":"0","amount":{"value":"0.00","currency":"IDR"}},"totalDebitEntries":{"numberOfEntries":"6","amount":{"value":"45500.00","currency":"IDR"}},"detailData":[{"amount":{"value":"10000.00","currency":"IDR"},"transactionDate":"2025-02-19T11:57:29+07:00","remark":"SWITCHING TRSF DB 10000.00 Yories Yolanda","type":"DEBIT"},{"amount":{"value":"6500.00","currency":"IDR"},"transactionDate":"2025-02-19T11:57:29+07:00","remark":"SWITCHING BIAYA DB 6500.00 Yories Yolanda","type":"DEBIT"},{"amount":{"value":"10000.00","currency":"IDR"},"transactionDate":"2025-02-19T17:33:03+07:00","remark":"SWITCHING TRSF DB 10000.00 Yories Yolanda","type":"DEBIT"},{"amount":{"value":"6500.00","currency":"IDR"},"transactionDate":"2025-02-19T17:33:03+07:00","remark":"SWITCHING BIAYA DB 6500.00 Yories Yolanda","type":"DEBIT"},{"amount":{"value":"10000.00","currency":"IDR"},"transactionDate":"2025-02-19T17:41:59+07:00","remark":"BI-FAST TRSF DB 10000.00 Yories Yolanda","type":"DEBIT"},{"amount":{"value":"2500.00","currency":"IDR"},"transactionDate":"2025-02-19T17:41:59+07:00","remark":"BI-FAST BIAYA DB 2500.00 Yories Yolanda","type":"DEBIT"}]}`

	if err := json.Unmarshal([]byte(bcaResponseStr), &bcaObj); err != nil {
		t.Error("failed to unmarshal bca response", err)
	}

	obj.FromBCAResponse(&bcaObj)

	slog.Debug("bank statement", "obj", obj)
	time.Sleep(100 * time.Millisecond)
}
