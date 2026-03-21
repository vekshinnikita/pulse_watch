package handler_helpers

type PaginationData interface {
	GetPage() int
	GetPageSize() int
	SetPage(page int)
	SetPageSize(pageSize int)
}
