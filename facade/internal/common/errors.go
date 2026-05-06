package common

import "errors"

var (
	ErrorParseLink = errors.New("fail parse link to uuid")

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
	ErrorInvalidContentType   = errors.New("incorrect content type file")

	ErrorCardNotFound             = errors.New("card not found")
	ErrorSectionNotFound          = errors.New("section not found")
	ErrorCommentNotFound          = errors.New("comment not found")
	ErrorSubtaskNotFound          = errors.New("subtask not found")
	ErrorPermissionDenied         = errors.New("permission denied")
	ErrorCardAlreadyExists        = errors.New("card already exists")
	ErrorTaskLimitReached         = errors.New("task limit reached")
	ErrCannotSkipMandatorySection = errors.New("cannot skip mandatory section")

	ErrorBoardNotFound           = errors.New("board not found")
	ErrorBoardPermissionDenied   = errors.New("board permission denied")
	ErrorSectionPermissionDenied = errors.New("section permission denied")

	ErrorAppealNotFound   = errors.New("appeal not found")
	ErrInvalidCategory    = errors.New("invalid category")
	ErrCannotGetStats     = errors.New("cannot get stats")
	ErrCannotChangeStatus = errors.New("cannot change status")

	ErrInvalidCSRFToken               = errors.New("invalid csrf token")
	ErrCannotParseExpireTimeCSRFToken = errors.New("cannot parse expire time csrf token")
	ErrCSRFTokenExpired               = errors.New("csrf token expired")
	ErrCannotDecodeReceivedCSRFToken  = errors.New("cannot decode received csrf token")
	ErrCSRFTokensDoNotEqual           = errors.New("csrf tokens do not equal")
)
