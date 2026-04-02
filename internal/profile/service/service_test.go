package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/service/dto"
	mockProfileRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/service/mock_profile_rep"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProfileUser(t *testing.T) {
	targetUserID := uuid.New()

	expectedUser := dto.UserInfo{
		Link:        targetUserID,
		DisplayName: "Artem",
		Email:       "test@mail.ru",
	}

	someRepoError := errors.New("database connection lost")

	tests := []struct {
		nameTest      string
		userID        uuid.UUID
		mockBehavior  func(m *mockProfileRep.ProfileRepository)
		expectedUser  dto.UserInfo
		expectedError error
	}{
		{
			nameTest: "Success get profile",
			userID:   targetUserID,
			mockBehavior: func(m *mockProfileRep.ProfileRepository) {
				m.On("GetProfile", mock.Anything, targetUserID).Return(repositoryDto.UserInfoEntity{
					Link:        targetUserID,
					DisplayName: "Artem",
					Email:       "test@mail.ru"}, nil)
			},
			expectedUser:  expectedUser,
			expectedError: nil,
		},
		{
			nameTest: "Error from repository",
			userID:   targetUserID,
			mockBehavior: func(m *mockProfileRep.ProfileRepository) {
				m.On("GetProfile", mock.Anything, targetUserID).Return(repositoryDto.UserInfoEntity{}, someRepoError)
			},
			expectedUser:  dto.UserInfo{},
			expectedError: fmt.Errorf("rep.GetProfile: %w", someRepoError),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockProfileRepo := mockProfileRep.NewProfileRepository(t)

			if test.mockBehavior != nil {
				test.mockBehavior(mockProfileRepo)
			}

			profileService := NewService(mockProfileRepo, nil, "") // TODO: править
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
