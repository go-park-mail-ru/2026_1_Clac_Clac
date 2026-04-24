package dto

import "time"

type RateLimiterConfig struct {
	UserIP string
	Action string
	Window time.Duration
}

type CooldownConfig struct {
	Key        string
	Expiration time.Duration
}
