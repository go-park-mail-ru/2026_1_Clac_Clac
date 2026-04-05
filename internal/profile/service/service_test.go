package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/service/dto"
	mockProfileRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/service/mock_profile_rep"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProfileUser(t *testing.T) {
	targetUserID := common.FixedUserUuiD

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

			profileService := NewService(mockProfileRepo, nil, "")
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

func TestUpdateProfile(t *testing.T) {
	targetUserLink := common.FixedUserUuiD

	expectedUser := repositoryDto.UpdatedInfo{
		Link:            targetUserLink,
		NameUser:        "Artem",
		DescriptionUser: "bobr",
	}

	tests := []struct {
		nameTest      string
		info          dto.UpdatedUserInfo
		mockBehavior  func(m *mockProfileRep.ProfileRepository)
		expectedError error
	}{
		{
			nameTest: "Success update profile",
			info: dto.UpdatedUserInfo{
				Link:        targetUserLink,
				DisplayName: "Artem",
				Description: "bobr",
			},
			mockBehavior: func(m *mockProfileRep.ProfileRepository) {
				m.On("UpdateProfile", mock.Anything, expectedUser).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error update profile",
			info: dto.UpdatedUserInfo{
				Link:        targetUserLink,
				DisplayName: "Artem",
				Description: "bobr",
			},
			mockBehavior: func(m *mockProfileRep.ProfileRepository) {
				m.On("UpdateProfile", mock.Anything, expectedUser).Return(errors.New("can not update"))
			},
			expectedError: fmt.Errorf("rep.UpdateProfile: %w", errors.New("can not update")),
		},
	}

	for _, test := range tests {
		mockRep := mockProfileRep.NewProfileRepository(t)
		if test.mockBehavior != nil {
			test.mockBehavior(mockRep)
		}

		srv := NewService(mockRep, nil, "")

		err := srv.UpdateProfile(context.Background(), test.info)

		assert.Equal(t, test.expectedError, err)
	}
}

func TestUpdateAvatar(t *testing.T) {
	targetUserLink := common.FixedUserUuiD
	validJpgBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}

	buffer := &bytes.Buffer{}
	writer := multipart.NewWriter(buffer)

	part, err := writer.CreateFormFile("avatar", "avatar.jpg")
	assert.NoError(t, err, "not wait error")

	_, err = part.Write(validJpgBytes)
	assert.NoError(t, err, "not wait error")

	err = writer.Close()
	assert.NoError(t, err, "not wait error")

	objectKey := "profiles/123-12/23.jpg"
	fullUrl := "https://nexus/profiles/123-12/23.jpg"
	baseUrl := "https://nexus/"

	dbErr := errors.New("fail db upload")
	deleteErr := errors.New("fail delete s3")

	tests := []struct {
		nameTest      string
		userLink      uuid.UUID
		file          io.Reader
		extension     string
		mockBehavior  func(m *mockProfileRep.ProfileRepository)
		expectedURL   string
		expectedError error
	}{
		{
			nameTest:  "Success update avatar",
			userLink:  targetUserLink,
			file:      buffer,
			extension: "image/jpg",
			mockBehavior: func(m *mockProfileRep.ProfileRepository) {
				m.On("UploadAvatarS3", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(objectKey, nil)
				m.On("UploadURLAvatar", mock.Anything, targetUserLink, mock.Anything).Return(nil)
			},
			expectedURL:   fullUrl,
			expectedError: nil,
		},
		{
			nameTest:  "Error update in S3",
			userLink:  targetUserLink,
			file:      buffer,
			extension: "image/jpg",
			mockBehavior: func(m *mockProfileRep.ProfileRepository) {
				m.On("UploadAvatarS3", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("fail upload"))
			},
			expectedURL:   "",
			expectedError: fmt.Errorf("UploadAvatar: %w", errors.New("fail upload")),
		},
		{
			nameTest:  "Error upload DB and success rollback S3",
			userLink:  targetUserLink,
			file:      buffer,
			extension: "image/jpg",
			mockBehavior: func(m *mockProfileRep.ProfileRepository) {
				m.On("UploadAvatarS3", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(objectKey, nil)
				m.On("UploadURLAvatar", mock.Anything, targetUserLink, mock.Anything).Return(dbErr)
				m.On("DeleteAvatarS3", mock.Anything, objectKey).Return(nil)
			},
			expectedURL:   "",
			expectedError: fmt.Errorf("rep.UploadAvatarURL: %w", dbErr),
		},
		{
			nameTest:  "Error upload DB and error rollback S3",
			userLink:  targetUserLink,
			file:      buffer,
			extension: "image/jpg",
			mockBehavior: func(m *mockProfileRep.ProfileRepository) {
				m.On("UploadAvatarS3", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(objectKey, nil)
				m.On("UploadURLAvatar", mock.Anything, targetUserLink, mock.Anything).Return(dbErr)
				m.On("DeleteAvatarS3", mock.Anything, objectKey).Return(deleteErr)
			},
			expectedURL:   "",
			expectedError: errors.Join(fmt.Errorf("rep.UploadAvatarURL: %w", dbErr), deleteErr),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := mockProfileRep.NewProfileRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRep)
			}

			srv := NewService(mockRep, GenerateAvatarKey, baseUrl)

			fullKey, err := srv.UpdateAvatar(context.Background(), dto.UpdatedAvatar{
				UserLink: test.userLink,
				File:     buffer,
				MimeType: "image/jpg",
			})

			if test.expectedError != nil {
				assert.Equal(t, test.expectedError, err)
			}

			assert.Equal(t, test.expectedURL, fullKey)
		})
	}
}

func TestDeleteAvatar(t *testing.T) {
	targetUserLink := common.FixedUserUuiD
	existingAvatarKey := "profiles/123-12/23.jpg"

	tests := []struct {
		nameTest      string
		userLink      uuid.UUID
		mockBehavior  func(m *mockProfileRep.ProfileRepository)
		expectedError error
	}{
		{
			nameTest: "Success delete avatar",
			userLink: targetUserLink,
			mockBehavior: func(m *mockProfileRep.ProfileRepository) {
				m.On("GetAvatarKey", mock.Anything, targetUserLink).Return(existingAvatarKey, nil)
				m.On("DeleteAvatarS3", mock.Anything, existingAvatarKey).Return(nil)
				m.On("DeleteURLAvatar", mock.Anything, targetUserLink).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Success empty avatar key",
			userLink: targetUserLink,
			mockBehavior: func(m *mockProfileRep.ProfileRepository) {
				m.On("GetAvatarKey", mock.Anything, targetUserLink).Return("", nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error get avatar key",
			userLink: targetUserLink,
			mockBehavior: func(m *mockProfileRep.ProfileRepository) {
				m.On("GetAvatarKey", mock.Anything, targetUserLink).Return("", errors.New("db error"))
			},
			expectedError: fmt.Errorf("rep.GetAvatarKey: %w", errors.New("db error")),
		},
		{
			nameTest: "Error delete S3",
			userLink: targetUserLink,
			mockBehavior: func(m *mockProfileRep.ProfileRepository) {
				m.On("GetAvatarKey", mock.Anything, targetUserLink).Return(existingAvatarKey, nil)
				m.On("DeleteAvatarS3", mock.Anything, existingAvatarKey).Return(errors.New("s3 error"))
			},
			expectedError: fmt.Errorf("rep.DeleteAvatar: %w", errors.New("s3 error")),
		},
		{
			nameTest: "Error delete URL avatar DB",
			userLink: targetUserLink,
			mockBehavior: func(m *mockProfileRep.ProfileRepository) {
				m.On("GetAvatarKey", mock.Anything, targetUserLink).Return(existingAvatarKey, nil)
				m.On("DeleteAvatarS3", mock.Anything, existingAvatarKey).Return(nil)
				m.On("DeleteURLAvatar", mock.Anything, targetUserLink).Return(errors.New("db error"))
			},
			expectedError: fmt.Errorf("rep.DeleteAvatarURL: %w", errors.New("db error")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := mockProfileRep.NewProfileRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRep)
			}

			srv := NewService(mockRep, nil, "")
			err := srv.DeleteAvatar(context.Background(), test.userLink)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err, "not wait error")
			}
		})
	}
}
