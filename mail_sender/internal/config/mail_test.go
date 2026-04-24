package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMailSenderConfig(t *testing.T) {
	t.Run("test reading from env", func(t *testing.T) {
		want := Mail{
			Host:     "smtp.mail.ru",
			Port:     "465",
			Username: "test_user",
			Password: "supersecretpassword",
		}

		t.Setenv("MAIL_SENDER_HOST", want.Host)
		t.Setenv("MAIL_SENDER_PORT", want.Port)
		t.Setenv("MAIL_SENDER_USERNAME", want.Username)
		t.Setenv("MAIL_SENDER_PASSWORD", want.Password)

		v := viper.New()

		SetupEnvMailSender(v)

		var conf struct {
			Mail Mail `mapstructure:"mail"`
		}

		err := v.Unmarshal(&conf)
		require.NoError(t, err, "viper must not return error")
		assert.Equal(t, want, conf.Mail)
	})
}
