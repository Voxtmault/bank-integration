package bank_integration_models

type AuthenticatedBank struct {
	ID            uint   `json:"id"`
	BankName      string `json:"bank_name"`
	PublicKeyPath string `json:"public_key_path"`
	Note          string `json:"note"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type BankClientCredential struct {
	ClientID      string `json:"client_id"`
	ClientSecret  string `json:"client_secret"`
	PublicKeyPath string `json:"public_key_path"`
}
