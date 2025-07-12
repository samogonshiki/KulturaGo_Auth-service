package custom_err

import "errors"

var (
	ErrExists       = errors.New("user exists")
	ErrInvalidCreds = errors.New("invalid credentials")
)
