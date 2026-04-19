package common

import "errors"

var (
	ErrBoardNotFound         = errors.New("board not found")
	ErrActionDenied          = errors.New("action denied")
	ErrNotNullValue          = errors.New("required field is missing")
	ErrInvalidBoardData      = errors.New("invalid board data provided")
	ErrInvalidBoardReference = errors.New("referenced entity does not exist")
	ErrUserAlreadyMember     = errors.New("this conection board and user is already exist")
)
