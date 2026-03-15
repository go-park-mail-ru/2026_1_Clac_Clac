package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/models"
	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service"
	mockAuthRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service/tests/mock_auth_rep"
	mockSender "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service/tests/mock_sender"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendRecoveryCode(t *testing.T) {
	targetEmail := "test@mail.ru"

	tests := []struct {
		nameTest      string
		email         string
		generator     func() (string, error)
		mockBehavior  func(m *mockAuthRep.AuthRepository)
		senderMock    func(m *mockSender.SenderLetters)
		expectedError error
	}{
		{
			nameTest:  "Success delivery code",
			email:     targetEmail,
			generator: func() (string, error) { return "123456", nil },
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("GetUser", mock.Anything, targetEmail).Return(models.User{ID: common.FixedUserUuiD}, nil)
				m.On("AddResetToken", mock.Anything, mock.AnythingOfType("db.ResetToken")).Return(nil)
			},
			senderMock: func(m *mockSender.SenderLetters) {
				m.On("SendLetter", targetEmail, "Code to create a new password", mock.AnythingOfType("string")).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest:  "Error user not found",
			email:     "testing@mail.ru",
			generator: func() (string, error) { return "123456", nil },
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("GetUser", mock.Anything, "testing@mail.ru").Return(models.User{}, common.ErrorNonexistentUser)
			},
			senderMock:    func(m *mockSender.SenderLetters) {},
			expectedError: fmt.Errorf("rep.GetUser: %w", common.ErrorNonexistentUser),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			mockMail := mockSender.NewSenderLetters(t)

			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}
			if test.senderMock != nil {
				test.senderMock(mockMail)
			}

			service := service.NewService(mockRepo, mockMail, nil, nil, test.generator, test.generator)

			err := service.SendRecoveryCode(context.Background(), test.email)

			time.Sleep(10 * time.Millisecond)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
