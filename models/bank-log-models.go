package bank_integration_models

type BankLog struct {
	ID               uint    `json:"id"`
	HostIP           string  `json:"host_ip"`
	ClientIP         string  `json:"client_ip"`
	Latency          float32 `json:"latency"`
	HTTPMethod       string  `json:"http_method"`
	Protocol         string  `json:"protocol"`
	URI              string  `json:"uri"`
	RequestParameter string  `json:"request_parameter"` // In JSON Format
	RequestBody      string  `json:"request_body"`      // In JSON Format
	ResponseCode     uint    `json:"response_code"`
	ResponseMessage  string  `json:"response_message"`
	ResponseContent  string  `json:"response_content"` // In JSON Format
	CreatedAt        string  `json:"created_at"`
}
