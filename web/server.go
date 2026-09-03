package web

import (
	"devcloud/app"
	"devcloud/web/handler"
	"devcloud/web/middleware"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
)

func NewServer(app *app.App, cookieSecure bool, jwtSecret string) *echo.Echo {
	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: HTTPErrorHandler,
		JSONSerializer:   &JSONSerializer{},
		Validator:        &Validator{validator: validator.New()},
	})

	e.Use(echomw.RequestID())
	e.Use(echomw.RequestLogger())
	e.Use(echomw.Recover())

	loginLimiter := echomw.RateLimiterWithConfig(echomw.RateLimiterConfig{
		Store: echomw.NewRateLimiterMemoryStore(10),
		IdentifierExtractor: func(c *echo.Context) (string, error) {
			return c.RealIP(), nil
		},
	})

	health := handler.NewHealth()
	user := handler.NewUser(cookieSecure, app.UserService)
	resource := handler.NewResource(app.ResourceService)
	task := handler.NewTask(app.TaskService)

	{
		e.GET("/health", health.Liveness)
	}

	api := e.Group("/api")
	{
		api.POST("/login", user.Login, loginLimiter)
		api.POST("/refresh", user.RefreshToken)
	}

	authApi := api.Group("")
	authApi.Use(middleware.Auth(jwtSecret))
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
