package user

import (
	"devcloud/config"
	"devcloud/ent"
)

func New(cfg *config.Config, db *ent.Client) *Handler {
	userRepo := UserRepo{user: db.User}
	tokenRepo := TokenRepo{token: db.Token}

	service := &Service{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		options: &ServiceOptions{
			JWTSecret:                 cfg.JWTSecret,
			TokenSalt:                 cfg.TokenSalt,
			AccessTokenExpireSeconds:  cfg.AccessTTL,
			RefreshTokenExpireSeconds: cfg.RefreshTTL,
		},
	}

	return &Handler{srv: service, cookieSecure: *cfg.CookieSecure}
}
