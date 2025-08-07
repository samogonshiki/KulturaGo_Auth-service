package custom_err

import "errors"

var (
	ErrExists             = errors.New("user exists")
	ErrInvalidCreds       = errors.New("invalid credentials")
	ErrKeySize            = errors.New("key must be 16, 24 or 32 bytes")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNotAdmin           = errors.New("not an admin")
)
