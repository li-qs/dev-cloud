package service

import (
	"context"
	"devcloud/ent"
	"devcloud/repo"
	"devcloud/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	repo    *repo.Repo
	Options UserOptions
}

type UserOptions struct {
	JWTSecret                 string
	TokenSalt                 string
	AccessTokenExpireSeconds  int
	RefreshTokenExpireSeconds int
}

func NewUser(
	repo *repo.Repo,
	options UserOptions,
) *User {
	return &User{
		repo:    repo,
		Options: options,
	}
}

type jwtClaims struct {
	jwt.RegisteredClaims
	UserID   int    `json:"uid"`
	Username string `json:"username"`
}

func (s *User) AuthUser(ctx context.Context, username, password string) (user *ent.User, valid bool, err error) {
	valid = false
	user, err = s.repo.User.GetByUsername(ctx, username)
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
	if err := s.repo.RefreshToken.Create(ctx, user.ID, hash, expiresAt); err != nil {
		return "", "", 0, err
	}

	return accessToken, refreshRaw, s.Options.AccessTokenExpireSeconds, nil
}

type RefreshToken struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

func (s *User) RefreshTokens(ctx context.Context, refreshRaw string) (token *RefreshToken, valid bool, err error) {
	tokenHash := s.hashRefreshToken(refreshRaw)
	rt, err := s.repo.RefreshToken.Get(ctx, tokenHash)
	if err != nil {
		return
	}
	if time.Now().After(rt.ExpiresAt) {
		s.repo.RefreshToken.Delete(ctx, rt.ID)
		return
	}

	user, err := s.repo.User.Get(ctx, rt.UserID)
	if err != nil {
		return
	}

	accessToken, newRefreshRaw, expiresIn, err := s.GenerateTokens(ctx, user)
	if err != nil {
		return
	}

	s.repo.RefreshToken.Delete(ctx, rt.ID)

	return &RefreshToken{
		AccessToken:  accessToken,
		RefreshToken: newRefreshRaw,
		ExpiresIn:    expiresIn,
	}, true, nil
}

func (s *User) Logout(ctx context.Context, refreshRaw string) error {
	tokenHash := s.hashRefreshToken(refreshRaw)
	rt, err := s.repo.RefreshToken.Get(ctx, tokenHash)
	if err != nil {
		return err
	}
	return s.repo.RefreshToken.Delete(ctx, rt.ID)
}

func (s *User) RevokeAllUserTokens(ctx context.Context, userID int) error {
	return s.repo.RefreshToken.DeleteByUserID(ctx, userID)
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
	return s.repo.User.UpdatePassword(ctx, userID, string(pwdHash))
}

func (s *User) UpdateNickname(ctx context.Context, userID int, nickname string) error {
	return s.repo.User.UpdateNickname(ctx, userID, nickname)
}
