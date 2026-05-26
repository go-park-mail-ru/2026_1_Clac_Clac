package config

var (
	defaultChunkSize      = 1024 * 1024
	defaultMaxDisplayName = 128
)

type BoardHandler struct {
	MultipartBackgroundFileKey string `mapstructure:"multipart_background_file_key"`
	MaxDisplayName             int    `mapstructure:"max_display_name"`
}

type ClientBoard struct {
	ClientConfig `mapstructure:",squash"`
	ChunkSize    int `mapstructure:"chunk_size"`
}

type Board struct {
	Handler BoardHandler `mapstructure:"handler"`
	Client  ClientBoard  `mapstructure:"client"`
}

func DefaultBoardConfig() Board {
	return Board{
		Handler: BoardHandler{
			MultipartBackgroundFileKey: "",
			MaxDisplayName:             defaultMaxDisplayName,
		},
		Client: ClientBoard{
			ClientConfig: DefaultClientConfig(),
			ChunkSize:    defaultChunkSize,
		},
	}
}
