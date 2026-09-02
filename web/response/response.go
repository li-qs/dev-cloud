package response

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

const CodeOK = 0

type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
}

type ListData[T any] struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
	List     T   `json:"list"`
}

func JsonData[T any](c *echo.Context, data T) error {
	return c.JSON(http.StatusOK, &Response[T]{
		Code: CodeOK,
		Data: data,
	})
}

func JsonList[T any](c *echo.Context, list T, page int, pageSize int, total int) error {
	return c.JSON(http.StatusOK, &Response[ListData[T]]{
		Code: CodeOK,
		Data: ListData[T]{
			List:     list,
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

func JsonError(c *echo.Context, code int, msg string) error {
	return c.JSON(http.StatusOK, &Response[any]{
		Code:    code,
		Message: msg,
	})
}
