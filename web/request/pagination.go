package request

import (
	"strconv"

	"github.com/labstack/echo/v5"
)

const (
	minPageSize = 10
	maxPageSize = 100
)

func ParsePagination(c *echo.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.QueryParam("page"))
	pageSize, _ = strconv.Atoi(c.QueryParam("page_size"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = minPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return
}
