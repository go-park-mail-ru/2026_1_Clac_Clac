package usecase

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	mockUserClient "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/usecase/mock_user_client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessUserWithVK(t *testing.T) {
	tests := []struct {
		name         string
		accessToken  string
		email        string
		mockBehavior func(m *mockUserClient.UserClient)
		expectedLink uuid.UUID
		expectError  bool
	}{
		{
			name:        "Success",
			accessToken: "vk_token",
			email:       "test@vk.com",
			mockBehavior: func(m *mockUserClient.UserClient) {
				m.On("ProcessUserWithVK", context.Background(), "vk_token", "test@vk.com").Return(fixedUserLink, nil)
			},
			expectedLink: fixedUserLink,
			expectError:  false,
		},
		{
			name:        "ClientError",
			accessToken: "bad_token",
			email:       "test@vk.com",
			mockBehavior: func(m *mockUserClient.UserClient) {
				m.On("ProcessUserWithVK", context.Background(), "bad_token", "test@vk.com").Return(uuid.Nil, testError)
			},
			expectedLink: uuid.Nil,
			expectError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockUserClient.NewUserClient(t)
			tc.mockBehavior(m)

			link, err := NewUser(m).ProcessUserWithVK(context.Background(), tc.accessToken, tc.email)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedLink, link)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	cred := domain.Credentials{Email: "test@mail.ru", Password: "password"}
	expectedUser := domain.FullInfoUser{
		UserLink:    fixedUserLink,
		DisplayName: "Test",
		Email:       "test@mail.ru",
		AvatarURL:   "http://avatar.com",
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mockUserClient.UserClient)
		expectedUser domain.FullInfoUser
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockUserClient.UserClient) {
				m.On("GetUser", context.Background(), cred).Return(expectedUser, nil)
			},
			expectedUser: expectedUser,
			expectError:  false,
		},
		{
			name: "UserNotFound",
			mockBehavior: func(m *mockUserClient.UserClient) {
				m.On("GetUser", context.Background(), cred).Return(domain.FullInfoUser{}, common.ErrorNonexistentUser)
			},
			expectedUser: domain.FullInfoUser{},
			expectError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockUserClient.NewUserClient(t)
			tc.mockBehavior(m)

			user, err := NewUser(m).GetUser(context.Background(), cred)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedUser, user)
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	cred := domain.NewCredentialsUser{Email: "test@mail.ru", Password: "password", DisplayName: "Test"}
	expectedUser := domain.FullInfoUser{
		UserLink:    fixedUserLink,
		DisplayName: "Test",
		Email:       "test@mail.ru",
		AvatarURL:   "http://avatar.com",
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mockUserClient.UserClient)
		expectedUser domain.FullInfoUser
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockUserClient.UserClient) {
				m.On("CreateUser", context.Background(), cred).Return(expectedUser, nil)
			},
			expectedUser: expectedUser,
			expectError:  false,
		},
		{
			name: "AlreadyExists",
			mockBehavior: func(m *mockUserClient.UserClient) {
				m.On("CreateUser", context.Background(), cred).Return(domain.FullInfoUser{}, common.ErrorExistingUser)
			},
			expectedUser: domain.FullInfoUser{},
			expectError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockUserClient.NewUserClient(t)
			tc.mockBehavior(m)

			user, err := NewUser(m).CreateUser(context.Background(), cred)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedUser, user)
			}
		})
	}
}

func TestGetProfile(t *testing.T) {
	expectedUser := domain.FullInfoUser{UserLink: fixedUserLink, Email: "test@mail.ru", DisplayName: "Test"}

	tests := []struct {
		name         string
		mockBehavior func(m *mockUserClient.UserClient)
		expectedUser domain.FullInfoUser
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockUserClient.UserClient) {
				m.On("GetProfile", context.Background(), fixedUserLink).Return(expectedUser, nil)
			},
			expectedUser: expectedUser,
			expectError:  false,
		},
		{
			name: "UserNotFound",
			mockBehavior: func(m *mockUserClient.UserClient) {
				m.On("GetProfile", context.Background(), fixedUserLink).Return(domain.FullInfoUser{}, common.ErrorNonexistentUser)
			},
			expectedUser: domain.FullInfoUser{},
			expectError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockUserClient.NewUserClient(t)
			tc.mockBehavior(m)

			result, err := NewUser(m).GetProfile(context.Background(), fixedUserLink)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedUser, result)
			}
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	info := domain.UpdatedInfo{UserLink: fixedUserLink, DisplayName: "NewName", Description: "desc"}

	tests := []struct {
		name         string
		mockBehavior func(m *mockUserClient.UserClient)
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockUserClient.UserClient) {
				m.On("UpdateProfile", context.Background(), info).Return(nil)
			},
			expectError: false,
		},
		{
			name: "MissingField",
			mockBehavior: func(m *mockUserClient.UserClient) {
				m.On("UpdateProfile", context.Background(), info).Return(common.ErrorMissingRequiredField)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockUserClient.NewUserClient(t)
			tc.mockBehavior(m)

			err := NewUser(m).UpdateProfile(context.Background(), info)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpdateAvatar(t *testing.T) {
	info := domain.AvatarInfo{UserLink: fixedUserLink, FileData: []byte{0xFF, 0xD8}, ContentType: "image/jpeg"}

	tests := []struct {
		name         string
		mockBehavior func(m *mockUserClient.UserClient)
		expectedURL  string
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockUserClient.UserClient) {
				m.On("UpdateAvatar", context.Background(), info).Return("https://cdn.example.com/avatar.jpg", nil)
			},
			expectedURL: "https://cdn.example.com/avatar.jpg",
			expectError: false,
		},
		{
			name: "UserNotFound",
			mockBehavior: func(m *mockUserClient.UserClient) {
				m.On("UpdateAvatar", context.Background(), info).Return("", common.ErrorNonexistentUser)
			},
			expectedURL: "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockUserClient.NewUserClient(t)
			tc.mockBehavior(m)

			url, err := NewUser(m).UpdateAvatar(context.Background(), info)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedURL, url)
			}
		})
	}
}

func TestResetPassword(t *testing.T) {
	updatedPassword := domain.UpdatedPassword{UserLink: fixedUserLink, Password: "newpassword123"}

	tests := []struct {
		name         string
		mockBehavior func(m *mockUserClient.UserClient)
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockUserClient.UserClient) {
				m.On("ResetPassword", context.Background(), updatedPassword).Return(nil)
			},
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockUserClient.UserClient) {
				m.On("ResetPassword", context.Background(), updatedPassword).Return(testError)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockUserClient.NewUserClient(t)
			tc.mockBehavior(m)

			err := NewUser(m).ResetPassword(context.Background(), updatedPassword)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeleteAvatar(t *testing.T) {
	tests := []struct {
		name         string
		mockBehavior func(m *mockUserClient.UserClient)
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockUserClient.UserClient) {
				m.On("DeleteAvatar", context.Background(), fixedUserLink).Return(nil)
			},
			expectError: false,
		},
		{
			name: "UserNotFound",
			mockBehavior: func(m *mockUserClient.UserClient) {
				m.On("DeleteAvatar", context.Background(), fixedUserLink).Return(common.ErrorNonexistentUser)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockUserClient.NewUserClient(t)
			tc.mockBehavior(m)

			err := NewUser(m).DeleteAvatar(context.Background(), fixedUserLink)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetUserLink(t *testing.T) {
	email := "test@mail.ru"

	tests := []struct {
		name         string
		mockBehavior func(m *mockUserClient.UserClient)
		expectedLink uuid.UUID
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockUserClient.UserClient) {
				m.On("GetUserLink", context.Background(), email).Return(fixedUserLink, nil)
			},
			expectedLink: fixedUserLink,
			expectError:  false,
		},
		{
			name: "NotFound",
			mockBehavior: func(m *mockUserClient.UserClient) {
				m.On("GetUserLink", context.Background(), email).Return(uuid.Nil, common.ErrorNonexistentUser)
			},
			expectedLink: uuid.Nil,
			expectError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockUserClient.NewUserClient(t)
			tc.mockBehavior(m)

			link, err := NewUser(m).GetUserLink(context.Background(), email)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedLink, link)
			}
		})
	}
}
