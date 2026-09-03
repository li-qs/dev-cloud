package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

type Health struct{}

func NewHealth() *Health {
	return &Health{}
}

func (h *Health) Liveness(c *echo.Context) error {
	return c.String(http.StatusOK, "OK")
}
