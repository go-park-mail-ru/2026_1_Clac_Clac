package common

import "errors"

var (
	ErrorExistingUser     = errors.New("user with this email alreday exists")
	ErrorNotNullValue     = errors.New("put null value in not null field")
	ErrorNonexistentUser  = errors.New("user with this ID not exist")
	ErrorNonexistentEmail = errors.New("user with this email not exist")

	ErrorNotExistingResetToken = errors.New("reset token not found or expire")

	ErrorMissingRequiredField = errors.New("required field is missing")
	ErrorInvalidProfileData   = errors.New("incorrect profile data")
)
