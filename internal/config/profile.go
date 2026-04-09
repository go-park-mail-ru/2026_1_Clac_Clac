package config

const (
	profileConfigDefaultSiganatureBytes       = 512
	profileConfigDefaultMaxReadBytes          = 5 << 20
	profileConfigDefaultMaxLenNameUser        = 128
	profileConfigDefaultMaxLenDescriptionUser = 500
)

type ProfileHandler struct {
	SiganatureTypeBytes   int   `mapstructure:"siganature_type_bytes"`
	MaxReadBytes          int64 `mapstructure:"max_read_bytes"`
	MaxLenNameUser        int   `mapstructure:"max_len_name_user"`
	MaxLenDescriptionUser int   `mapstructure:"max_len_description_user"`
}

type Profile struct {
	Handler ProfileHandler `mapstructure:"handler"`
}

func DefaultProfileConfig() Profile {
	return Profile{
		Handler: ProfileHandler{
			SiganatureTypeBytes:   profileConfigDefaultSiganatureBytes,
			MaxReadBytes:          profileConfigDefaultMaxReadBytes,
			MaxLenNameUser:        profileConfigDefaultMaxLenNameUser,
			MaxLenDescriptionUser: profileConfigDefaultMaxLenDescriptionUser,
		},
	}
}
