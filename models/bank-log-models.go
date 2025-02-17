package bank_integration_models

import (
	"strconv"
	"time"

	biCache "github.com/voxtmault/bank-integration/cache"
)

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

type BankLogV2 struct {
	IDBank       uint   `json:"id_bank" validate:"required,number,gte=1,min=1"`
	IDFeature    uint   `json:"id_feature" validate:"required,number,gte=1,min=1"`
	Latency      string `json:"latency" validate:"omitempty"`
	ResponseCode uint   `json:"response_code" validate:"required,number"`
	HostIP       string `json:"host_ip" validate:"omitempty,ipv4"`
	ClientIP     string `json:"client_ip" validate:"omitempty,ipv4"`
	HTTPMethod   string `json:"http_method" validate:"required"`
	Protocol     string `json:"protocol" validate:"omitempty"`
	URI          string `json:"uri" validate:"required,uri"`

	// In JSON Format

	RequestHeader  string `json:"request_header" validate:"json"`
	RequestBody    string `json:"request_body" validate:"json"`
	ResponseHeader string `json:"response_header" validate:"json"`
	ResponseBody   string `json:"response_body" validate:"json"`

	// To calculate latency

	BeginAt time.Time `json:"-"`
	EndAt   time.Time `json:"-"`
}

type BankLogPublic struct {
	ID              uint                   `json:"id"`
	Bank            *Helper                `json:"bank"`
	RelatedFeature  *Helper                `json:"related_feature"`
	Latency         string                 `json:"latency"`
	ResponseCode    uint                   `json:"response_code"`
	HostIP          string                 `json:"host_ip,omitempty"`
	ClientIP        string                 `json:"client_ip,omitempty"`
	HTTPMethod      string                 `json:"http_method"`
	Protocol        string                 `json:"protocol"`
	URI             string                 `json:"uri"`
	RequestHeader   map[string]interface{} `json:"request_header"`
	RequestBody     map[string]interface{} `json:"request_body"`
	ResponseHeader  map[string]interface{} `json:"response_header"`
	ResponsePayload map[string]interface{} `json:"response_payload"`
	CreatedAt       string                 `json:"created_at"`
}

func (b *BankLogPublic) GetHelperName() {
	var ok bool
	if b.Bank.Name, ok = biCache.PartneredBanksMap[strconv.Itoa(int(b.Bank.ID))]; !ok {
		b.Bank.Name = ""
	}
	if b.RelatedFeature.Name, ok = biCache.BankFeaturesMap[strconv.Itoa(int(b.RelatedFeature.ID))]; !ok {
		b.RelatedFeature.Name = ""
	}
}

type BankLogSearchFilter struct {
	ID             uint   `query:"id" validate:"omitempty,number,gte=1,min=1" example:"1"`
	IDBank         uint   `query:"id_bank" validate:"omitempty,number,gte=1,min=1" example:"1"`
	IDFeature      uint   `query:"id_feature" validate:"omitempty,number,gte=1,min=1" example:"1"`
	ResponseCode   uint   `query:"response_code" validate:"omitempty,number" example:"200"`
	HostIP         string `query:"host_ip" validate:"omitempty,ipv4" example:"192.168.1.1"`
	ClientIP       string `query:"client_ip" validate:"omitempty,ipv4" example:"10.0.0.1"`
	HTTPMethod     string `query:"http_method" validate:"omitempty" example:"GET"`
	URI            string `query:"uri" validate:"omitempty,uri" example:"/api/v1/bank"`
	StartDateRange string `query:"start_date_range" validate:"omitempty,datetime=2006-01-02" example:"2021-01-01"`
	EndDateRange   string `query:"end_date_range" validate:"omitempty,datetime=2006-01-02" example:"2021-01-01"`

	PaginationFilter
}
