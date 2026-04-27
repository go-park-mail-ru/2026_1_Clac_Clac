package dto

import "time"

type RateLimiterConfig struct {
	UserIP string
	Action string
	Window time.Duration
	Limit  int64
}

type CooldownConfig struct {
	Name       string
	Email      string
	Expiration time.Duration
}
