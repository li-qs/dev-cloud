package reqctx

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v5"
)

type User struct {
	ID        int
	Username  string
	Nickname  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

const userKey = "user"

func GetUser(c *echo.Context) (*User, error) {
	user, ok := c.Get(userKey).(*User)
	if !ok || user.ID == 0 || user.Username == "" {
		return nil, fmt.Errorf("invalid token")
	}
	return user, nil
}

func SetUser(c *echo.Context, user *User) {
	c.Set(userKey, user)
}
