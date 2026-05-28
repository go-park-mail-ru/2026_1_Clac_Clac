package config

const (
	defaultMaxTask = 200
)

type ClientSection struct {
	ClientConfig `mapstructure:",squash"`
}

type HandlerSection struct {
	MaxLenDisplayName int   `mapstructure:"max_len_display_name"`
	MaxQuantityTasks  int64 `mapstructure:"max_quantity_tasks"`
}

type Section struct {
	Client  ClientSection  `mapstructure:"client"`
	Handler HandlerSection `mapstructure:"handler"`
}

func DefaultSectionConfig() Section {
	return Section{
		Client: ClientSection{
			ClientConfig: DefaultClientConfig(),
		},
		Handler: HandlerSection{
			MaxLenDisplayName: defaultMaxDisplayName,
			MaxQuantityTasks:  defaultMaxTask,
		},
	}
}
