package health

import (
	"context"
	"devcloud/ent"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	db *ent.Client
}

func New(db *ent.Client) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Liveness(c *echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (h *Handler) Readiness(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	_, err := h.db.User.Query().Limit(1).All(ctx)
	if err != nil {
		return c.String(http.StatusServiceUnavailable, "unavailable")
	}
	return c.String(http.StatusOK, "OK")
}
