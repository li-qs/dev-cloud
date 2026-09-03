package errmsg

import (
	"errors"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrResourceNotFound   = errors.New("resource not found")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrResourceState      = errors.New("invalid resource state")
)
