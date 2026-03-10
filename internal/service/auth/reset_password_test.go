package auth

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	dbConnection "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/db_connection"
	mockAuthRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth/mock_auth_rep"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestResetPassword(t *testing.T) {
	tests := []struct {
		nameTest     string
		tokenID      string
		newPassword  string
		hasher       func(string) (string, error)
		mockBehavior func(m *mockAuthRep.AuthRepository)
	}{
		{
			nameTest:    "Success reset password",
			tokenID:     common.FixedResetTokenID,
			newPassword: "new_password",
			hasher:      spyHasher,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				validToken := dbConnection.ResetToken{
					ResetTokenID: common.FixedResetTokenID,
					UserID:       common.FixedUserUuiD,
					ExpiresAt:    time.Now().Add(15 * time.Minute),
				}

				m.On("GetResetToken", ctx, common.FixedResetTokenID).Return(validToken, nil)
				m.On("UpdatePassword", ctx, common.FixedUserUuiD, "hash_new_password").Return(nil)
				m.On("DeleteResetToken", ctx, common.FixedResetTokenID).Return(nil)
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

			serviceAuth := NewAuthService(mockRepo, nil, test.hasher, nil, nil, nil)

			err := serviceAuth.ResetPassword(ctx, test.tokenID, test.newPassword)

			assert.NoError(t, err, "expected no error")
		})
	}
}

func TestResetPasswordError(t *testing.T) {
	targetUserID := uuid.New()

	tests := []struct {
		nameTest      string
		tokenID       string
		newPassword   string
		hasher        func(string) (string, error)
		mockBehavior  func(m *mockAuthRep.AuthRepository)
		expectedError error
	}{
		{
			nameTest:    "Error token expired",
			tokenID:     common.FixedResetTokenID,
			newPassword: "new_password",
			hasher:      spyHasher,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				m.On("GetResetToken", ctx, common.FixedResetTokenID).Return(dbConnection.ResetToken{}, common.ErrorResetTokenExpired)
			},
			expectedError: fmt.Errorf("rep.GetResetToken: %w", common.ErrorResetTokenExpired),
		},
		{
			nameTest:    "Error hasher fails",
			tokenID:     common.FixedResetTokenID,
			newPassword: "new_password",
			hasher:      spyHasherError,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				validToken := dbConnection.ResetToken{
					ResetTokenID: common.FixedResetTokenID,
					UserID:       targetUserID,
					ExpiresAt:    time.Now().Add(15 * time.Minute),
				}
				m.On("GetResetToken", ctx, common.FixedResetTokenID).Return(validToken, nil)
			},
			expectedError: errors.New("hasher: failed to create hash: \"error bcrypt\""),
		},
		{
			nameTest:    "Error update password in DB",
			tokenID:     common.FixedResetTokenID,
			newPassword: "new_password",
			hasher:      spyHasher,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				validToken := dbConnection.ResetToken{
					ResetTokenID: common.FixedResetTokenID,
					UserID:       targetUserID,
					ExpiresAt:    time.Now().Add(15 * time.Minute),
				}

				m.On("GetResetToken", ctx, common.FixedResetTokenID).Return(validToken, nil)
				m.On("UpdatePassword", ctx, targetUserID, "hash_new_password").
					Return(errors.New("db connection lost"))
			},
			expectedError: fmt.Errorf("rep.UpdatePassword: %w", errors.New("db connection lost")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			ctx := context.Background()

			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			serviceAuth := NewAuthService(mockRepo, nil, test.hasher, nil, nil, nil)

			err := serviceAuth.ResetPassword(ctx, test.tokenID, test.newPassword)

			assert.EqualError(t, err, test.expectedError.Error(), "incorrect error message")
		})
	}
}
