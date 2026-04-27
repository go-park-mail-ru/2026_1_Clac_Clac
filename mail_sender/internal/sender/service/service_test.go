package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/common"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/sender/repository/dto"
	mockSenderLetters "github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/sender/service/mock_sender_letters"
	mockSenderRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/sender/service/mock_sender_rep"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testLifeTime  = 15 * time.Minute
	testSleepTime = 0
	testRetries   = 1
)

func spyCodeGenerator() (string, error) {
	return "123456", nil
}

func spyCodeGeneratorError() (string, error) {
	return "", errors.New("generator failed")
}

func TestSendRecoveryCode(t *testing.T) {
	userUUID := uuid.New()
	email := "test@mail.ru"

	tests := []struct {
		nameTest       string
		userLink       uuid.UUID
		email          string
		generator      func() (string, error)
		mockRep        func(m *mockSenderRep.SenderRepository)
		mockLetters    func(m *mockSenderLetters.SenderLetters)
		expectedErr    string
	}{
		{
			nameTest:  "Success send recovery code",
			userLink:  userUUID,
			email:     email,
			generator: spyCodeGenerator,
			mockRep: func(m *mockSenderRep.SenderRepository) {
				m.On("AddResetToken", mock.Anything, repositoryDto.ResetTokenEntity{
					ResetTokenKey: "reset_token:123456",
					UserLink:      userUUID,
					LifeTime:      testLifeTime,
				}).Return(nil)
			},
			mockLetters: func(m *mockSenderLetters.SenderLetters) {
				m.On("SendLetter", mock.Anything, email, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			nameTest:  "Error generator fails",
			userLink:  userUUID,
			email:     email,
			generator: spyCodeGeneratorError,
			expectedErr: fmt.Errorf("generatorResetCode: %w", errors.New("generator failed")).Error(),
		},
		{
			nameTest:  "Error AddResetToken fails",
			userLink:  userUUID,
			email:     email,
			generator: spyCodeGenerator,
			mockRep: func(m *mockSenderRep.SenderRepository) {
				m.On("AddResetToken", mock.Anything, mock.Anything).Return(errors.New("redis error"))
			},
			expectedErr: fmt.Errorf("rep.AddResetToken: %w", errors.New("redis error")).Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			rep := mockSenderRep.NewSenderRepository(t)
			letters := mockSenderLetters.NewSenderLetters(t)

			if test.mockRep != nil {
				test.mockRep(rep)
			}
			if test.mockLetters != nil {
				test.mockLetters(letters)
			}

			svc := NewService(rep, letters, Config{
				LifeTimeResetToken: testLifeTime,
				SleepTime:          testSleepTime,
				CountRetries:       testRetries,
			}, Tools{
				GeneratorResetCode: test.generator,
				CreatorResetKey:    CreatorResetKey,
			})

			err := svc.SendRecoveryCode(context.Background(), test.userLink, test.email)

			time.Sleep(5 * time.Millisecond)

			if test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckRecoveryCode(t *testing.T) {
	tests := []struct {
		nameTest    string
		tokenID     string
		mockRep     func(m *mockSenderRep.SenderRepository)
		expectedErr string
	}{
		{
			nameTest: "Success check recovery code",
			tokenID:  "123456",
			mockRep: func(m *mockSenderRep.SenderRepository) {
				m.On("GetUserLinkByResetToken", mock.Anything, "reset_token:123456").
					Return(uuid.New().String(), nil)
			},
		},
		{
			nameTest: "Token not found",
			tokenID:  "123456",
			mockRep: func(m *mockSenderRep.SenderRepository) {
				m.On("GetUserLinkByResetToken", mock.Anything, "reset_token:123456").
					Return("", common.ErrorNotExistingResetToken)
			},
			expectedErr: fmt.Errorf("rep.GetUserLinkByResetToken: %w", common.ErrorNotExistingResetToken).Error(),
		},
		{
			nameTest: "Repository error",
			tokenID:  "123456",
			mockRep: func(m *mockSenderRep.SenderRepository) {
				m.On("GetUserLinkByResetToken", mock.Anything, "reset_token:123456").
					Return("", errors.New("redis error"))
			},
			expectedErr: fmt.Errorf("rep.GetUserLinkByResetToken: %w", errors.New("redis error")).Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			rep := mockSenderRep.NewSenderRepository(t)
			if test.mockRep != nil {
				test.mockRep(rep)
			}

			svc := NewService(rep, nil, Config{}, Tools{
				CreatorResetKey: CreatorResetKey,
			})

			err := svc.CheckRecoveryCode(context.Background(), test.tokenID)

			if test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetUserLink(t *testing.T) {
	expectedLink := uuid.New().String()

	tests := []struct {
		nameTest     string
		tokenID      string
		mockRep      func(m *mockSenderRep.SenderRepository)
		expectedLink string
		expectedErr  string
	}{
		{
			nameTest: "Success get user link",
			tokenID:  "123456",
			mockRep: func(m *mockSenderRep.SenderRepository) {
				m.On("GetUserLinkByResetToken", mock.Anything, "reset_token:123456").
					Return(expectedLink, nil)
				m.On("DeleteResetToken", mock.Anything, "reset_token:123456").Return(nil)
			},
			expectedLink: expectedLink,
		},
		{
			nameTest: "Token not found",
			tokenID:  "123456",
			mockRep: func(m *mockSenderRep.SenderRepository) {
				m.On("GetUserLinkByResetToken", mock.Anything, "reset_token:123456").
					Return("", common.ErrorNotExistingResetToken)
			},
			expectedErr: fmt.Errorf("rep.GetUserLinkByResetToken: %w", common.ErrorNotExistingResetToken).Error(),
		},
		{
			nameTest: "Error on delete token",
			tokenID:  "123456",
			mockRep: func(m *mockSenderRep.SenderRepository) {
				m.On("GetUserLinkByResetToken", mock.Anything, "reset_token:123456").
					Return(expectedLink, nil)
				m.On("DeleteResetToken", mock.Anything, "reset_token:123456").
					Return(errors.New("delete failed"))
			},
			expectedErr: fmt.Errorf("rep.DeleteResetToken: %w", errors.New("delete failed")).Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			rep := mockSenderRep.NewSenderRepository(t)
			if test.mockRep != nil {
				test.mockRep(rep)
			}

			svc := NewService(rep, nil, Config{}, Tools{
				CreatorResetKey: CreatorResetKey,
			})

			link, err := svc.GetUserLink(context.Background(), test.tokenID)

			if test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
				assert.Empty(t, link)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedLink, link)
			}
		})
	}
}
