package bank_integration_models

type InternalVAInformation struct {
	IDBank        uint   `json:"id_bank"`
	BankName      string `json:"bank_name"`
	BankIconLink  string `json:"bank_icon_link"`
	VANumber      string `json:"va_number"`
	VAAccountName string `json:"va_account_name"`
	TotalAmount   string `json:"total_amount"`
	ExpiredAt     string `json:"expired_at"`
}
