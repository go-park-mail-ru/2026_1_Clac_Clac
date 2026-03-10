package config_test

import (
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMailSenderConfig(t *testing.T) {
	t.Run("test reading from env", func(t *testing.T) {
		want := config.MailSender{
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

		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()

		config.SetupEnvMailSender(v)

		// Надо, чтобы viper корректно обработал префикс mail_sender
		var conf struct {
			MailSender config.MailSender `mapstructure:"mail_sender"`
		}

		err := v.Unmarshal(&conf)
		require.NoError(t, err, "viper must not return error")
		assert.Equal(t, want, conf.MailSender)
	})
}
