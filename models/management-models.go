package bank_integration_models

import (
	"strconv"

	biCache "github.com/voxtmault/bank-integration/cache"
)

type BankClientCredential struct {
	ID               uint    `json:"id"`
	Bank             *Helper `json:"bank"`
	ClientID         string  `json:"client_id"`
	ClientSecret     string  `json:"client_secret"`
	CredentialStatus bool    `json:"credential_status"`
	CredentialNote   string  `json:"credential_note"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

func (c BankClientCredential) Default() BankClientCredential {
	return BankClientCredential{
		Bank: &Helper{},
	}
}

func (c *BankClientCredential) GetHelperName() {
	var ok bool
	if c.Bank.Name, ok = biCache.PartneredBanksMap[strconv.Itoa(int(c.Bank.ID))]; !ok {
		c.Bank.Name = ""
	}
}

type PartneredBank struct {
	ID                 uint                    `json:"id"`
	BankName           string                  `json:"bank_name"`
	DefaultPicturePath string                  `json:"default_picture_path"`
	PartnershipStatus  bool                    `json:"partnership_status"`
	PaymentMethod      []*PaymentMethod        `json:"payment_method"`
	IntegratedFeature  []*IntegratedFeature    `json:"integrated_feature"`
	ClientCredentials  []*BankClientCredential `json:"client_credentials"`
	CreatedAt          string                  `json:"created_at"`
	UpdatedAt          string                  `json:"updated_at"`
}

func (p PartneredBank) Default() PartneredBank {
	return PartneredBank{
		PaymentMethod:     []*PaymentMethod{},
		IntegratedFeature: []*IntegratedFeature{},
	}
}

type PartneredBankAdd struct {
	BankName           string `json:"bank_name" validate:"required"`
	DefaultPicturePath string `json:"default_picture_path"`
}

type PaymentMethod struct {
	ID                  uint    `json:"id"`
	Bank                *Helper `json:"bank"`
	PaymentMethod       *Helper `json:"payment_method"`
	DefaultPicturePath  string  `json:"default_picture_path"`
	PaymentMethodStatus bool    `json:"payment_method_status"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
}

func (m PaymentMethod) Default() PaymentMethod {
	return PaymentMethod{
		Bank:          &Helper{},
		PaymentMethod: &Helper{},
	}
}

func (m *PaymentMethod) GetHelperName() {
	var ok bool
	if m.Bank.Name, ok = biCache.PartneredBanksMap[strconv.Itoa(int(m.Bank.ID))]; !ok {
		m.Bank.Name = ""
	}
	if m.PaymentMethod.Name, ok = biCache.PaymentMethodsMap[strconv.Itoa(int(m.PaymentMethod.ID))]; !ok {
		m.PaymentMethod.Name = ""
	}
}

type PaymentMethodAdd struct {
	IDBank              uint   `json:"id_bank" validate:"required,number,gte=1,min=1"`
	IDPaymentMethod     uint   `json:"id_payment_method" validate:"required,number,gte=1,min=1"`
	PaymentMethodStatus bool   `json:"payment_method_status" validate:"omitempty"`
	DefaultPicturePath  string `json:"default_picture_path"`
}

type IntegratedFeature struct {
	ID          uint    `json:"id"`
	Bank        *Helper `json:"bank"`
	Feature     *Helper `json:"feature"`
	FeatureType *Helper `json:"feature_type"`
	Note        string  `json:"note"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func (f IntegratedFeature) Default() IntegratedFeature {
	return IntegratedFeature{
		Bank:        &Helper{},
		Feature:     &Helper{},
		FeatureType: &Helper{},
	}
}

func (f *IntegratedFeature) GetHelperName() {
	var ok bool
	if f.Bank.Name, ok = biCache.PartneredBanksMap[strconv.Itoa(int(f.Bank.ID))]; !ok {
		f.Bank.Name = ""
	}
	if f.Feature.Name, ok = biCache.BankFeaturesMap[strconv.Itoa(int(f.Feature.ID))]; !ok {
		f.Feature.Name = ""
	}
	if f.FeatureType.Name, ok = biCache.BankFeatureTypesMap[strconv.Itoa(int(f.FeatureType.ID))]; !ok {
		f.FeatureType.Name = ""
	}
}

type IntegratedFeatureAdd struct {
	IDBank        uint   `json:"id_bank" validate:"required,number,gte=1,min=1"`
	IDFeature     uint   `json:"id_feature" validate:"required,number,gte=1,min=1"`
	IDFeatureType uint   `json:"id_feature_type" validate:"required,number,gte=1,min=1"`
	Note          string `json:"note" validate:"omitempty"`
	Status        string `json:"status" validate:"omitempty"`
}
