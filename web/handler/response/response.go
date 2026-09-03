package response

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
}

type ListData[T any] struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
	List     []T `json:"list"`
}

func JsonSuccess(c *echo.Context) error {
	return JsonData[any](c, nil)
}

func JsonData[T any](c *echo.Context, data T) error {
	return c.JSON(http.StatusOK, &Response[T]{
		Code: http.StatusOK,
		Data: data,
	})
}

func JsonList[T any](c *echo.Context, list []T, page int, pageSize int, total int) error {
	return c.JSON(http.StatusOK, &Response[ListData[T]]{
		Code: http.StatusOK,
		Data: ListData[T]{
			List:     list,
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}
