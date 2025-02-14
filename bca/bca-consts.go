package bca

const (
	TransactionPurposePurchaseOfGoods               = "01" // Pembelian barang
	TransactionPurposePurchaseOfService             = "02" // Pembayaran jasa
	TransactionPurposeDebtPayment                   = "03" // Pembayaran hutang
	TransactionPurposeTaxPayment                    = "04" // Pembayaran pajak
	TransactionPurposeImport                        = "05" // Impor
	TransactionPurposeBusinessTravelExpenses        = "06" // Biaya perjalanan luar negeri
	TransactionPurposeRenumerationExpenses          = "07" // Biaya renumerasi pegawai
	TransactionPurposeAdministrationFees            = "08" // Biaya administrasi
	TransactionPurposeOverhedExpenses               = "09" // Biaya overhead
	TransactionPurposeSaving                        = "10" // Simpanan dalam valas
	TransactionPurposePurchaseOfCorporateBonds      = "11" // Investasi pembelian obligasi
	TransactionPurposePurchaseOfGovermentSecurities = "12" // Investasi pembelian surat berharga negara	(SBN)
	TransactionPurposeDirectCapitalInvestment       = "13" // Investasi penyertaan langsung
	TransactionPurposeWorkingCapital                = "14" // Penambahan modal kerja
	TransactionPurposeApprovedByBI                  = "15" // Telah mendapat izin dari Bank Indonesia (BI)
)

// To be used for BCA Intterbank Transfers
const (
	BCAInterbankPurposeInvestment       = "01"
	BCAInterbankPurposeTransferOfWealth = "02"
	BCAInterbankPurposePurchase         = "03"
	BCAInterbankPurposeOthers           = "99"

	BCAInterbankSwitcher = "1"
	BCAInterbankBiFAST   = "2"
)

// To be used for BCA Transfer Status Inquiry
const (
	BCAServiceIntrabankTransfer     = "17"
	BCAServiceInterbankTransfer     = "18"
	BCAServiceInterbankTransferRTGS = "22"
	BCAServiceInterbankTransferSKN  = "23"
	BCAServicePaymentToVAIntrabank  = "33"
)
