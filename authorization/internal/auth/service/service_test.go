package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/auth/repository/dto"
	mockAuthRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/auth/service/mock_auth_rep"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testSessionLifetime = 24 * time.Hour

func spyGenerator() (string, error) {
	return "sessionCLAC", nil
}

func spyGeneratorError() (string, error) {
	return "", errors.New("generator failed")
}

func TestCreateSession(t *testing.T) {
	userUUID := uuid.New()

	tests := []struct {
		nameTest     string
		userLink     uuid.UUID
		generator    func() (string, error)
		mockBehavior func(m *mockAuthRep.AuthRepository)
		expectedID   string
		expectedErr  string
	}{
		{
			nameTest:  "Success create session",
			userLink:  userUUID,
			generator: spyGenerator,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				expected := repositoryDto.SessionEntity{
					SessionKey: "session:sessionCLAC",
					UserLink:   userUUID,
					LifeTime:   testSessionLifetime,
				}
				m.On("AddSession", mock.Anything, expected).Return(nil)
			},
			expectedID: "sessionCLAC",
		},
		{
			nameTest:    "Error generator fails",
			userLink:    userUUID,
			generator:   spyGeneratorError,
			expectedErr: fmt.Errorf("tools.generatorSessionID: %w", errors.New("generator failed")).Error(),
		},
		{
			nameTest:  "Error AddSession fails",
			userLink:  userUUID,
			generator: spyGenerator,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("AddSession", mock.Anything, mock.Anything).Return(errors.New("redis error"))
			},
			expectedErr: fmt.Errorf("rep.AddSession: %w", errors.New("redis error")).Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			svc := NewService(mockRepo, Config{SessionLifetime: testSessionLifetime}, Tools{
				GeneratorSessionID: test.generator,
				CreateSessionKey:   CreateSessionKey,
			})

			sessionID, err := svc.CreateSession(context.Background(), test.userLink)

			if test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
				assert.Empty(t, sessionID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedID, sessionID)
			}
		})
	}
}

func TestGetUserLink(t *testing.T) {
	expectedLink := uuid.New().String()

	tests := []struct {
		nameTest     string
		sessionID    string
		mockBehavior func(m *mockAuthRep.AuthRepository)
		expectedLink string
		expectedErr  string
	}{
		{
			nameTest:  "Success get user link",
			sessionID: "sessionCLAC",
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("GetUserLink", mock.Anything, "session:sessionCLAC").Return(expectedLink, nil)
			},
			expectedLink: expectedLink,
		},
		{
			nameTest:  "Error repository fails",
			sessionID: "sessionCLAC",
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("GetUserLink", mock.Anything, "session:sessionCLAC").Return("", errors.New("not found"))
			},
			expectedErr: fmt.Errorf("rep.GetUserLink: %w", errors.New("not found")).Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			svc := NewService(mockRepo, Config{}, Tools{
				CreateSessionKey: CreateSessionKey,
			})

			link, err := svc.GetUserLink(context.Background(), test.sessionID)

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

func TestDeleteSession(t *testing.T) {
	tests := []struct {
		nameTest     string
		sessionID    string
		mockBehavior func(m *mockAuthRep.AuthRepository)
		expectedErr  string
	}{
		{
			nameTest:  "Success delete session",
			sessionID: "sessionCLAC",
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("DeleteSession", mock.Anything, "session:sessionCLAC").Return(nil)
			},
		},
		{
			nameTest:  "Error repository fails",
			sessionID: "sessionCLAC",
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("DeleteSession", mock.Anything, "session:sessionCLAC").Return(errors.New("redis error"))
			},
			expectedErr: fmt.Errorf("rep.DeleteSession: %w", errors.New("redis error")).Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			svc := NewService(mockRepo, Config{}, Tools{
				CreateSessionKey: CreateSessionKey,
			})

			err := svc.DeleteSession(context.Background(), test.sessionID)

			if test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtendSession(t *testing.T) {
	tests := []struct {
		nameTest     string
		sessionID    string
		mockBehavior func(m *mockAuthRep.AuthRepository)
		expectedErr  string
	}{
		{
			nameTest:  "Success extend session",
			sessionID: "sessionCLAC",
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				expected := repositoryDto.ExtendedSession{
					SessionKey: "session:sessionCLAC",
					Expiration: testSessionLifetime,
				}
				m.On("ExtendSession", mock.Anything, expected).Return(nil)
			},
		},
		{
			nameTest:  "Error repository fails",
			sessionID: "sessionCLAC",
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("ExtendSession", mock.Anything, mock.Anything).Return(errors.New("session not found"))
			},
			expectedErr: fmt.Errorf("rep.ExtendSession: %w", errors.New("session not found")).Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			svc := NewService(mockRepo, Config{SessionLifetime: testSessionLifetime}, Tools{
				CreateSessionKey: CreateSessionKey,
			})

			err := svc.ExtendSession(context.Background(), test.sessionID)

			if test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
