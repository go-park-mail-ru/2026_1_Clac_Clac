package config

type Services struct {
	MailSender    ClientConfig `mapstructure:"mail_sender"`
	User          ClientConfig `mapstructure:"user"`
	Authorization ClientConfig `mapstructure:"authorization"`
	Board         ClientConfig `mapstructure:"board"`
	RateLimiters  RateLimiters `mapstructure:"rate_limiters"`
}

func DefaultServicesConfig() Services {
	return Services{
		MailSender:    DefaultClientConfig(),
		User:          DefaultClientConfig(),
		Authorization: DefaultClientConfig(),
		Board:         DefaultClientConfig(),
		RateLimiters:  DefaultActionsRateLimiters(),
	}
}
