package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/common"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/repository/dto"
	dto "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/service/dto"
	mockAuthRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/service/mock_auth_rep"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var fixedUserUUID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

func newService(rep *mockAuthRep.AuthRepository) *Service {
	return NewService(rep, Config{BaseURLAvatar: "https://cdn.example.com"}, Tools{
		Hasher:            spyHasher,
		Checker:           spyChecker,
		GenerateAvatarKey: spyGenerator,
	})
}

func TestLogIn(t *testing.T) {
	tests := []struct {
		nameTest     string
		email        string
		password     string
		mockBehavior func(m *mockAuthRep.AuthRepository)
		expectedUser dto.UserInfo
	}{
		{
			nameTest: "Success login",
			email:    "user@mail.ru",
			password: "hash_secret",
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("GetUser", mock.Anything, "user@mail.ru").Return(repositoryDto.UserEntity{
					Link:         fixedUserUUID,
					DisplayName:  "Artem",
					Email:        "user@mail.ru",
					PasswordHash: "hash_secret",
				}, nil)
			},
			expectedUser: dto.UserInfo{
				Link:        fixedUserUUID,
				DisplayName: "Artem",
				Email:       "user@mail.ru",
				AvatarURL:   "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			m := mockAuthRep.NewAuthRepository(t)
			test.mockBehavior(m)

			user, err := newService(m).LogIn(context.Background(), dto.LogInUser{
				Email:    test.email,
				Password: test.password,
			})

			assert.NoError(t, err)
			assert.Equal(t, test.expectedUser, user)
		})
	}
}

func TestLogInError(t *testing.T) {
	tests := []struct {
		nameTest      string
		email         string
		password      string
		mockBehavior  func(m *mockAuthRep.AuthRepository)
		expectedError error
	}{
		{
			nameTest: "Error user not found",
			email:    "user@mail.ru",
			password: "secret",
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("GetUser", mock.Anything, "user@mail.ru").Return(repositoryDto.UserEntity{}, common.ErrorNonexistentEmail)
			},
			expectedError: fmt.Errorf("rep.GetUser: %w", common.ErrorNonexistentEmail),
		},
		{
			nameTest: "Error wrong password",
			email:    "user@mail.ru",
			password: "wrong_password",
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("GetUser", mock.Anything, "user@mail.ru").Return(repositoryDto.UserEntity{
					PasswordHash: "correct_hash",
				}, nil)
			},
			expectedError: fmt.Errorf("rep.CheckPassword: %w", ErrorWrongPassword),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			m := mockAuthRep.NewAuthRepository(t)
			test.mockBehavior(m)

			_, err := newService(m).LogIn(context.Background(), dto.LogInUser{
				Email:    test.email,
				Password: test.password,
			})

			assert.EqualError(t, err, test.expectedError.Error())
		})
	}
}

func TestRegister(t *testing.T) {
	tests := []struct {
		nameTest     string
		displayName  string
		email        string
		password     string
		mockBehavior func(m *mockAuthRep.AuthRepository)
	}{
		{
			nameTest:    "Success registration",
			displayName: "Artem",
			email:       "test@mail.ru",
			password:    "password123",
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("AddUser", mock.Anything, mock.AnythingOfType("dto.UserInitialize")).Return(nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			m := mockAuthRep.NewAuthRepository(t)
			test.mockBehavior(m)

			user, err := newService(m).Register(context.Background(), dto.RegistrationUser{
				DisplayName: test.displayName,
				Email:       test.email,
				Password:    test.password,
			})

			assert.NoError(t, err)
			assert.Equal(t, test.email, user.Email)
			assert.Equal(t, test.displayName, user.DisplayName)
		})
	}
}

func TestRegisterError(t *testing.T) {
	tests := []struct {
		nameTest      string
		tools         Tools
		mockBehavior  func(m *mockAuthRep.AuthRepository)
		expectedError error
	}{
		{
			nameTest: "Error existing user",
			tools:    Tools{Hasher: spyHasher, Checker: spyChecker, GenerateAvatarKey: spyGenerator},
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("AddUser", mock.Anything, mock.AnythingOfType("dto.UserInitialize")).Return(common.ErrorExistingUser)
			},
			expectedError: fmt.Errorf("rep.AddUser: %w", common.ErrorExistingUser),
		},
		{
			nameTest:      "Error hash password",
			tools:         Tools{Hasher: spyHasherError, Checker: spyChecker, GenerateAvatarKey: spyGenerator},
			mockBehavior:  func(m *mockAuthRep.AuthRepository) {},
			expectedError: fmt.Errorf("HashPassword: %w: %q", ErrorCreateHash, "error bcrypt"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			m := mockAuthRep.NewAuthRepository(t)
			test.mockBehavior(m)

			svc := NewService(m, Config{}, test.tools)
			_, err := svc.Register(context.Background(), dto.RegistrationUser{
				DisplayName: "Artem",
				Email:       "test@mail.ru",
				Password:    "password123",
			})

			assert.EqualError(t, err, test.expectedError.Error())
		})
	}
}

func TestResetPassword(t *testing.T) {
	t.Run("Success reset password", func(t *testing.T) {
		m := mockAuthRep.NewAuthRepository(t)
		m.On("UpdatePassword", mock.Anything, fixedUserUUID, "hash_newpass").Return(nil)

		svc := NewService(m, Config{}, Tools{Hasher: spyHasher})
		err := svc.ResetPassword(context.Background(), dto.ResetPasswordInfo{
			UserLink:    fixedUserUUID.String(),
			NewPassword: "newpass",
		})

		assert.NoError(t, err)
	})
}

func TestResetPasswordError(t *testing.T) {
	_, invalidUUIDErr := uuid.Parse("bad-uuid")

	tests := []struct {
		nameTest      string
		info          dto.ResetPasswordInfo
		tools         Tools
		mockBehavior  func(m *mockAuthRep.AuthRepository)
		expectedError error
	}{
		{
			nameTest:      "Error invalid UUID",
			info:          dto.ResetPasswordInfo{UserLink: "bad-uuid", NewPassword: "pass"},
			tools:         Tools{Hasher: spyHasher},
			mockBehavior:  func(m *mockAuthRep.AuthRepository) {},
			expectedError: fmt.Errorf("uuid.Parse: %w", invalidUUIDErr),
		},
		{
			nameTest:      "Error hasher fails",
			info:          dto.ResetPasswordInfo{UserLink: fixedUserUUID.String(), NewPassword: "pass"},
			tools:         Tools{Hasher: spyHasherError},
			mockBehavior:  func(m *mockAuthRep.AuthRepository) {},
			expectedError: fmt.Errorf("hasher: %w", fmt.Errorf("%w: %q", ErrorCreateHash, "error bcrypt")),
		},
		{
			nameTest: "Error update password in DB",
			info:     dto.ResetPasswordInfo{UserLink: fixedUserUUID.String(), NewPassword: "pass"},
			tools:    Tools{Hasher: spyHasher},
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("UpdatePassword", mock.Anything, fixedUserUUID, "hash_pass").Return(errors.New("db error"))
			},
			expectedError: fmt.Errorf("rep.UpdatePassword: %w", errors.New("db error")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			m := mockAuthRep.NewAuthRepository(t)
			test.mockBehavior(m)

			svc := NewService(m, Config{}, test.tools)
			err := svc.ResetPassword(context.Background(), test.info)

			assert.EqualError(t, err, test.expectedError.Error())
		})
	}
}

func TestGetUserLink(t *testing.T) {
	t.Run("Success get user link", func(t *testing.T) {
		m := mockAuthRep.NewAuthRepository(t)
		m.On("GetUserLink", mock.Anything, "user@mail.ru").Return(fixedUserUUID, nil)

		link, err := newService(m).GetUserLink(context.Background(), "user@mail.ru")

		assert.NoError(t, err)
		assert.Equal(t, fixedUserUUID.String(), link)
	})

	t.Run("Error user not found", func(t *testing.T) {
		m := mockAuthRep.NewAuthRepository(t)
		m.On("GetUserLink", mock.Anything, "user@mail.ru").Return(uuid.Nil, common.ErrorNonexistentEmail)

		_, err := newService(m).GetUserLink(context.Background(), "user@mail.ru")

		assert.ErrorIs(t, err, common.ErrorNonexistentEmail)
	})
}

func TestGetUserByEmail(t *testing.T) {
	t.Run("Success get user by email", func(t *testing.T) {
		m := mockAuthRep.NewAuthRepository(t)
		m.On("GetUser", mock.Anything, "user@mail.ru").Return(repositoryDto.UserEntity{
			Link:        fixedUserUUID,
			DisplayName: "Artem",
			Email:       "user@mail.ru",
		}, nil)

		user, err := newService(m).GetUserByEmail(context.Background(), "user@mail.ru")

		assert.NoError(t, err)
		assert.Equal(t, "user@mail.ru", user.Email)
		assert.Equal(t, fixedUserUUID, user.Link)
	})

	t.Run("Error user not found", func(t *testing.T) {
		m := mockAuthRep.NewAuthRepository(t)
		m.On("GetUser", mock.Anything, "unknown@mail.ru").Return(repositoryDto.UserEntity{}, common.ErrorNonexistentEmail)

		_, err := newService(m).GetUserByEmail(context.Background(), "unknown@mail.ru")

		assert.ErrorIs(t, err, common.ErrorNonexistentEmail)
	})
}

func TestEnsureUserByEmail(t *testing.T) {
	info := dto.RegistrationUser{DisplayName: "Artem", Email: "user@mail.ru"}

	tests := []struct {
		nameTest     string
		mockBehavior func(m *mockAuthRep.AuthRepository)
		expectError  bool
	}{
		{
			nameTest: "User already exists",
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("GetUser", mock.Anything, info.Email).Return(repositoryDto.UserEntity{
					Link:  fixedUserUUID,
					Email: info.Email,
				}, nil)
			},
			expectError: false,
		},
		{
			nameTest: "User does not exist — registers successfully",
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("GetUser", mock.Anything, info.Email).Return(repositoryDto.UserEntity{}, common.ErrorNonexistentEmail)
				m.On("AddUser", mock.Anything, mock.AnythingOfType("dto.UserInitialize")).Return(nil)
			},
			expectError: false,
		},
		{
			nameTest: "Error get user — not a nonexistent error",
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("GetUser", mock.Anything, info.Email).Return(repositoryDto.UserEntity{}, errors.New("db error"))
			},
			expectError: true,
		},
		{
			nameTest: "User does not exist — register fails",
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("GetUser", mock.Anything, info.Email).Return(repositoryDto.UserEntity{}, common.ErrorNonexistentEmail)
				m.On("AddUser", mock.Anything, mock.AnythingOfType("dto.UserInitialize")).Return(errors.New("cannot register"))
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			m := mockAuthRep.NewAuthRepository(t)
			test.mockBehavior(m)

			link, err := newService(m).EnsureUserByEmail(context.Background(), info)

			if test.expectError {
				assert.Error(t, err)
				assert.Empty(t, link)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, link)
			}
		})
	}
}

func TestGetProfile(t *testing.T) {
	t.Run("Success get profile without avatar", func(t *testing.T) {
		m := mockAuthRep.NewAuthRepository(t)
		m.On("GetProfile", mock.Anything, fixedUserUUID).Return(repositoryDto.UserInfoEntity{
			Link:            fixedUserUUID,
			DisplayName:     "Artem",
			DescriptionUser: "Developer",
			Email:           "user@mail.ru",
			AvatarKey:       "",
		}, nil)

		user, err := newService(m).GetProfile(context.Background(), fixedUserUUID)

		assert.NoError(t, err)
		assert.Equal(t, fixedUserUUID, user.Link)
		assert.Empty(t, user.AvatarURL)
	})

	t.Run("Success get profile with avatar", func(t *testing.T) {
		m := mockAuthRep.NewAuthRepository(t)
		m.On("GetProfile", mock.Anything, fixedUserUUID).Return(repositoryDto.UserInfoEntity{
			Link:        fixedUserUUID,
			DisplayName: "Artem",
			Email:       "user@mail.ru",
			AvatarKey:   "avatars/key.jpg",
		}, nil)

		user, err := newService(m).GetProfile(context.Background(), fixedUserUUID)

		assert.NoError(t, err)
		assert.Contains(t, user.AvatarURL, "avatars/key.jpg")
	})

	t.Run("Error user not found", func(t *testing.T) {
		m := mockAuthRep.NewAuthRepository(t)
		m.On("GetProfile", mock.Anything, fixedUserUUID).Return(repositoryDto.UserInfoEntity{}, common.ErrorNonexistentUser)

		_, err := newService(m).GetProfile(context.Background(), fixedUserUUID)

		assert.ErrorIs(t, err, common.ErrorNonexistentUser)
	})
}

func TestUpdateProfile(t *testing.T) {
	t.Run("Success update profile", func(t *testing.T) {
		m := mockAuthRep.NewAuthRepository(t)
		m.On("UpdateProfile", mock.Anything, repositoryDto.UpdatedInfo{
			Link:            fixedUserUUID,
			NameUser:        "NewName",
			DescriptionUser: "New description",
		}).Return(nil)

		err := newService(m).UpdateProfile(context.Background(), dto.UpdatedUserInfo{
			Link:        fixedUserUUID,
			DisplayName: "NewName",
			Description: "New description",
		})

		assert.NoError(t, err)
	})

	t.Run("Error repository failure", func(t *testing.T) {
		m := mockAuthRep.NewAuthRepository(t)
		m.On("UpdateProfile", mock.Anything, mock.Anything).Return(errors.New("db error"))

		err := newService(m).UpdateProfile(context.Background(), dto.UpdatedUserInfo{
			Link: fixedUserUUID,
		})

		assert.Error(t, err)
	})
}

func TestUpdateAvatar(t *testing.T) {
	t.Run("Success update avatar", func(t *testing.T) {
		m := mockAuthRep.NewAuthRepository(t)
		m.On("UploadAvatarS3", mock.Anything, mock.Anything, mock.AnythingOfType("string"), "image/jpeg").Return("avatars/sessionCLAC.jpeg", nil)
		m.On("UploadURLAvatar", mock.Anything, fixedUserUUID, "avatars/sessionCLAC.jpeg").Return(nil)

		svc := NewService(m, Config{BaseURLAvatar: "https://cdn.example.com"}, Tools{
			Hasher:            spyHasher,
			Checker:           spyChecker,
			GenerateAvatarKey: spyGenerator,
		})

		url, err := svc.UpdateAvatar(context.Background(), dto.UpdatedAvatar{
			UserLink: fixedUserUUID,
			MimeType: "image/jpeg",
			File:     strings.NewReader("fake-image-data"),
		})

		assert.NoError(t, err)
		assert.Contains(t, url, "avatars/sessionCLAC.jpeg")
	})

	t.Run("Error avatar key generation fails", func(t *testing.T) {
		m := mockAuthRep.NewAuthRepository(t)

		svc := NewService(m, Config{}, Tools{
			GenerateAvatarKey: func() (string, error) {
				return "", errors.New("key gen error")
			},
		})

		_, err := svc.UpdateAvatar(context.Background(), dto.UpdatedAvatar{
			UserLink: fixedUserUUID,
			MimeType: "image/jpeg",
			File:     strings.NewReader("data"),
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot generate key")
	})

	t.Run("Error upload to S3 fails", func(t *testing.T) {
		m := mockAuthRep.NewAuthRepository(t)
		m.On("UploadAvatarS3", mock.Anything, mock.Anything, mock.AnythingOfType("string"), "image/png").Return("", errors.New("s3 error"))

		svc := NewService(m, Config{}, Tools{GenerateAvatarKey: spyGenerator})
		_, err := svc.UpdateAvatar(context.Background(), dto.UpdatedAvatar{
			UserLink: fixedUserUUID,
			MimeType: "image/png",
			File:     strings.NewReader("data"),
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "UploadAvatar")
	})

	t.Run("Error upload URL to DB fails — S3 object deleted", func(t *testing.T) {
		m := mockAuthRep.NewAuthRepository(t)
		m.On("UploadAvatarS3", mock.Anything, mock.Anything, mock.AnythingOfType("string"), "image/png").Return("avatars/key.png", nil)
		m.On("UploadURLAvatar", mock.Anything, fixedUserUUID, "avatars/key.png").Return(errors.New("db error"))
		m.On("DeleteAvatarS3", mock.Anything, "avatars/key.png").Return(nil)

		svc := NewService(m, Config{}, Tools{GenerateAvatarKey: spyGenerator})
		_, err := svc.UpdateAvatar(context.Background(), dto.UpdatedAvatar{
			UserLink: fixedUserUUID,
			MimeType: "image/png",
			File:     strings.NewReader("data"),
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rep.UploadAvatarURL")
	})
}

func TestDeleteAvatar(t *testing.T) {
	t.Run("Success delete avatar", func(t *testing.T) {
		m := mockAuthRep.NewAuthRepository(t)
		m.On("GetAvatarKey", mock.Anything, fixedUserUUID).Return("avatars/key.jpg", nil)
		m.On("DeleteURLAvatar", mock.Anything, fixedUserUUID).Return(nil)
		m.On("DeleteAvatarS3", mock.Anything, "avatars/key.jpg").Return(nil)

		err := newService(m).DeleteAvatar(context.Background(), fixedUserUUID)
		assert.NoError(t, err)
	})

	t.Run("Success no avatar to delete", func(t *testing.T) {
		m := mockAuthRep.NewAuthRepository(t)
		m.On("GetAvatarKey", mock.Anything, fixedUserUUID).Return("", nil)

		err := newService(m).DeleteAvatar(context.Background(), fixedUserUUID)
		assert.NoError(t, err)
	})

	t.Run("Error get avatar key fails", func(t *testing.T) {
		m := mockAuthRep.NewAuthRepository(t)
		m.On("GetAvatarKey", mock.Anything, fixedUserUUID).Return("", errors.New("db error"))

		err := newService(m).DeleteAvatar(context.Background(), fixedUserUUID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rep.GetAvatarKey")
	})

	t.Run("Error delete URL from DB fails", func(t *testing.T) {
		m := mockAuthRep.NewAuthRepository(t)
		m.On("GetAvatarKey", mock.Anything, fixedUserUUID).Return("avatars/key.jpg", nil)
		m.On("DeleteURLAvatar", mock.Anything, fixedUserUUID).Return(errors.New("db error"))

		err := newService(m).DeleteAvatar(context.Background(), fixedUserUUID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rep.DeleteAvatarURL")
	})

	t.Run("Error delete from S3 fails", func(t *testing.T) {
		m := mockAuthRep.NewAuthRepository(t)
		m.On("GetAvatarKey", mock.Anything, fixedUserUUID).Return("avatars/key.jpg", nil)
		m.On("DeleteURLAvatar", mock.Anything, fixedUserUUID).Return(nil)
		m.On("DeleteAvatarS3", mock.Anything, "avatars/key.jpg").Return(errors.New("s3 error"))

		err := newService(m).DeleteAvatar(context.Background(), fixedUserUUID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rep.DeleteAvatar")
	})
}

func TestUpdateAvatarMimeTypes(t *testing.T) {
	mimeTypes := []struct {
		mime   string
		format string
	}{
		{"image/jpg", ".jpg"},
		{"image/jpeg", ".jpeg"},
		{"image/png", ".png"},
		{"image/webp", ".webp"},
	}

	for _, tc := range mimeTypes {
		t.Run("mime "+tc.mime, func(t *testing.T) {
			m := mockAuthRep.NewAuthRepository(t)
			expectedKey := fmt.Sprintf("%s/sessionCLAC%s", fixedUserUUID.String(), tc.format)
			m.On("UploadAvatarS3", mock.Anything, mock.Anything, expectedKey, tc.mime).Return(expectedKey, nil)
			m.On("UploadURLAvatar", mock.Anything, fixedUserUUID, expectedKey).Return(nil)

			svc := NewService(m, Config{BaseURLAvatar: "https://cdn.example.com"}, Tools{
				GenerateAvatarKey: spyGenerator,
			})

			_, err := svc.UpdateAvatar(context.Background(), dto.UpdatedAvatar{
				UserLink: fixedUserUUID,
				MimeType: tc.mime,
				File:     strings.NewReader("data"),
			})

			assert.NoError(t, err)
		})
	}
}
