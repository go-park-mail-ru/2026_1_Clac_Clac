package common

import "errors"

var (
	ErrBoardNotFound = errors.New("board not found")
	ErrActionDenied  = errors.New("action denied")
)
