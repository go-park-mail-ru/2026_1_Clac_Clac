package common

import "errors"

var (
	ErrTimeout           = errors.New("timeout")
	ErrBoardLinkMissing  = errors.New("board link missing")
	ErrInvalidBoardLink  = errors.New("board link invalid")
	ErrParseLink         = errors.New("auth parse link")
	ErrUserNotAuthorized = errors.New("user not authorized")
	ErrCannotGetEvents   = errors.New("cannot get events")
)
