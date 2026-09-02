package user

import (
	"context"
	"devcloud/ent"
	"devcloud/pkg/utils"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	userRepo  UserRepo
	tokenRepo TokenRepo
	options   *ServiceOptions
}

type ServiceOptions struct {
	JWTSecret                 string
	TokenSalt                 string
	AccessTokenExpireSeconds  int
	RefreshTokenExpireSeconds int
}

type jwtClaims struct {
	jwt.RegisteredClaims
	UserID   int    `json:"uid"`
	Username string `json:"username"`
}

func (s *Service) AuthUser(ctx context.Context, username, password string) (*ent.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (s *Service) GenerateTokens(ctx context.Context, user *ent.User) (string, string, int, error) {
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
	expiresAt := now.Add(time.Duration(s.options.RefreshTokenExpireSeconds) * time.Second)
	if err := s.tokenRepo.Create(ctx, user.ID, hash, expiresAt); err != nil {
		return "", "", 0, err
	}

	return accessToken, refreshRaw, s.options.AccessTokenExpireSeconds, nil
}

func (s *Service) RefreshTokens(ctx context.Context, refreshRaw string) (string, string, int, error) {
	tokenHash := s.hashRefreshToken(refreshRaw)
	rt, err := s.tokenRepo.GetByToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", "", 0, ErrInvalidRefreshToken
		}
		return "", "", 0, fmt.Errorf("find refresh token: %w", err)
	}
	if time.Now().After(rt.ExpiresAt) {
		s.tokenRepo.Delete(ctx, rt.ID)
		return "", "", 0, ErrInvalidRefreshToken
	}

	user, err := s.userRepo.GetByID(ctx, rt.UID)
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

func (s *Service) Logout(ctx context.Context, refreshRaw string) error {
	tokenHash := s.hashRefreshToken(refreshRaw)
	rt, err := s.tokenRepo.GetByToken(ctx, tokenHash)
	if err != nil {
		return err
	}
	return s.tokenRepo.Delete(ctx, rt.ID)
}

func (s *Service) RevokeAllUserTokens(ctx context.Context, userID int) error {
	return s.tokenRepo.DeleteByUserID(ctx, userID)
}

func (s *Service) generateJWT(user *ent.User, now time.Time) (string, error) {
	claims := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.options.AccessTokenExpireSeconds) * time.Second)),
		},
		UserID:   user.ID,
		Username: user.Username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.options.JWTSecret))
}

func (s *Service) hashRefreshToken(raw string) string {
	return utils.HMACSHA256Hex(s.options.TokenSalt, raw)
}

func (s *Service) UpdatePassword(ctx context.Context, userID int, pwd string) error {
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
	if err != nil {
		return err
	}
	return s.userRepo.UpdatePasswordHash(ctx, userID, string(pwdHash))
}
