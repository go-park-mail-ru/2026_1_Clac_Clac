package config

type Services struct {
	MailSender   ClientConfig `mapstructure:"mail_sender"`
	User         User         `mapstructure:"user"`
	Auth         Auth         `mapstructure:"auth"`
	Board        Board        `mapstructure:"board"`
	Section      Section      `mapstructure:"section"`
	Card         ClientConfig `mapstructure:"card"`
	RateLimiters RateLimiters `mapstructure:"rate_limiters"`
}

func DefaultServicesConfig() Services {
	return Services{
		MailSender:   DefaultClientConfig(),
		User:         DefaultUserConfig(),
		Auth:         DefaultAuthConfig(),
		Board:        DefaultBoardConfig(),
		Section:      DefaultSectionConfig(),
		RateLimiters: DefaultActionsRateLimiters(),
	}
}
