package common

import "errors"

var (
	ErrorNotExistingResetToken = errors.New("reset token not found or expired")
)
