package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	repository "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/auth"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth/mocks"
	"github.com/stretchr/testify/assert"
)

func TestLogOut(t *testing.T) {
	tests := []struct {
		nameTest     string
		sessionID    string
		hasher       func(string) (string, error)
		checker      func(string, string) error
		generator    func() (string, error)
		mockBehavior func(m *mocks.Database)
	}{
		{
			nameTest:  "Success log out",
			sessionID: common.FixedSessionID,
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mocks.Database) {
				ctx := context.Background()
				m.On("DeleteSession", ctx, common.FixedSessionID).Return(nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mocks.NewDatabase(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			ctx := context.Background()

			serviceLogOut := NewAuthService(mockRepo, test.hasher, test.checker, test.generator)

			err := serviceLogOut.LogOut(ctx, test.sessionID)
			assert.NoError(t, err, "not expected error")
		})
	}
}

func TestLogOutError(t *testing.T) {
	tests := []struct {
		nameTest      string
		sessionID     string
		hasher        func(string) (string, error)
		checker       func(string, string) error
		generator     func() (string, error)
		mockBehavior  func(m *mocks.Database)
		expectedError error
	}{
		{
			nameTest:  "Success log out",
			sessionID: common.FixedSessionID,
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mocks.Database) {
				ctx := context.Background()
				m.On("DeleteSession", ctx, common.FixedSessionID).Return(repository.ErrorNotExistingSession)
			},
			expectedError: fmt.Errorf("repo.DeleteSession: %w", repository.ErrorNotExistingSession),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mocks.NewDatabase(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			ctx := context.Background()

			serviceLogOut := NewAuthService(mockRepo, test.hasher, test.checker, test.generator)

			err := serviceLogOut.LogOut(ctx, test.sessionID)
			assert.Error(t, err, "expected error")
			assert.EqualError(t, test.expectedError, err.Error(), "incorrect error message")
		})
	}
}
