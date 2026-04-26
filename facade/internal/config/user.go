package config

const (
	profileConfigDefaultSiganatureBytes       = 512
	profileConfigDefaultMaxReadBytes          = 5 << 20
	profileConfigDefaultMaxLenNameUser        = 128
	profileConfigDefaultMaxLenDescriptionUser = 500
)

type UserHandler struct {
	SiganatureTypeBytes   int                 `mapstructure:"signature_type_bytes"`
	MaxReadBytes          int64               `mapstructure:"max_read_bytes"`
	MaxLenNameUser        int                 `mapstructure:"max_len_name_user"`
	MaxLenDescriptionUser int                 `mapstructure:"max_len_description_user"`
	ValidExtensions       map[string]struct{} `mapstructure:"valid_extensions"`
}

type UserClient struct {
	ClientConfig `mapstructure:",squash"`
}

type User struct {
	Handler UserHandler `mapstructure:"handler"`
	Client  UserClient  `mapstructure:"client"`
}

func DefaultUserConfig() User {
	return User{
		Handler: UserHandler{
			SiganatureTypeBytes:   profileConfigDefaultSiganatureBytes,
			MaxReadBytes:          profileConfigDefaultMaxReadBytes,
			MaxLenNameUser:        profileConfigDefaultMaxLenNameUser,
			MaxLenDescriptionUser: profileConfigDefaultMaxLenDescriptionUser,
			ValidExtensions:       DefaultValidExtensions(),
		},
		Client: UserClient{
			ClientConfig: DefaultClientConfig(),
		},
	}
}

func DefaultValidExtensions() map[string]struct{} {
	return map[string]struct{}{
		"image/png":  {},
		"image/jpeg": {},
		"image/jpg":  {},
		"image/webp": {},
	}
}
