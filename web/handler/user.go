package handler

import (
	"devcloud/web/handler/dto"
	"devcloud/web/handler/reqctx"
	"devcloud/web/handler/response"
	"devcloud/web/service"
	"net/http"

	"github.com/labstack/echo/v5"
)

type User struct {
	cookieSecure bool
	userSrv      *service.User
}

func NewUser(
	cookieSecure bool,
	userSrv *service.User,
) *User {
	return &User{
		cookieSecure: cookieSecure,
		userSrv:      userSrv,
	}
}

func (h *User) Login(c *echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}
	if err := c.Validate(&req); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	user, valid, err := h.userSrv.AuthUser(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		return err
	}
	if !valid {
		return echo.ErrUnauthorized
	}

	accessToken, refreshToken, expireIn, err := h.userSrv.GenerateTokens(c.Request().Context(), user)
	if err != nil {
		return err
	}
	setRefreshCookie(c, refreshToken, h.userSrv.Options.RefreshTokenExpireSeconds, h.cookieSecure)
	return response.JsonData(c, dto.LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expireIn,
	})
}

func (h *User) Logout(c *echo.Context) error {
	if cookie, err := c.Cookie("refresh_token"); err == nil {
		if err := h.userSrv.Logout(c.Request().Context(), cookie.Value); err != nil {
			return err
		}
	}
	setRefreshCookie(c, "", -1, h.cookieSecure)
	return response.JsonSuccess(c)
}

func (h *User) RefreshToken(c *echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return echo.ErrUnauthorized
	}

	t, valid, err := h.userSrv.RefreshTokens(c.Request().Context(), cookie.Value)
	if err != nil {
		return err
	}
	if !valid {
		setRefreshCookie(c, "", -1, h.cookieSecure)
		return echo.ErrUnauthorized
	}

	setRefreshCookie(c, t.RefreshToken, h.userSrv.Options.RefreshTokenExpireSeconds, h.cookieSecure)
	return response.JsonData(c, dto.LoginResponse{
		AccessToken: t.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   t.ExpiresIn,
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

func (h *User) UpdatePassword(c *echo.Context) error {
	user, err := reqctx.GetUser(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	var req dto.UpdatePasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	if req.Password == "" || req.NewPassword == "" {
		return echo.ErrBadRequest
	}

	_, valid, err := h.userSrv.AuthUser(c.Request().Context(), user.Username, req.Password)
	if err != nil {
		return err
	}
	if !valid {
		return echo.ErrUnauthorized
	}

	err = h.userSrv.UpdatePassword(c.Request().Context(), user.ID, req.NewPassword)
	if err != nil {
		return err
	}

	h.userSrv.RevokeAllUserTokens(c.Request().Context(), user.ID)

	return response.JsonSuccess(c)
}

func (h *User) UserInfo(c *echo.Context) error {
	user, err := reqctx.GetUser(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	return response.JsonData(c, dto.UserInfoResponse{
		ID:       user.ID,
		Username: user.Username,
	})
}
