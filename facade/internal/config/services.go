package config

type Services struct {
	MailSender   ClientConfig `mapstructure:"mail_sender"`
	User         User         `mapstructure:"user"`
	Auth         Auth         `mapstructure:"auth"`
	Board        ClientConfig `mapstructure:"board"`
	RateLimiters RateLimiters `mapstructure:"rate_limiters"`
	Appeal       Appeal       `mapstructure:"appeal"`
}

func DefaultServicesConfig() Services {
	return Services{
		MailSender:   DefaultClientConfig(),
		User:         DefaultUserConfig(),
		Auth:         DefaultAuthConfig(),
		Board:        DefaultClientConfig(),
		RateLimiters: DefaultActionsRateLimiters(),
		Appeal:       DefaultAppealConfig(),
	}
}
