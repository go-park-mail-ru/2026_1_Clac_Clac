package common

import "errors"

var (
	ErrBoardNotFound         = errors.New("board not found")
	ErrNotNullValue          = errors.New("required field is missing")
	ErrInvalidBoardData      = errors.New("invalid board data provided")
	ErrInvalidBoardReference = errors.New("referenced entity does not exist")
	ErrUserAlreadyMember     = errors.New("this connection board and user already exists")

	ErrInviteNotFound    = errors.New("invite not found")
	ErrInviteExpired     = errors.New("invite is expired")
	ErrInviteClosed      = errors.New("invite is closed")
	ErrInviteUserMismatch = errors.New("this invite is for a specific user")
	ErrInviteNotForUser  = errors.New("this invite targets another user")
)
