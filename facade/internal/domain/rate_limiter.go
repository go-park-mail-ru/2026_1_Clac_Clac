package domain

import "time"

type RateLimitCheck struct {
	UserIP  string
	Action  string
	WindowS int64
	Limit   int64
}

type Cooldown struct {
	Name        string
	Email       string
	ExpirationS int64
}

type CooldownResult struct {
	Allowed bool
	WaitS   int64
}

type RateLimitConfig struct {
	Limit   int64
	Action  string
	WindowS int64
	TTL     time.Duration
}
