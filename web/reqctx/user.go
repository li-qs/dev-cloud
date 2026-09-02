package reqctx

import (
	"time"

	"github.com/labstack/echo/v5"
)

type UserCtx struct {
	ID           int
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

const userKey = "user"

func GetUser(c *echo.Context) (user *UserCtx, ok bool) {
	user, ok = c.Get(userKey).(*UserCtx)
	return
}

func SetUser(c *echo.Context, user *UserCtx) {
	c.Set(userKey, user)
}
