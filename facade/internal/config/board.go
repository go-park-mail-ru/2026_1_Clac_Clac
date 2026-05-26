package config

var (
	defaultChunkSize = 1024 * 1024
)

type BoardHandler struct {
	MultipartBackgroundFileKey string `mapstructure:"multipart_background_file_key"`
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
		},
		Client: ClientBoard{
			ClientConfig: DefaultClientConfig(),
			ChunkSize:    defaultChunkSize,
		},
	}
}
