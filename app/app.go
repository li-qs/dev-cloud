package app

import (
	"devcloud/config"
	"devcloud/ent"
	"devcloud/repo"
	"devcloud/service"

	"github.com/redis/go-redis/v9"
)

type App struct {
	UserService     *service.User
	ResourceService *service.Resource
	TaskService     *service.Task
}

func New(cfg *config.Config, db *ent.Client, rdb *redis.Client) *App {
	userRepo := repo.NewUser(db)
	tokenRepo := repo.NewToken(db)
	resourceRepo := repo.NewResource(db)
	taskRepo := repo.NewTask(db)

	userService := service.NewUser(
		service.UserOptions{
			JWTSecret:                 cfg.JWTSecret,
			TokenSalt:                 cfg.TokenSalt,
			AccessTokenExpireSeconds:  cfg.AccessTTL,
			RefreshTokenExpireSeconds: cfg.RefreshTTL,
		},
		userRepo,
		tokenRepo,
	)
	resourceService := service.NewResource(resourceRepo, taskRepo)
	taskService := service.NewTask(taskRepo)

	return &App{
		UserService:     userService,
		ResourceService: resourceService,
		TaskService:     taskService,
	}
}
