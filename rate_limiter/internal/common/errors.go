package common

import "errors"

var (
	ErrorRateLimitExceeded = errors.New("rate limit exceeded")
	ErrorCooldownActive    = errors.New("cooldown is still active")
)
