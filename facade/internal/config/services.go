package config

type Services struct {
	MailSender   ClientConfig `mapstructure:"mail_sender"`
	User         User         `mapstructure:"user"`
	Auth         Auth         `mapstructure:"auth"`
	Card         Card         `mapstructure:"card"`
	Board        Board        `mapstructure:"board"`
	Section      Section      `mapstructure:"section"`
	RateLimiters RateLimiters `mapstructure:"rate_limiters"`
}

func DefaultServicesConfig() Services {
	return Services{
		MailSender:   DefaultClientConfig(),
		User:         DefaultUserConfig(),
		Auth:         DefaultAuthConfig(),
		Card:         DefaultCardConfig(),
		Board:        DefaultBoardConfig(),
		Section:      DefaultSectionConfig(),
		RateLimiters: DefaultActionsRateLimiters(),
	}
}
