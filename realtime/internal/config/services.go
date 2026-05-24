package config

type Services struct {
	Auth  Auth  `mapstructure:"auth"`
	Board Board `mapstructure:"board"`
}

func DefaultServicesConfig() Services {
	return Services{
		Auth:  DefaultAuthConfig(),
		Board: DefaultBoardConfig(),
	}
}
