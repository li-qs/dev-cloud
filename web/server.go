package web

import (
	"bytes"
	"devcloud/config"
	"devcloud/ent"
	"devcloud/web/domain/health"
	"devcloud/web/domain/user"
	myMiddleware "devcloud/web/middleware"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func NewServer(cfg *config.Config, db *ent.Client) *echo.Echo {
	e := echo.NewWithConfig(echo.Config{
		JSONSerializer: &JSONSerializer{},
		Validator:      &Validator{validator: validator.New()},
	})

	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	healthHandler := health.New(db)
	userHandler := user.New(cfg, db)

	{
		e.GET("/health", healthHandler.Liveness)
		e.GET("/ready", healthHandler.Readiness)
	}

	api := e.Group("/api")
	{
		api.POST("/login", userHandler.Login, middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
			Store: middleware.NewRateLimiterMemoryStore(10),
			IdentifierExtractor: func(c *echo.Context) (string, error) {
				return c.RealIP(), nil
			},
		}))
		api.POST("/logout", userHandler.Logout)
		api.POST("/refresh", userHandler.RefreshToken)
	}

	authApi := e.Group("/api")
	authApi.Use(myMiddleware.Auth(cfg.JWTSecret))
	{
		authApi.GET("/user", userHandler.UserInfo)
		authApi.PUT("/user/password", userHandler.UpdatePassword)
	}

	return e
}

type Validator struct {
	validator *validator.Validate
}

func (v *Validator) Validate(i any) error {
	if err := v.validator.Struct(i); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}
	return nil
}

const maxPooledJSONBuf = 1 << 16 // 64 KiB

var jsonBufPool = sync.Pool{New: newJSONBuf}

func newJSONBuf() any { return new(bytes.Buffer) }

type JSONSerializer struct{}

func (s *JSONSerializer) Serialize(c *echo.Context, target any, indent string) error {
	enc := sonic.ConfigFastest.NewEncoder(c.Response())
	if indent != "" {
		enc.SetIndent("", indent)
	}
	return enc.Encode(target)
}

func (s *JSONSerializer) Deserialize(c *echo.Context, target any) error {
	buf := jsonBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer func() {
		if buf.Cap() <= maxPooledJSONBuf {
			jsonBufPool.Put(buf)
		}
	}()

	if _, err := buf.ReadFrom(c.Request().Body); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}
	if err := sonic.Unmarshal(buf.Bytes(), target); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}
	return nil
}
