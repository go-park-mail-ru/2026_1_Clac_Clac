package repository

import "errors"

var (
	ErrorExistingUser       = errors.New("user with this ID alreday exists")
	ErrorNonexistentUser    = errors.New("user with this ID not exist")
	ErrorDetectingCollision = errors.New("session collision detected")

	ErrorNotExistingSession = errors.New("session not found or expired")
	ErrorSeesionExpired     = errors.New("time life session expired")
)
