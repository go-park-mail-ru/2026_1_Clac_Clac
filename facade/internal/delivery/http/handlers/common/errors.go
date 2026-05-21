package common

import "errors"

var (
	ErrInvalidRequestSchema   = errors.New("invalid schema")
	ErrInvalidEmailOrPassword = errors.New("invalid email or password")
	ErrWrongEmailOrPassword   = errors.New("wrong email or password")
	ErrCannotSendRecoveryCode = errors.New("cannot send recovery code")
	ErrCannotResetPassword    = errors.New("cannot reset password")
	ErrInternalServerError    = errors.New("something went wrong")
	ErrUserNotAuthorized      = errors.New("user not authorized")
	ErrUserDoesNotExists      = errors.New("user does not exists")
	ErrUserAlreadyExists      = errors.New("user already exist")
	ErrNullInNotNullField     = errors.New("field can not be null")

	ErrOAuthCodeEmpty              = errors.New("oauth_code_empty")
	ErrOAuthExchangeFailed         = errors.New("oauth_error")
	ErrOAuthNoEmailProvided        = errors.New("oauth_no_email")
	ErrOAuthInvalidEmail           = errors.New("oauth_invalid_email")
	ErrOAuthCannotRequestUserData  = errors.New("oauth_cannot_request_user_data")
	ErrOAuthEmptyUserData          = errors.New("oauth_no_user_data")
	ErrOAuthInternalServerError    = errors.New("oauth_something_went_wrong")
	ErrOAuthCannotSaveRefreshToken = errors.New("oauth_cannot_save_refresh_token")
	ErrOAuthUnavailable            = errors.New("oauth_service_unavailable")

	ErrResetTokenNotExistOrExpired = errors.New("reset token not found or expired")

	ErrCannotCreateCSRFToken        = errors.New("cannot create csrf token")
	ErrCannotGetCSRFTokenExpireTime = errors.New("cannot get csrf token expire time")

	ErrFindCard    = errors.New("can not find card")
	ErrGetCard     = errors.New("can not get info card")
	ErrDeleteCard  = errors.New("can not delete card")
	ErrUpdateCard  = errors.New("can not update card")
	ErrReorderCard = errors.New("can not reorder card")
	ErrCreateCard  = errors.New("can not create new card")
	ErrFindSection = errors.New("can not find section")
	ErrNullValue   = errors.New("can not use null element")
	ErrSetTimeLine = errors.New("incorrect time line for card")

	ErrIncorrectMoveCard   = errors.New("can not skip mandatory section")
	ErrIncorrectUniqCard   = errors.New("link card must be unique")
	ErrIncorrectReferences = errors.New("incorrect foreign key")
	ErrInvalidCardData     = errors.New("invalid card data")
)
