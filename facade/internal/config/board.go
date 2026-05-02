package config
<<<<<<< HEAD
=======

type BoardHandler struct {
	MultipartBackgroundFileKey string `json:"multipart_background_file_key"`
}

type ClientBoard struct {
	ClientConfig `mapstructure:",squash"`
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
		},
	}
}
>>>>>>> feat/add-sections-to-facade
