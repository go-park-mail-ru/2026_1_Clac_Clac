package domain

type RateLimitCheck struct {
	UserIp   string
	Action   string
	WindowMs int64
	Limit    int64
}

type Cooldown struct {
	Name         string
	Email        string
	ExpirationMs int64
}

type CooldownResult struct {
	Allowed bool
	WaitMs  int64
}
