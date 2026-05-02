package common

import "errors"

var (
	ErrBoardNotFound         = errors.New("board not found")
	ErrNotNullValue          = errors.New("required field is missing")
	ErrInvalidBoardData      = errors.New("invalid board data provided")
	ErrInvalidBoardReference = errors.New("referenced entity does not exist")
	ErrUserAlreadyMember     = errors.New("this connection board and user already exists")
)
