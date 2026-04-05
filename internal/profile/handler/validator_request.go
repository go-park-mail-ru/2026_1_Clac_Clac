package handler

import (
	"errors"
)

var (
	ErrorIncorrectSymbol = errors.New("allowed only a-z, A-Z, 0-9, and /?!@")
	ErrorIncorrectLength = errors.New("name must contain maximum 128 symbols")
)

func ValidateInfo(info string, maxLen int) error {
	if len(info) > maxLen {
		return ErrorIncorrectLength
	}
	return nil
}
