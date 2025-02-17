package bank_integration_models

type PaginationMetadata struct {
	TotalRecords uint `json:"total_records"`
	TotalPages   uint `json:"total_pages"`
	CurrentLimit uint `json:"page_size"`
	CurrentPage  uint `json:"current_page"`
}

type PaginationFilter struct {
	Limit      uint `query:"page_size" validate:"required,number,min=1"`
	PageNumber uint `query:"page_number" validate:"required,number,min=1"`
}
