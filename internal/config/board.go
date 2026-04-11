package config

const (
	defaultMultipartBackgroundFileKey = "background"
	defaultMaxBackgroundSize          = 10 * 1024 * 1024 // 10 МБайт
	defaultCreateBoardDefaultUserRole = "creator"
)

type BoardHandler struct {
	MultipartBackgroundFileKey string `mapstructure:"multipart_background_file_key"`
	MaxBackgroundSize          int64  `mapstructure:"max_background_size"`
}

type BoardRepository struct {
	CreateBoardDefaultUserRole string `mapstructure:"create_board_default_user_role"`
}

type Board struct {
	Handler    BoardHandler    `mapstructure:"handler"`
	Repository BoardRepository `mapstructure:"repository"`
}

func DefaultBoardConfig() Board {
	return Board{
		Handler: BoardHandler{
			MultipartBackgroundFileKey: defaultMultipartBackgroundFileKey,
			MaxBackgroundSize:          defaultMaxBackgroundSize,
		},
		Repository: BoardRepository{
			CreateBoardDefaultUserRole: defaultCreateBoardDefaultUserRole,
		},
	}
}
