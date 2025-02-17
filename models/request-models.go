package bank_integration_models

type BankRequestHandlerResponse struct {
	StatusCode     uint
	ResponseBody   string
	ResponseHeader string
	Log            *BankLogV2
}
