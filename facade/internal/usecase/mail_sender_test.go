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
	errTest    = errors.New("test client error")
)

func TestSendRecoveryCode(t *testing.T) {
	recoveryInfo := domain.RecoveryCode{UserLink: testUserLink, Email: "test@mail.ru"}

	tests := []struct {
		name         string
		mockBehavior func(m *mockMailClient.MailSenderClient)
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockMailClient.MailSenderClient) {
				m.On("SendRecoveryCode", context.Background(), recoveryInfo).Return(nil)
			},
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockMailClient.MailSenderClient) {
				m.On("SendRecoveryCode", context.Background(), recoveryInfo).Return(errTest)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockMailClient.NewMailSenderClient(t)
			tc.mockBehavior(m)

			err := NewMailSender(m).SendRecoveryCode(context.Background(), testUserLink, "test@mail.ru")

			if tc.expectError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, errTest))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckRecoveryCode(t *testing.T) {
	checkInfo := domain.RecoveryCodeCheck{Code: "123456"}

	tests := []struct {
		name         string
		mockBehavior func(m *mockMailClient.MailSenderClient)
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockMailClient.MailSenderClient) {
				m.On("CheckRecoveryCode", context.Background(), checkInfo).Return(nil)
			},
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockMailClient.MailSenderClient) {
				m.On("CheckRecoveryCode", context.Background(), checkInfo).Return(errTest)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockMailClient.NewMailSenderClient(t)
			tc.mockBehavior(m)

			err := NewMailSender(m).CheckRecoveryCode(context.Background(), "123456")

			if tc.expectError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, errTest))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExchangeTokenForUser(t *testing.T) {
	resetToken := domain.ResetToken{Token: "valid_token"}

	tests := []struct {
		name         string
		resetToken   domain.ResetToken
		mockBehavior func(m *mockMailClient.MailSenderClient)
		expectedLink uuid.UUID
		expectError  bool
	}{
		{
			name:       "Success",
			resetToken: resetToken,
			mockBehavior: func(m *mockMailClient.MailSenderClient) {
				m.On("ExchangeTokenForUser", context.Background(), resetToken).Return(testUserLink, nil)
			},
			expectedLink: testUserLink,
			expectError:  false,
		},
		{
			name:       "ClientError",
			resetToken: domain.ResetToken{Token: "invalid_token"},
			mockBehavior: func(m *mockMailClient.MailSenderClient) {
				m.On("ExchangeTokenForUser", context.Background(), domain.ResetToken{Token: "invalid_token"}).
					Return(uuid.Nil, errTest)
			},
			expectedLink: uuid.Nil,
			expectError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockMailClient.NewMailSenderClient(t)
			tc.mockBehavior(m)

			resultLink, err := NewMailSender(m).ExchangeTokenForUser(context.Background(), tc.resetToken)

			if tc.expectError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, errTest))
				assert.Equal(t, uuid.Nil, resultLink)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedLink, resultLink)
			}
		})
	}
}
