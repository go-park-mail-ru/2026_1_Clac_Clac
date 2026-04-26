package clients

import "errors"

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailNotFound      = errors.New("user with this email not found")
	ErrWrongCredentials   = errors.New("wrong email or password")
	ErrNullInNotNullField = errors.New("null value in not null field")
	ErrInvalidInput       = errors.New("invalid input parameters")
	ErrVKOAuthUnavailable = errors.New("vk oauth service unavailable")

	ErrSessionNotFound = errors.New("session not found or expired")

	ErrResetTokenNotFound = errors.New("reset token does not exist")
)
