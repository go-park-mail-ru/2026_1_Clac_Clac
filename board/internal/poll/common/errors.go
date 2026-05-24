package common

import "errors"

var (
	ErrPollAlreadyExists = errors.New("poll already exists for this board")
	ErrPollNotFound      = errors.New("no active poll for this board")
	ErrNotPollAdmin      = errors.New("only poll admin can perform this action")
	ErrUserNotInvited    = errors.New("user is not invited to this poll")
	ErrPollNoMoreCards   = errors.New("no more cards in poll")
)
