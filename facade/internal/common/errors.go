package common

import "errors"

var (
	ErrorParseLink = errors.New("fail parse user link to uuid")

	ErrorExistingUser         = errors.New("user already exists")
	ErrorNotNullValue         = errors.New("null value in not null field")
	ErrorNonexistentUser      = errors.New("user not found")
	ErrorNonexistentEmail     = errors.New("user with this email not found")
	ErrorWrongCredentials     = errors.New("wrong email or password")
	ErrorInvalidInput         = errors.New("invalid input parameters")
	ErrorSessionNotFound      = errors.New("session not found or expired")
	ErrorResetTokenNotFound   = errors.New("reset token not found or expired")
	ErrorVKOAuthUnavailable   = errors.New("vk oauth service unavailable")
	ErrorMissingRequiredField = errors.New("required field is missing")
	ErrorInvalidProfileData   = errors.New("incorrect profile data")

	ErrInvalidCSRFToken               = errors.New("invalid csrf token")
	ErrCannotParseExpireTimeCSRFToken = errors.New("cannot parse expire time csrf token")
	ErrCSRFTokenExpired               = errors.New("csrf token expired")
	ErrCannotDecodeReceivedCSRFToken  = errors.New("cannot decode received csrf token")
	ErrCSRFTokensDoNotEqual           = errors.New("csrf tokens do not equal")
)
