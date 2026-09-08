package web

import (
	"errors"
	"log/slog"
	"net/http"

	"devcloud/ent"
	"devcloud/web/handler/errmsg"
	"devcloud/web/handler/response"

	"github.com/labstack/echo/v5"
)

func HTTPErrorHandler(c *echo.Context, err error) {
	status := http.StatusInternalServerError
	message := "internal server error"

	var he *echo.HTTPError
	if errors.As(err, &he) {
		status = he.Code
	}

	switch {
	case ent.IsNotFound(err):
		status = http.StatusNotFound
		message = "resource not found"

	case errors.Is(err, errmsg.ErrInvalidCredentials):
		status = http.StatusUnauthorized
		message = "invalid username or password"

	case errors.Is(err, errmsg.ErrInvalidToken):
		status = http.StatusUnauthorized
		message = "invalid token"

	case errors.Is(err, errmsg.ErrResourceNotFound):
		status = http.StatusNotFound
		message = "resource not found"

	case errors.Is(err, errmsg.ErrPermissionDenied):
		status = http.StatusForbidden
		message = "permission denied"

	case errors.Is(err, errmsg.ErrResourceState):
		status = http.StatusConflict
		message = "invalid resource state"
	}

	if status >= 500 {
		slog.Error("request failed",
			"error", err,
			"method", c.Request().Method,
			"path", c.Path(),
		)
	}

	_ = c.JSON(status, response.Response[any]{
		Code:    status,
		Message: message,
	})
}
