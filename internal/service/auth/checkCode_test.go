package auth

import (
	"context"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	dbConnection "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/db_connection"
	mockAuthRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth/mock_auth_rep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckCode(t *testing.T) {
	validToken := "123456"

	tests := []struct {
		nameTest      string
		tokenID       string
		mockBehavior  func(m *mockAuthRep.AuthRepository)
		expectedError error
	}{
		{
			nameTest: "Success check code",
			tokenID:  validToken,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				token := dbConnection.ResetToken{
					ResetTokenID: validToken,
					ExpiresAt:    time.Now().Add(15 * time.Minute),
				}
				m.On("GetResetToken", mock.Anything, validToken).Return(token, nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error code expired",
			tokenID:  validToken,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				token := dbConnection.ResetToken{
					ResetTokenID: validToken,
					ExpiresAt:    time.Now().Add(-1 * time.Minute),
				}
				m.On("GetResetToken", mock.Anything, validToken).Return(token, nil)
				m.On("DeleteResetToken", mock.Anything, validToken).Return(nil)
			},
			expectedError: common.ErrorResetTokenExpired,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			service := NewAuthService(mockRepo, nil, nil, nil, nil, nil)
			err := service.CheckRecoveryCode(context.Background(), test.tokenID)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
