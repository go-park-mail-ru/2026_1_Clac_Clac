package config

type ClientSection struct {
	ClientConfig `mapstructure:",squash"`
}

type Section struct {
	Client ClientSection `mapstructure:"client"`
}

func DefaultSectionConfig() Section {
	return Section{
		Client: ClientSection{
			ClientConfig: DefaultClientConfig(),
		},
	}
}
