package common

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page  int
	Limit int
}

// GetOffset calculates the offset for database queries
func (p PaginationParams) GetOffset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	return (p.Page - 1) * p.Limit
}

// Validate validates pagination parameters and sets defaults
func (p *PaginationParams) Validate() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 || p.Limit > 100 {
		p.Limit = 20
	}
}

// PaginatedResult represents a paginated result
type PaginatedResult struct {
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
	Total int64       `json:"total"`
	Data  interface{} `json:"data"`
}
