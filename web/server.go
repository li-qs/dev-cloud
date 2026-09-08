package web

import (
	"devcloud/config"
	"devcloud/repo"
	"devcloud/store"
	"devcloud/web/handler"
	"devcloud/web/middleware"
	"devcloud/web/service"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
)

func NewServer(cfg *config.Config, repo *repo.Repo, store *store.Store, checks []handler.HealthCheck) *echo.Echo {
	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: HTTPErrorHandler,
		JSONSerializer:   &JSONSerializer{},
		Validator:        &Validator{validator: validator.New()},
	})

	e.Use(echomw.RequestID())
	e.Use(echomw.RequestLogger())
	e.Use(echomw.Recover())
	e.Use(echomw.BodyLimit(1 << 20))

	loginLimiter := echomw.RateLimiterWithConfig(echomw.RateLimiterConfig{
		Store: echomw.NewRateLimiterMemoryStore(10),
		IdentifierExtractor: func(c *echo.Context) (string, error) {
			return c.RealIP(), nil
		},
	})

	userSrv := service.NewUser(
		repo,
		service.UserOptions{
			JWTSecret:                 cfg.JWTSecret,
			TokenSalt:                 cfg.TokenSalt,
			AccessTokenExpireSeconds:  cfg.AccessTTL,
			RefreshTokenExpireSeconds: cfg.RefreshTTL,
		},
	)
	resourceSrv := service.NewResource(repo)
	taskSrv := service.NewTask(repo)

	health := handler.NewHealth(checks...)
	user := handler.NewUser(*cfg.CookieSecure, userSrv)
	resource := handler.NewResource(resourceSrv)
	task := handler.NewTask(taskSrv)

	{
		e.GET("/health", health.Liveness)
		e.GET("/ready", health.Readiness)
	}

	api := e.Group("/api")
	{
		api.POST("/login", user.Login, loginLimiter)
		api.POST("/refresh", user.RefreshToken)
	}

	authApi := api.Group("")
	authApi.Use(middleware.Auth(cfg.JWTSecret))
	{
		authApi.POST("/logout", user.Logout)
		authApi.PUT("/user/password", user.UpdatePassword)
		authApi.GET("/user", user.UserInfo)

		authApi.GET("/resources", resource.List)
		authApi.POST("/resources", resource.Create)
		authApi.GET("/resources/:id", resource.Get)
		authApi.DELETE("/resources/:id", resource.Delete)

		authApi.POST("/resources/:id/start", resource.Start)
		authApi.POST("/resources/:id/stop", resource.Stop)
		authApi.POST("/resources/:id/restart", resource.Restart)

		authApi.GET("/tasks/:id", task.Get)
	}

	return e
}
