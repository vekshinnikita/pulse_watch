package entities

type PaginationResult[T any] struct {
	Page  int `json:"page"`
	Total int `json:"total"`
	Items []T `json:"items"`
}

type PaginationData struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

func (p *PaginationData) GetPage() int {
	return p.Page
}

func (p *PaginationData) GetPageSize() int {
	return p.PageSize
}

func (p *PaginationData) SetPage(page int) {
	p.Page = page
}

func (p *PaginationData) SetPageSize(pageSize int) {
	p.PageSize = pageSize
}

func (p *PaginationData) Offset() int {
	return p.PageSize * (p.Page - 1)
}

func (p *PaginationData) Limit() int {
	return p.PageSize
}
