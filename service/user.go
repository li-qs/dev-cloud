package service

import (
	"context"
	"devcloud/ent"
	"devcloud/repo"
	"devcloud/utils"
	"devcloud/web/errmsg"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	userRepo  *repo.User
	tokenRepo *repo.RefreshToken

	Options UserOptions
}

type UserOptions struct {
	JWTSecret                 string
	TokenSalt                 string
	AccessTokenExpireSeconds  int
	RefreshTokenExpireSeconds int
}

const (
	UserStatusEnabled  = 1
	UserStatusDisabled = 0
)

func NewUser(
	options UserOptions,
	userRepo *repo.User,
	tokenRepo *repo.RefreshToken,
) *User {
	return &User{
		Options:   options,
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
	}
}

type jwtClaims struct {
	jwt.RegisteredClaims
	UserID   int    `json:"uid"`
	Username string `json:"username"`
}

func (s *User) AuthUser(ctx context.Context, username, password string) (user *ent.User, valid bool, err error) {
	valid = false
	user, err = s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return
	}

	if cmpErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); cmpErr != nil {
		return
	}

	valid = true
	return
}

func (s *User) GenerateTokens(ctx context.Context, user *ent.User) (string, string, int, error) {
	now := time.Now()
	accessToken, err := s.generateJWT(user, now)
	if err != nil {
		return "", "", 0, err
	}

	refreshRaw, err := utils.RandomString(32)
	if err != nil {
		return "", "", 0, err
	}

	hash := s.hashRefreshToken(refreshRaw)
	expiresAt := now.Add(time.Duration(s.Options.RefreshTokenExpireSeconds) * time.Second)
	if err := s.tokenRepo.Create(ctx, user.ID, hash, expiresAt); err != nil {
		return "", "", 0, err
	}

	return accessToken, refreshRaw, s.Options.AccessTokenExpireSeconds, nil
}

func (s *User) RefreshTokens(ctx context.Context, refreshRaw string) (string, string, int, error) {
	tokenHash := s.hashRefreshToken(refreshRaw)
	rt, err := s.tokenRepo.GetByToken(ctx, tokenHash)
	if err != nil {
		return "", "", 0, fmt.Errorf("find refresh token: %w", err)
	}
	if time.Now().After(rt.ExpiresAt) {
		s.tokenRepo.Delete(ctx, rt.ID)
		return "", "", 0, errmsg.ErrInvalidToken
	}

	user, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return "", "", 0, err
	}

	accessToken, newRefreshRaw, expiresIn, err := s.GenerateTokens(ctx, user)
	if err != nil {
		return "", "", 0, err
	}

	_ = s.tokenRepo.Delete(ctx, rt.ID)

	return accessToken, newRefreshRaw, expiresIn, nil
}

func (s *User) Logout(ctx context.Context, refreshRaw string) error {
	tokenHash := s.hashRefreshToken(refreshRaw)
	rt, err := s.tokenRepo.GetByToken(ctx, tokenHash)
	if err != nil {
		return err
	}
	return s.tokenRepo.Delete(ctx, rt.ID)
}

func (s *User) RevokeAllUserTokens(ctx context.Context, userID int) error {
	return s.tokenRepo.DeleteByUserID(ctx, userID)
}

func (s *User) generateJWT(user *ent.User, now time.Time) (string, error) {
	claims := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.Options.AccessTokenExpireSeconds) * time.Second)),
		},
		UserID:   user.ID,
		Username: user.Username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.Options.JWTSecret))
}

func (s *User) hashRefreshToken(raw string) string {
	return utils.HMACSHA256Hex(s.Options.TokenSalt, raw)
}

func (s *User) UpdatePassword(ctx context.Context, userID int, pwd string) error {
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
	if err != nil {
		return err
	}
	return s.userRepo.UpdatePassword(ctx, userID, string(pwdHash))
}

func (s *User) UpdateNickname(ctx context.Context, userID int, nickname string) error {
	return s.userRepo.UpdateNickname(ctx, userID, nickname)
}
