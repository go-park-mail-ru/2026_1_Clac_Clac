package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/service/dto"
	mockProfileRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/service/mock_profile_rep"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProfileUser(t *testing.T) {
	targetUserID := uuid.New()

	expectedUser := dto.UserInfoResponce{
		Link:        targetUserID,
		DisplayName: "Artem",
		Email:       "test@mail.ru",
	}

	someRepoError := errors.New("database connection lost")

	tests := []struct {
		nameTest      string
		userID        uuid.UUID
		mockBehavior  func(m *mockProfileRep.ProfileRepository)
		expectedUser  dto.UserInfoResponce
		expectedError error
	}{
		{
			nameTest: "Success get profile",
			userID:   targetUserID,
			mockBehavior: func(m *mockProfileRep.ProfileRepository) {
				m.On("GetProfile", mock.Anything, targetUserID).Return(expectedUser, nil)
			},
			expectedUser:  expectedUser,
			expectedError: nil,
		},
		{
			nameTest: "Error from repository",
			userID:   targetUserID,
			mockBehavior: func(m *mockProfileRep.ProfileRepository) {
				m.On("GetProfile", mock.Anything, targetUserID).Return(dto.UserInfoResponce{}, someRepoError)
			},
			expectedUser:  dto.UserInfoResponce{},
			expectedError: fmt.Errorf("rep.GetProfile: %w", someRepoError),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockProfileRepo := mockProfileRep.NewProfileRepository(t)

			if test.mockBehavior != nil {
				test.mockBehavior(mockProfileRepo)
			}

			profileService := NewService(mockProfileRepo)
			ctx := context.Background()

			user, err := profileService.GetProfileUser(ctx, test.userID)

			assert.Equal(t, test.expectedUser, user, "incorrect user returned")

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error(), "incorrect error message")
			} else {
				assert.NoError(t, err, "unexpected error")
			}
		})
	}
}
