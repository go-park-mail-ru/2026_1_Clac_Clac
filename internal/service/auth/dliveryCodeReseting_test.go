package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	mockAuthRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth/mock_auth_rep"
	mockSender "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth/mock_sender"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDiliveryCodeReseting(t *testing.T) {
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
				m.On("AddResetToken", mock.Anything, mock.AnythingOfType("dbConnection.ResetToken")).Return(nil)
			},
			senderMock: func(m *mockSender.SenderLetters) {
				m.On("SendLetter", targetEmail, "Code for create new pasword", mock.AnythingOfType("string")).Return(nil)
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

			service := NewAuthService(mockRepo, mockMail, nil, nil, test.generator, test.generator)

			err := service.DiliveryCodeReseting(context.Background(), test.email)

			time.Sleep(10 * time.Millisecond)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
