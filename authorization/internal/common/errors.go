package common

import (
	"errors"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

var (
	ErrorNotExistingSession = errors.New("session not found or expired")
)
