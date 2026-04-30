package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	mockMailClient "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/usecase/mock_mail_sender_client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testUserLink = uuid.New()
	testError    = errors.New("test client error")
)

func TestSendRecoveryCode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := mockMailClient.NewMailSenderClient(t)
		recoveryInfo := domain.RecoveryCode{
			UserLink: testUserLink,
			Email:    "test@mail.ru",
		}
		m.On("SendRecoveryCode", context.Background(), recoveryInfo).Return(nil)

		err := NewMailSender(m).SendRecoveryCode(context.Background(), testUserLink, "test@mail.ru")
		require.NoError(t, err)
	})

	t.Run("ClientError", func(t *testing.T) {
		m := mockMailClient.NewMailSenderClient(t)
		recoveryInfo := domain.RecoveryCode{
			UserLink: testUserLink,
			Email:    "test@mail.ru",
		}
		m.On("SendRecoveryCode", context.Background(), recoveryInfo).Return(testError)

		err := NewMailSender(m).SendRecoveryCode(context.Background(), testUserLink, "test@mail.ru")
		require.Error(t, err)
		assert.True(t, errors.Is(err, testError))
	})
}

func TestCheckRecoveryCode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := mockMailClient.NewMailSenderClient(t)
		checkInfo := domain.RecoveryCodeCheck{
			Code: "123456",
		}
		m.On("CheckRecoveryCode", context.Background(), checkInfo).Return(nil)

		err := NewMailSender(m).CheckRecoveryCode(context.Background(), "123456")
		require.NoError(t, err)
	})

	t.Run("ClientError", func(t *testing.T) {
		m := mockMailClient.NewMailSenderClient(t)
		checkInfo := domain.RecoveryCodeCheck{
			Code: "123456",
		}
		m.On("CheckRecoveryCode", context.Background(), checkInfo).Return(testError)

		err := NewMailSender(m).CheckRecoveryCode(context.Background(), "123456")
		require.Error(t, err)
		assert.True(t, errors.Is(err, testError))
	})
}

func TestExchangeTokenForUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := mockMailClient.NewMailSenderClient(t)
		resetToken := domain.ResetToken{
			Token: "valid_token",
		}
		m.On("ExchangeTokenForUser", context.Background(), resetToken).Return(testUserLink, nil)

		resultLink, err := NewMailSender(m).ExchangeTokenForUser(context.Background(), resetToken)
		require.NoError(t, err)
		assert.Equal(t, testUserLink, resultLink)
	})

	t.Run("ClientError", func(t *testing.T) {
		m := mockMailClient.NewMailSenderClient(t)
		resetToken := domain.ResetToken{
			Token: "invalid_token",
		}
		m.On("ExchangeTokenForUser", context.Background(), resetToken).Return(uuid.Nil, testError)

		resultLink, err := NewMailSender(m).ExchangeTokenForUser(context.Background(), resetToken)
		require.Error(t, err)
		assert.True(t, errors.Is(err, testError))
		assert.Equal(t, uuid.Nil, resultLink)
	})
}
