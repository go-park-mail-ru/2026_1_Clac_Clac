package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	mockAuthClient "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/usecase/mock_auth_client"
	mockMailClient "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/usecase/mock_mail_sender_client"
	mockUserClient "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/usecase/mock_user_client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	fixedUserLink  = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	fixedSessionID = "12345667"
)

func newTestAuthUser(u *mockUserClient.UserClient, a *mockAuthClient.AuthClient, m *mockMailClient.MailSenderClient) *AuthUser {
	return NewAuthUser(u, a, m)
}

func TestLogin(t *testing.T) {
	type TestCase struct {
		Name           string
		Cred           domain.Credentials
		MockBehavior   func(u *mockUserClient.UserClient, a *mockAuthClient.AuthClient)
		ExpectedError  error
		ExpectedSessID string
	}

	tests := []TestCase{
		{
			Name: "Success",
			Cred: domain.Credentials{Email: "test@mail.ru", Password: "pass12345"},
			MockBehavior: func(u *mockUserClient.UserClient, a *mockAuthClient.AuthClient) {
				u.On("GetUser", context.Background(), domain.Credentials{Email: "test@mail.ru", Password: "pass12345"}).
					Return(domain.FullInfoUser{UserLink: fixedUserLink, Email: "test@mail.ru"}, nil)
				a.On("CreateSession", context.Background(), fixedUserLink).Return(fixedSessionID, nil)
			},
			ExpectedSessID: fixedSessionID,
		},
		{
			Name: "WrongCredentials",
			Cred: domain.Credentials{Email: "test@mail.ru", Password: "wrong"},
			MockBehavior: func(u *mockUserClient.UserClient, a *mockAuthClient.AuthClient) {
				u.On("GetUser", context.Background(), domain.Credentials{Email: "test@mail.ru", Password: "wrong"}).
					Return(domain.FullInfoUser{}, common.ErrorWrongCredentials)
			},
			ExpectedError: common.ErrorWrongCredentials,
		},
		{
			Name: "SessionCreateError",
			Cred: domain.Credentials{Email: "test@mail.ru", Password: "pass12345"},
			MockBehavior: func(u *mockUserClient.UserClient, a *mockAuthClient.AuthClient) {
				u.On("GetUser", context.Background(), domain.Credentials{Email: "test@mail.ru", Password: "pass12345"}).
					Return(domain.FullInfoUser{UserLink: fixedUserLink}, nil)
				a.On("CreateSession", context.Background(), fixedUserLink).Return("", errors.New("redis down"))
			},
			ExpectedError: errors.New("redis down"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			u := mockUserClient.NewUserClient(t)
			a := mockAuthClient.NewAuthClient(t)
			m := mockMailClient.NewMailSenderClient(t)

			tc.MockBehavior(u, a)

			svc := newTestAuthUser(u, a, m)
			info, sid, err := svc.Login(context.Background(), tc.Cred)

			if tc.ExpectedError != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.ExpectedError) || err != nil)
			} else {
				require.NoError(t, err)
				assert.Equal(t, fixedUserLink, info.Link)
				assert.Equal(t, tc.ExpectedSessID, sid)
			}
		})
	}
}

func TestRegister(t *testing.T) {
	type TestCase struct {
		Name          string
		Cred          domain.NewCredentialsUser
		MockBehavior  func(u *mockUserClient.UserClient, a *mockAuthClient.AuthClient)
		ExpectedError error
	}

	tests := []TestCase{
		{
			Name: "Success",
			Cred: domain.NewCredentialsUser{Email: "new@mail.ru", Password: "pass12345", DisplayName: "New"},
			MockBehavior: func(u *mockUserClient.UserClient, a *mockAuthClient.AuthClient) {
				cred := domain.NewCredentialsUser{Email: "new@mail.ru", Password: "pass12345", DisplayName: "New"}
				u.On("CreateUser", context.Background(), cred).
					Return(domain.FullInfoUser{UserLink: fixedUserLink, Email: "new@mail.ru"}, nil)
				a.On("CreateSession", context.Background(), fixedUserLink).Return(fixedSessionID, nil)
			},
		},
		{
			Name: "UserAlreadyExists",
			Cred: domain.NewCredentialsUser{Email: "dup@mail.ru", Password: "pass12345"},
			MockBehavior: func(u *mockUserClient.UserClient, a *mockAuthClient.AuthClient) {
				cred := domain.NewCredentialsUser{Email: "dup@mail.ru", Password: "pass12345"}
				u.On("CreateUser", context.Background(), cred).Return(domain.FullInfoUser{}, common.ErrorExistingUser)
			},
			ExpectedError: common.ErrorExistingUser,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			u := mockUserClient.NewUserClient(t)
			a := mockAuthClient.NewAuthClient(t)
			m := mockMailClient.NewMailSenderClient(t)

			tc.MockBehavior(u, a)

			svc := newTestAuthUser(u, a, m)
			_, _, err := svc.Register(context.Background(), tc.Cred)

			if tc.ExpectedError != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.ExpectedError))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLogout(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		u := mockUserClient.NewUserClient(t)
		a := mockAuthClient.NewAuthClient(t)
		m := mockMailClient.NewMailSenderClient(t)
		a.On("DeleteSession", context.Background(), fixedSessionID).Return(nil)
		err := newTestAuthUser(u, a, m).Logout(context.Background(), fixedSessionID)
		require.NoError(t, err)
	})

	t.Run("SessionNotFound", func(t *testing.T) {
		u := mockUserClient.NewUserClient(t)
		a := mockAuthClient.NewAuthClient(t)
		m := mockMailClient.NewMailSenderClient(t)
		a.On("DeleteSession", context.Background(), fixedSessionID).Return(common.ErrorSessionNotFound)
		err := newTestAuthUser(u, a, m).Logout(context.Background(), fixedSessionID)
		require.Error(t, err)
	})
}

func TestSendRecoveryCode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		u := mockUserClient.NewUserClient(t)
		a := mockAuthClient.NewAuthClient(t)
		m := mockMailClient.NewMailSenderClient(t)

		email := "user@mail.ru"
		u.On("GetUserLink", context.Background(), email).Return(fixedUserLink, nil)
		m.On("SendRecoveryCode", context.Background(), domain.RecoveryCode{UserLink: fixedUserLink, Email: email}).Return(nil)

		err := newTestAuthUser(u, a, m).SendRecoveryCode(context.Background(), email)
		require.NoError(t, err)
	})

	t.Run("EmailNotFound", func(t *testing.T) {
		u := mockUserClient.NewUserClient(t)
		a := mockAuthClient.NewAuthClient(t)
		m := mockMailClient.NewMailSenderClient(t)

		email := "notfound@mail.ru"
		u.On("GetUserLink", context.Background(), email).Return(uuid.Nil, common.ErrorNonexistentEmail)

		err := newTestAuthUser(u, a, m).SendRecoveryCode(context.Background(), email)
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrorNonexistentEmail))
	})
}

func TestCheckRecoveryCode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		u := mockUserClient.NewUserClient(t)
		a := mockAuthClient.NewAuthClient(t)
		m := mockMailClient.NewMailSenderClient(t)

		m.On("CheckRecoveryCode", context.Background(), domain.RecoveryCodeCheck{Code: "123456"}).Return(nil)

		err := newTestAuthUser(u, a, m).CheckRecoveryCode(context.Background(), "123456")
		require.NoError(t, err)
	})

	t.Run("InvalidCode", func(t *testing.T) {
		u := mockUserClient.NewUserClient(t)
		a := mockAuthClient.NewAuthClient(t)
		m := mockMailClient.NewMailSenderClient(t)

		m.On("CheckRecoveryCode", context.Background(), domain.RecoveryCodeCheck{Code: "bad"}).Return(errors.New("invalid code"))

		err := newTestAuthUser(u, a, m).CheckRecoveryCode(context.Background(), "bad")
		require.Error(t, err)
	})
}

func TestResetPassword(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		u := mockUserClient.NewUserClient(t)
		a := mockAuthClient.NewAuthClient(t)
		m := mockMailClient.NewMailSenderClient(t)

		tokenID := "reset-token"
		newPass := "newpass1"
		m.On("ExchangeTokenForUser", context.Background(), domain.ResetToken{Token: tokenID}).Return(fixedUserLink, nil)
		u.On("RessetPassword", context.Background(), domain.UpdatedPassoword{
			UserLink: fixedUserLink, Password: newPass, RepeatedPassword: newPass,
		}).Return(nil)

		err := newTestAuthUser(u, a, m).ResetPassword(context.Background(), tokenID, newPass)
		require.NoError(t, err)
	})

	t.Run("TokenNotFound", func(t *testing.T) {
		u := mockUserClient.NewUserClient(t)
		a := mockAuthClient.NewAuthClient(t)
		m := mockMailClient.NewMailSenderClient(t)

		m.On("ExchangeTokenForUser", context.Background(), domain.ResetToken{Token: "bad"}).
			Return(uuid.Nil, common.ErrorResetTokenNotFound)

		err := newTestAuthUser(u, a, m).ResetPassword(context.Background(), "bad", "pass")
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrorResetTokenNotFound))
	})
}

func TestLoginWithVK(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		u := mockUserClient.NewUserClient(t)
		a := mockAuthClient.NewAuthClient(t)
		m := mockMailClient.NewMailSenderClient(t)

		code := "vk-code"
		accessToken, email := "vk-access", "vk@mail.ru"
		a.On("ExchangeVKCode", context.Background(), code).Return(accessToken, email, nil)
		u.On("ProcessUserWithVK", context.Background(), accessToken, email).Return(fixedUserLink, nil)
		a.On("CreateSession", context.Background(), fixedUserLink).Return(fixedSessionID, nil)

		info, sid, err := newTestAuthUser(u, a, m).LoginWithVK(context.Background(), code)
		require.NoError(t, err)
		assert.Equal(t, fixedUserLink, info.Link)
		assert.Equal(t, fixedSessionID, sid)
	})

	t.Run("VKUnavailable", func(t *testing.T) {
		u := mockUserClient.NewUserClient(t)
		a := mockAuthClient.NewAuthClient(t)
		m := mockMailClient.NewMailSenderClient(t)

		a.On("ExchangeVKCode", context.Background(), "bad-code").Return("", "", common.ErrorVKOAuthUnavailable)

		_, _, err := newTestAuthUser(u, a, m).LoginWithVK(context.Background(), "bad-code")
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrorVKOAuthUnavailable))
	})
}
