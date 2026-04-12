package common

import "errors"

var (
	ErrBoardNotFound           = errors.New("board not found")
	ErrActionDenied            = errors.New("action denied")
	ErrorNotNullValue          = errors.New("required field is missing")
	ErrorInvalidBoardData      = errors.New("invalid board data provided")
	ErrorInvalidBoardReference = errors.New("referenced entity does not exist")
	ErrorUserAlreadyMember     = errors.New("this conection board and user is already exist")
)
