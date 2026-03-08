package repository

import "errors"

var (
	ErrorExistingUser       = errors.New("user with this email alreday exists")
	ErrorNonexistentUser    = errors.New("user with this email not exist")
	ErrorDetectingCollision = errors.New("session collision detected")

	ErrorNotExistingSession = errors.New("session not found or expired")
	ErrorSeesionExpired     = errors.New("time life session expired")
)
