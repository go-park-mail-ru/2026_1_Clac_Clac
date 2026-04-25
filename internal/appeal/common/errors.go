package common

import "errors"

var (
	ErrorExistingUser = errors.New("user with this email alreday exists")
	ErrorNotNullValue = errors.New("put null value in not null field")
)
