package bank_integration_models

import (
	"log/slog"
	"strings"
	"time"

	biUtil "github.com/voxtmault/bank-integration/utils"
)

type InternalVAInformation struct {
	TrxID         string `json:"trx_id"`
	IDBank        uint   `json:"id_bank"`
	BankName      string `json:"bank_name"`
	BankIconLink  string `json:"bank_icon_link"`
	VANumber      string `json:"va_number"`
	VAAccountName string `json:"va_account_name"`
	TotalAmount   string `json:"total_amount"`
	ExpiredAt     string `json:"expired_at"`
}

type HintText struct {
	English    string `json:"en"`
	Indonesian string `json:"id"`
}

type BalanceAmount struct {
	Value    string    `json:"value"`
	Currency string    `json:"currency"`
	HintText *HintText `json:"hint_text,omitempty"`
}

func (b BalanceAmount) Default() *BalanceAmount {
	return &BalanceAmount{
		Value:    "",
		Currency: "",
	}
}

type BalanceDetails struct {
	AvailableBalance *BalanceAmount `json:"available_balance"`
	FloatBalance     *BalanceAmount `json:"float_balance"`
	HoldBalance      *BalanceAmount `json:"hold_balance"`
}

func (b BalanceDetails) Default() BalanceDetails {
	return BalanceDetails{
		AvailableBalance: &BalanceAmount{},
		FloatBalance:     &BalanceAmount{},
		HoldBalance:      &BalanceAmount{},
	}
}

type BankAccountBalance struct {
	AccountNumber  string          `json:"account_number"`
	AccountName    string          `json:"account_name"`
	Balance        *BalanceAmount  `json:"balance"`
	BalanceDetails *BalanceDetails `json:"balance_details"`
}

func (b BankAccountBalance) Default() BankAccountBalance {
	obj := BankAccountBalance{
		AccountNumber: "",
		AccountName:   "",
		Balance:       &BalanceAmount{},
	}
	balanceDetails := BalanceDetails{}.Default()
	obj.BalanceDetails = &balanceDetails

	return obj
}

func (b *BankAccountBalance) FromBCAResponse(bca *BCAAccountBalance) {
	b.AccountName = bca.AccountName
	b.AccountNumber = bca.AccountNumber

	if len(bca.AccountInfos) == 0 {
		slog.Warn("no account info found")
		return
	}

	info := bca.AccountInfos[0]

	b.Balance.Value = strings.TrimSpace(info.Amount.Value)
	b.Balance.Currency = info.Amount.Currency
	b.Balance.HintText = &HintText{
		English:    biUtil.EN_BCABalanceTypeAmountHintText,
		Indonesian: biUtil.ID_BCABalanceTypeAmountHintText,
	}

	b.BalanceDetails.AvailableBalance.Value = strings.TrimSpace(info.AvailableBalance.Value)
	b.BalanceDetails.AvailableBalance.Currency = info.AvailableBalance.Currency
	b.BalanceDetails.AvailableBalance.HintText = &HintText{
		English:    biUtil.EN_BCABalanceTypeAvailableHint,
		Indonesian: biUtil.ID_BCABalanceTypeAvailableHint,
	}

	b.BalanceDetails.FloatBalance.Value = strings.TrimSpace(info.FloatAmount.Value)
	b.BalanceDetails.FloatBalance.Currency = info.FloatAmount.Currency
	b.BalanceDetails.FloatBalance.HintText = &HintText{
		English:    biUtil.EN_BCABalanceTypeFloatHintText,
		Indonesian: biUtil.ID_BCABalanceTypeFloatHintText,
	}

	b.BalanceDetails.HoldBalance.Value = strings.TrimSpace(info.HoldAmount.Value)
	b.BalanceDetails.HoldBalance.Currency = info.HoldAmount.Currency
	b.BalanceDetails.HoldBalance.HintText = &HintText{
		English:    biUtil.EN_BCABalanceTypeHoldHintText,
		Indonesian: biUtil.ID_BCABalanceTypeHoldHintText,
	}
}

type BankStatementBalanceDetail struct {
	BalanceDetails *BalanceAmount `json:"balance_details"`
	DateTime       string         `json:"date_time"`
}

type BankStatementBalance struct {
	Amount          *BankStatementBalanceDetail `json:"amount"`
	StartingBalance *BankStatementBalanceDetail `json:"starting_balance"`
	EndingBalance   *BankStatementBalanceDetail `json:"ending_balance"`
}

type BankTransactionEntry struct {
	NumberOfEntries string         `json:"number_of_entries"`
	Amount          *BalanceAmount `json:"amount"`
}

type BankStatementDetail struct {
	Amount            *BalanceAmount `json:"amount"`
	TransactionType   string         `json:"transaction_type"`
	TransactionRemark string         `json:"transaction_remark"`
	TransactionDate   string         `json:"transaction_date"`
}

type BankStatement struct {
	Balance            *BankStatementBalance  `json:"balance"`
	TotalCreditEntries *BankTransactionEntry  `json:"total_credit_entries"`
	TotalDebitEntries  *BankTransactionEntry  `json:"total_debit_entries"`
	Statements         []*BankStatementDetail `json:"statements"`
}

func (b BankStatement) Default() BankStatement {
	return BankStatement{
		Balance: &BankStatementBalance{
			Amount: &BankStatementBalanceDetail{
				BalanceDetails: BalanceAmount{}.Default(),
			},
			StartingBalance: &BankStatementBalanceDetail{
				BalanceDetails: BalanceAmount{}.Default(),
			},
			EndingBalance: &BankStatementBalanceDetail{
				BalanceDetails: BalanceAmount{}.Default(),
			},
		},
		TotalCreditEntries: &BankTransactionEntry{
			Amount: BalanceAmount{}.Default(),
		},
		TotalDebitEntries: &BankTransactionEntry{
			Amount: BalanceAmount{}.Default(),
		},
		Statements: []*BankStatementDetail{},
	}
}

func (b *BankStatement) FromBCAResponse(bca *BCABankStatementResponse) {
	b.TotalCreditEntries.NumberOfEntries = bca.TotalCreditEntries.NumberOfEntries
	b.TotalCreditEntries.Amount.Value = strings.TrimSpace(bca.TotalCreditEntries.Amount.Value)
	b.TotalCreditEntries.Amount.Currency = bca.TotalCreditEntries.Amount.Currency

	b.TotalDebitEntries.NumberOfEntries = bca.TotalDebitEntries.NumberOfEntries
	b.TotalDebitEntries.Amount.Value = strings.TrimSpace(bca.TotalDebitEntries.Amount.Value)
	b.TotalDebitEntries.Amount.Currency = bca.TotalDebitEntries.Amount.Currency

	if len(bca.Balance) == 0 {
		slog.Error("no balance found")
		return
	}

	var tempTime time.Time
	var err error
	balance := bca.Balance[0]

	tempTime, err = biUtil.ParseDateTime(balance.Amount.DateTime)
	if err != nil {
		slog.Error("failed to parse balance amount date time", "error", err)
		return
	}
	b.Balance.Amount.DateTime = tempTime.Format(time.DateTime)
	b.Balance.Amount.BalanceDetails.Value = strings.TrimSpace(balance.Amount.Value)
	b.Balance.Amount.BalanceDetails.Currency = balance.Amount.Currency

	tempTime, err = biUtil.ParseDateTime(balance.StartingBalance.DateTime)
	if err != nil {
		slog.Error("failed to parse starting balance date time", "error", err)
		return
	}
	b.Balance.StartingBalance.DateTime = tempTime.Format(time.DateTime)
	b.Balance.StartingBalance.BalanceDetails.Value = strings.TrimSpace(balance.StartingBalance.Value)
	b.Balance.StartingBalance.BalanceDetails.Currency = balance.StartingBalance.Currency

	tempTime, err = biUtil.ParseDateTime(balance.EndingBalance.DateTime)
	if err != nil {
		slog.Error("failed to parse ending balance date time", "error", err)
		return
	}
	b.Balance.EndingBalance.DateTime = tempTime.Format(time.DateTime)
	b.Balance.EndingBalance.BalanceDetails.Value = strings.TrimSpace(balance.EndingBalance.Value)
	b.Balance.EndingBalance.BalanceDetails.Currency = balance.EndingBalance.Currency

	// Loop through statements
	for _, item := range bca.DetailData {
		obj := BankStatementDetail{
			Amount: BalanceAmount{}.Default(),
		}
		tempTime, err = biUtil.ParseDateTime(item.TransactionDate)
		if err != nil {
			slog.Error("failed to parse transaction date", "error", err)
			return
		}
		obj.TransactionDate = tempTime.Format(time.DateTime)
		obj.TransactionRemark = item.Remark
		obj.TransactionType = item.Type
		obj.Amount.Value = strings.TrimSpace(item.Amount.Value)
		obj.Amount.Currency = item.Amount.Currency

		b.Statements = append(b.Statements, &obj)
	}
}
