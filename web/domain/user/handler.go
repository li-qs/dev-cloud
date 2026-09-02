package user

import (
	"context"
	"devcloud/web/reqctx"
	"devcloud/web/response"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	srv          *Service
	cookieSecure bool
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (h *Handler) Login(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.JsonError(c, 400, "用户名或密码错误")
	}
	if err := c.Validate(&req); err != nil {
		return response.JsonError(c, 400, "用户名或密码错误")
	}

	user, err := h.srv.AuthUser(ctx, req.Username, req.Password)
	if err != nil {
		slog.Error("AuthUser failed", "error", err)
		if errors.Is(err, ErrInvalidCredentials) {
			return response.JsonError(c, 401, "用户名或密码错误")
		}
		return response.JsonError(c, 500, "服务器错误")
	}

	accessToken, refreshToken, expireIn, err := h.srv.GenerateTokens(ctx, user)
	if err != nil {
		slog.Error("GenerateTokens failed", "error", err)
		return response.JsonError(c, 500, "服务器错误")
	}
	setRefreshCookie(c, refreshToken, h.srv.options.RefreshTokenExpireSeconds, h.cookieSecure)
	return response.JsonData(c, LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expireIn,
	})
}

func (h *Handler) Logout(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	if cookie, err := c.Cookie("refresh_token"); err == nil {
		h.srv.Logout(ctx, cookie.Value)
	}
	setRefreshCookie(c, "", -1, h.cookieSecure)
	return response.JsonData(c, "")
}

func (h *Handler) RefreshToken(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return response.JsonError(c, 401, "请登录")
	}

	accessToken, refreshToken, expireIn, err := h.srv.RefreshTokens(ctx, cookie.Value)
	if err != nil {
		slog.Error("RefreshTokens failed", "error", err)
		if errors.Is(err, ErrInvalidRefreshToken) {
			setRefreshCookie(c, "", -1, h.cookieSecure)
			return response.JsonError(c, 401, "请登录")
		}
		return response.JsonError(c, 500, "服务器错误")
	}

	setRefreshCookie(c, refreshToken, h.srv.options.RefreshTokenExpireSeconds, h.cookieSecure)
	return response.JsonData(c, LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expireIn,
	})
}

func setRefreshCookie(c *echo.Context, value string, maxAge int, secure bool) {
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		MaxAge:   maxAge,
		Path:     "/",
		Domain:   "",
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

type UserInfoResponse struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

func (h *Handler) UserInfo(c *echo.Context) error {
	user, ok := reqctx.GetUser(c)
	if !ok || user == nil || user.ID == 0 {
		return response.JsonError(c, 403, "access denied")
	}

	return response.JsonData(c, UserInfoResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt.UnixMilli(),
		UpdatedAt: user.UpdatedAt.UnixMilli(),
	})
}

type UpdatePasswordRequest struct {
	Password    string `json:"password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

func (h *Handler) UpdatePassword(c *echo.Context) error {

	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	user, ok := reqctx.GetUser(c)
	if !ok || user == nil || user.ID == 0 {
		return response.JsonError(c, 403, "access denied")
	}

	var req UpdatePasswordRequest
	if err := c.Bind(&req); err != nil {
		return response.JsonError(c, 400, err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return response.JsonError(c, 400, err.Error())
	}

	if req.Password == "" || req.NewPassword == "" {
		return response.JsonError(c, 400, "参数错误")
	}

	_, err := h.srv.AuthUser(ctx, user.Username, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return response.JsonError(c, 400, "验证失败")
		}
		return response.JsonError(c, 500, "服务器错误")
	}

	err = h.srv.UpdatePassword(ctx, user.ID, req.NewPassword)
	if err != nil {
		return response.JsonError(c, 500, err.Error())
	}

	h.srv.RevokeAllUserTokens(ctx, user.ID)

	return response.JsonData(c, "")
}
