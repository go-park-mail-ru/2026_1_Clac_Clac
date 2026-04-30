package config

const (
	profileConfigDefaultSignatureBytes        = 512
	profileConfigDefaultMaxReadBytes          = 5 << 20
	profileConfigDefaultMaxLenNameUser        = 128
	profileConfigDefaultMaxLenDescriptionUser = 500
	profileConfigDefaultMaxLenPassword        = 128
	profileConfigDefaultMinLenPassword        = 8
)

type UserHandler struct {
	SignatureTypeBytes     int                 `mapstructure:"signature_type_bytes"`
	MaxReadBytes          int64               `mapstructure:"max_read_bytes"`
	MaxLenNameUser        int                 `mapstructure:"max_len_name_user"`
	MaxLenDescriptionUser int                 `mapstructure:"max_len_description_user"`
	MaxLenPassword        int                 `mapstructure:"max_len_password"`
	MinLenPassword        int                 `mapstructure:"min_len_password"`
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
			SignatureTypeBytes:    profileConfigDefaultSignatureBytes,
			MaxReadBytes:          profileConfigDefaultMaxReadBytes,
			MaxLenNameUser:        profileConfigDefaultMaxLenNameUser,
			MaxLenDescriptionUser: profileConfigDefaultMaxLenDescriptionUser,
			MaxLenPassword:        profileConfigDefaultMaxLenPassword,
			MinLenPassword:        profileConfigDefaultMinLenPassword,
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
