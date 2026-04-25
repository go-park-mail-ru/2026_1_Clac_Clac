package common

import "errors"

var (
	ErrorExistingUser     = errors.New("user with this email alreday exists")
	ErrorNotNullValue     = errors.New("put null value in not null field")
	ErrorPermissionDenied = errors.New("role is incorrect for this action")

	IncorrectRequest = "get incorrect format of request"
	IncorrectPath    = "get incorrect format of path"
)
