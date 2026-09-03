package dto

type LoginRequest struct {
	Username string `json:"username" validate:"required,max=64"`
	Password string `json:"password" validate:"required,min=6,max=72"`
}

type UpdatePasswordRequest struct {
	Password    string `json:"password" validate:"required,min=6,max=72"`
	NewPassword string `json:"new_password" validate:"required,min=6,max=72"`
}

type LoginResponse struct {
	AccessToken string           `json:"access_token"`
	TokenType   string           `json:"token_type"`
	ExpiresIn   int              `json:"expires_in"`
	User        UserInfoResponse `json:"user"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type UserInfoResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname,omitempty"`
}
