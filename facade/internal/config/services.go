package config

type Services struct {
	MailSender   ClientConfig `mapstructure:"mail_sender"`
	User         User         `mapstructure:"user"`
	Auth         Auth         `mapstructure:"auth"`
	Board        ClientConfig `mapstructure:"board"`
	Section      ClientConfig `mapstructure:"section"`
	Card         ClientConfig `mapstructure:"card"`
	RateLimiters RateLimiters `mapstructure:"rate_limiters"`
}

func DefaultServicesConfig() Services {
	return Services{
		MailSender:   DefaultClientConfig(),
		User:         DefaultUserConfig(),
		Auth:         DefaultAuthConfig(),
		Board:        DefaultClientConfig(),
		RateLimiters: DefaultActionsRateLimiters(),
	}
}
