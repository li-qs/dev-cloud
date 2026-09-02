package user

import "fmt"

var (
	ErrInvalidCredentials  = fmt.Errorf("invalid credentials")
	ErrInvalidRefreshToken = fmt.Errorf("invalid refresh token")
)
