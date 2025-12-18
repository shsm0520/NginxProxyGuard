package model

import "errors"

// Common errors
var (
	ErrNotFound           = errors.New("not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrDuplicateEntry     = errors.New("duplicate entry")
	ErrInvalidInput       = errors.New("invalid input")
	ErrCertificateExpired = errors.New("certificate expired")
	ErrDNSChallengeFailed = errors.New("DNS challenge failed")
	ErrACMEError          = errors.New("ACME error")
)
