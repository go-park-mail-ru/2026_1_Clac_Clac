package tests

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/models"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/service"
	mockProfileRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/service/tests/mock_profile_rep"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProfileUser(t *testing.T) {
	targetUserID := uuid.New()

	expectedUser := models.User{
		ID:          targetUserID,
		DisplayName: "Artem",
		Email:       "test@mail.ru",
	}

	someRepoError := errors.New("database connection lost")

	tests := []struct {
		nameTest      string
		userID        uuid.UUID
		mockBehavior  func(m *mockProfileRep.ProfileRepository)
		expectedUser  models.User
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
				m.On("GetProfile", mock.Anything, targetUserID).Return(models.User{}, someRepoError)
			},
			expectedUser:  models.User{},
			expectedError: fmt.Errorf("rep.GetProfile: %w", someRepoError),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockProfileRepo := mockProfileRep.NewProfileRepository(t)

			if test.mockBehavior != nil {
				test.mockBehavior(mockProfileRepo)
			}

			profileService := service.NewProfileService(mockProfileRepo)
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
