package redis

import (
	"time"
)

type Config struct {
	PingSleepTime time.Duration
	MaxRetries    int
}
