package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

type HealthCheck struct {
	Name  string
	Check func(ctx context.Context) error
}

type Health struct {
	checks []HealthCheck
}

func NewHealth(checks ...HealthCheck) *Health {
	return &Health{checks: checks}
}

func (h *Health) Liveness(c *echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

type readinessResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

func (h *Health) Readiness(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	out := readinessResponse{
		Status: "ok",
		Checks: make(map[string]string, len(h.checks)),
	}
	statusCode := http.StatusOK

	for _, chk := range h.checks {
		if err := chk.Check(ctx); err != nil {
			out.Checks[chk.Name] = "down: " + err.Error()
			out.Status = "degraded"
			statusCode = http.StatusServiceUnavailable
			continue
		}
		out.Checks[chk.Name] = "ok"
	}

	return c.JSON(statusCode, out)
}
