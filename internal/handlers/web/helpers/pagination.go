package handler_helpers

import (
	"github.com/gin-gonic/gin"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
)

func ParsePagination(c *gin.Context) (*entities.PaginationData, error) {
	var p entities.PaginationData

	if err := c.ShouldBindQuery(&p); err != nil {
		return nil, err
	}

	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}

	return &p, nil
}

func ParseParamsWithPagination(c *gin.Context, p PaginationData) error {
	if err := c.ShouldBindQuery(p); err != nil {
		return err
	}

	if p.GetPage() < 1 {
		p.SetPage(1)
	}

	pageSize := p.GetPageSize()
	if pageSize < 1 || pageSize > 100 {
		p.SetPageSize(20)
	}

	return nil
}
