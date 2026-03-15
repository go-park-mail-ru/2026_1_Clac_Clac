package tests

import (
	"context"
	"fmt"
	"testing"

	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service"
	mockAuthRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service/tests/mock_auth_rep"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/stretchr/testify/assert"
)

func TestLogOut(t *testing.T) {
	tests := []struct {
		nameTest     string
		sessionID    string
		hasher       func(string) (string, error)
		checker      func(string, string) error
		generator    func() (string, error)
		mockBehavior func(m *mockAuthRep.AuthRepository)
	}{
		{
			nameTest:  "Success log out",
			sessionID: common.FixedSessionID,
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				m.On("DeleteSession", ctx, common.FixedSessionID).Return(nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			ctx := context.Background()

			serviceLogOut := service.NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil)

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
		mockBehavior  func(m *mockAuthRep.AuthRepository)
		expectedError error
	}{
		{
			nameTest:  "Error session not found",
			sessionID: common.FixedSessionID,
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				m.On("DeleteSession", ctx, common.FixedSessionID).Return(common.ErrorNotExistingSession)
			},
			expectedError: fmt.Errorf("rep.DeleteSession: %w", common.ErrorNotExistingSession),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			ctx := context.Background()

			serviceLogOut := service.NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil)

			err := serviceLogOut.LogOut(ctx, test.sessionID)
			assert.Error(t, err, "expected error")
			assert.EqualError(t, test.expectedError, err.Error(), "incorrect error message")
		})
	}
}
