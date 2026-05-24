package config

type ClientBoard struct {
	ClientConfig `mapstructure:",squash"`
}

type Board struct {
	Client ClientBoard `mapstructure:"client"`
}

func DefaultBoardConfig() Board {
	return Board{
		Client: ClientBoard{
			ClientConfig: DefaultClientConfig(),
		},
	}
}
