package middleware

import (
	"strings"

	"devcloud/web/handler/reqctx"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

func Auth(jwtSecret string) echo.MiddlewareFunc {
	key := []byte(jwtSecret)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return echo.ErrUnauthorized
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return key, nil
			})
			if err != nil || !token.Valid {
				return echo.ErrUnauthorized
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.ErrUnauthorized
			}

			uid, ok := claims["uid"].(float64)
			if !ok {
				return echo.ErrUnauthorized
			}

			username, ok := claims["username"].(string)
			if !ok {
				return echo.ErrUnauthorized
			}

			if uid == 0 || username == "" {
				return echo.ErrUnauthorized
			}

			reqctx.SetUser(c, &reqctx.User{
				ID:       int(uid),
				Username: username,
			})

			return next(c)
		}
	}
}
