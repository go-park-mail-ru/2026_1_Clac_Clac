package common

import (
	"errors"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

var (
	ErrorExistingUser     = errors.New("user with this email alreday exists")
	ErrorNotNullValue     = errors.New("put null value in not null field")
	ErrorNonexistentUser  = errors.New("user with this ID not exist")
	ErrorNonexistentEmail = errors.New("user with this email not exist")

	ErrorDetectingSessionCollision = errors.New("session collision detected")

	ErrorNotExistingSession = errors.New("session not found or expired")
	ErrorSeesionExpired     = errors.New("time life session expired")

	ErrorDecodeRequest = errors.New("decoding request is incorrect")

	ErrorNotExistingResetToken   = errors.New("reset token not found or expire")
	ErrorResetTokenExpired       = errors.New("time life reset token expired")
	ErrorDetectingTokenCollision = errors.New("reset token collision detected")

	ErrorExistingBoard = errors.New("board with this ID alreday exists")
)
