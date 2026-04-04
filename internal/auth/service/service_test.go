package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service/dto"
	mockAuthRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service/mock_auth_rep"
	mockSender "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service/mock_sender"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	tests := []struct {
		nameTest          string
		displayName       string
		password          string
		email             string
		hasher            func(string) (string, error)
		checker           func(string, string) error
		generator         func() (string, error)
		mockBehavior      func(m *mockAuthRep.AuthRepository)
		expectedUser      dto.UserInfo
		expectedSessionID string
	}{
		{
			nameTest:    "Success registration",
			displayName: "Artem",
			password:    "1234567",
			email:       "test@mail.ru",
			hasher:      spyHasher,
			generator:   spyGenerator,
			checker:     spyChecker,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				m.On("AddUser", ctx, mock.AnythingOfType("dto.UserInitialize")).Return(nil)
				m.On("AddSession", ctx, mock.AnythingOfType("dto.SessionEntity")).Return(nil)
			},
			expectedUser: dto.UserInfo{
				Link:        common.FixedUserUuiD,
				DisplayName: "Artem",
				Email:       "test@mail.ru",
				Avatar:      "",
			},
			expectedSessionID: "sessionCLAC",
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			ctx := context.Background()

			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			serviceRegistration := NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil, CreaterResetKey, CreaterSessionKey)

			user, sessionID, err := serviceRegistration.Register(ctx, dto.RegistrationUser{
				DisplayName: test.displayName,
				Password:    test.password,
				Email:       test.email,
			})

			assert.NoError(t, err, "expected no error")
			assert.Equal(t, test.expectedSessionID, sessionID, "incorrect create sessionID")

			user.Link = common.FixedUserUuiD
			assert.Equal(t, test.expectedUser, user, "incorrect parse user")
		})
	}
}

func TestRegisterError(t *testing.T) {
	tests := []struct {
		nameTest     string
		displayName  string
		password     string
		email        string
		hasher       func(string) (string, error)
		generator    func() (string, error)
		checker      func(string, string) error
		mockBehavior func(m *mockAuthRep.AuthRepository)

		expectedError error
	}{
		{
			nameTest:    "Email is already existing",
			displayName: "Artem",
			password:    "1234567",
			email:       "test@mail.ru",
			hasher:      spyHasher,
			generator:   spyGenerator,
			checker:     spyChecker,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("AddUser", context.Background(), mock.AnythingOfType("dto.UserInitialize")).Return(common.ErrorExistingUser)
			},
			expectedError: fmt.Errorf("rep.AddUser: %w", common.ErrorExistingUser),
		},
		{
			nameTest:      "Error hash password",
			displayName:   "Artem",
			password:      "1234567",
			email:         "test@mail.ru",
			hasher:        spyHasherError,
			generator:     spyGenerator,
			checker:       spyChecker,
			mockBehavior:  nil,
			expectedError: fmt.Errorf("HashPassword: %w: %q", ErrorCreateHash, "error bcrypt"),
		},
		{
			nameTest:    "Error adding session",
			displayName: "Artem",
			password:    "1234567",
			email:       "test@mail.ru",
			hasher:      spyHasher,
			generator:   spyGenerator,
			checker:     spyChecker,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				m.On("AddUser", ctx, mock.AnythingOfType("dto.UserInitialize")).Return(nil)
				m.On("AddSession", ctx, mock.AnythingOfType("dto.SessionEntity")).Return(common.ErrorDetectingSessionCollision)
			},
			expectedError: fmt.Errorf("rep.AddSession: %w", common.ErrorDetectingSessionCollision),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)

			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			ctx := context.Background()
			serviceRegistration := NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil, CreaterResetKey, CreaterSessionKey)

			_, _, err := serviceRegistration.Register(ctx, dto.RegistrationUser{
				DisplayName: test.displayName,
				Password:    test.password,
				Email:       test.email,
			})

			assert.EqualError(t, err, test.expectedError.Error(), "incorrect error message")
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		id                uuid.UUID
		nameTest          string
		email             string
		password          string
		hasher            func(string) (string, error)
		checker           func(string, string) error
		generator         func() (string, error)
		mockBehavior      func(m *mockAuthRep.AuthRepository)
		expectedSessionID string
		expectedUser      dto.UserInfo
	}{
		{
			id:        common.FixedUserUuiD,
			nameTest:  "Success login",
			email:     "bobr@mail.ru",
			password:  "hash_12345",
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()

				userFromDB := repositoryDto.UserEntity{
					Link:         common.FixedUserUuiD,
					DisplayName:  "Artem",
					Email:        "bobr@mail.ru",
					PasswordHash: "hash_12345",
				}

				m.On("GetUser", ctx, "bobr@mail.ru").Return(userFromDB, nil)
				m.On("AddSession", ctx, mock.AnythingOfType("dto.SessionEntity")).Return(nil)
			},
			expectedUser: dto.UserInfo{
				Link:        common.FixedUserUuiD,
				DisplayName: "Artem",
				Email:       "bobr@mail.ru",
				Avatar:      "",
			},
			expectedSessionID: "sessionCLAC",
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			ctx := context.Background()

			serviceLogin := NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil, CreaterResetKey, CreaterSessionKey)

			user, sessionID, err := serviceLogin.LogIn(ctx, dto.LogInUser{
				Email:    test.email,
				Password: test.password,
			})

			assert.NoError(t, err, "expected no error")
			assert.Equal(t, test.expectedUser, user, "incorrect parsed user")
			assert.Equal(t, test.expectedSessionID, sessionID, "expected same sessionID")
		})
	}
}

func TestLoginError(t *testing.T) {
	tests := []struct {
		nameTest      string
		id            uuid.UUID
		email         string
		password      string
		checker       func(string, string) error
		hasher        func(string) (string, error)
		generator     func() (string, error)
		mockBehavior  func(m *mockAuthRep.AuthRepository)
		expectedError error
	}{
		{
			nameTest:  "Error user not found",
			id:        common.FixedUserUuiD,
			email:     "bobr@mail.ru",
			password:  "12345",
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("GetUser", context.Background(), "bobr@mail.ru").Return(repositoryDto.UserEntity{}, common.ErrorNonexistentUser)
			},
			expectedError: fmt.Errorf("rep.GetUser: %w", common.ErrorNonexistentUser),
		},
		{
			nameTest:  "Error wrong password",
			id:        common.FixedUserUuiD,
			email:     "test@mail.ru",
			password:  "wrong_password",
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("GetUser", context.Background(), "test@mail.ru").Return(repositoryDto.UserEntity{
					PasswordHash: "1234",
				}, nil)
			},
			expectedError: fmt.Errorf("rep.CheckPassword: %w", ErrorWrongPassword),
		},
		{
			nameTest:  "Error adding session to DB",
			id:        common.FixedUserUuiD,
			email:     "test@mail.ru",
			password:  "12345",
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				userFromDB := repositoryDto.UserEntity{
					Link:         uuid.New(),
					PasswordHash: "12345",
				}
				m.On("GetUser", ctx, "test@mail.ru").Return(userFromDB, nil)
				m.On("AddSession", ctx, mock.AnythingOfType("dto.SessionEntity")).Return(common.ErrorDetectingSessionCollision)
			},
			expectedError: fmt.Errorf("rep.AddSession: %w", common.ErrorDetectingSessionCollision),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			ctx := context.Background()

			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			serviceLogin := NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil, CreaterResetKey, CreaterSessionKey)

			_, _, err := serviceLogin.LogIn(ctx, dto.LogInUser{
				Email:    test.email,
				Password: test.password,
			})

			assert.EqualError(t, err, test.expectedError.Error(), "incorrect error message")
		})
	}
}

func TestCreateSessionForUser(t *testing.T) {
	userUUID := uuid.New()

	tests := []struct {
		nameTest     string
		userLink     uuid.UUID
		generatorID  func() (string, error)
		mockBehavior func(m *mockAuthRep.AuthRepository)
		expectedID   string
	}{
		{
			nameTest:    "Success create session",
			userLink:    userUUID,
			generatorID: spyGenerator,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				expectedSession := repositoryDto.SessionEntity{
					SessionKey: "sessionCLAC",
					UserLink:   userUUID,
					LifeTime:   SessionLifetime,
				}

				m.On("AddSession", ctx, expectedSession).Return(nil)
			},
			expectedID: "sessionCLAC",
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			ctx := context.Background()

			serviceAuth := NewService(mockRepo, nil, nil, nil, test.generatorID, nil, CreaterResetKey, CreaterSessionKey)

			sessionID, err := serviceAuth.CreateSessionForUser(ctx, test.userLink)

			assert.NoError(t, err, "expected no error")
			assert.Equal(t, test.expectedID, sessionID, "session IDs should match")
		})
	}
}

func TestCreateSessionForUserError(t *testing.T) {
	userUUID := uuid.New()

	tests := []struct {
		nameTest      string
		userLink      uuid.UUID
		generatorID   func() (string, error)
		mockBehavior  func(m *mockAuthRep.AuthRepository)
		expectedError error
	}{
		{
			nameTest: "Error generator ID fails",
			userLink: userUUID,
			generatorID: func() (string, error) {
				return "", errors.New("failed to generate id")
			},
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
			},
			expectedError: fmt.Errorf("GenerateID: %w", errors.New("failed to generate id")),
		},
		{
			nameTest: "Error AddSession to DB fails",
			userLink: userUUID,
			generatorID: func() (string, error) {
				return "sessionCLAC", nil
			},
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				expectedSession := repositoryDto.SessionEntity{
					SessionKey: "sessionCLAC",
					UserLink:   userUUID,
					LifeTime:   SessionLifetime,
				}

				m.On("AddSession", ctx, expectedSession).
					Return(errors.New("sessionCLAC"))
			},
			expectedError: fmt.Errorf("rep.AddSession: %w", errors.New("sessionCLAC")),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			ctx := context.Background()

			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			serviceAuth := NewService(mockRepo, nil, nil, nil, test.generatorID, nil, CreaterResetKey, CreaterSessionKey)

			sessionID, err := serviceAuth.CreateSessionForUser(ctx, test.userLink)

			assert.EqualError(t, err, test.expectedError.Error(), "incorrect error message")
			assert.Empty(t, sessionID, "sessionID should be empty on error")
		})
	}
}

func TestRefreshSession(t *testing.T) {
	t.Run("Success refresh session", func(t *testing.T) {
		mockRep := mockAuthRep.NewAuthRepository(t)
		mockRep.On("ExtendSession", mock.Anything, common.FixedSessionID, time.Hour*24).Return(nil)

		srv := NewService(mockRep, nil, nil, nil, nil, nil, nil, nil)

		err := srv.RefreshSession(context.Background(), common.FixedSessionID)

		assert.NoError(t, err, "not wait error")

		mockRep.AssertExpectations(t)
	})
}

func TestUpdateCountRequests(t *testing.T) {
	defaultConfig := dto.RateLimiterConfig{
		UserIP: common.FixedUserIP,
		Limit:  5,
		Action: "action",
		Window: 1 * time.Minute,
	}

	tests := []struct {
		nameTest      string
		config        dto.RateLimiterConfig
		mockAuthRep   func(m *mockAuthRep.AuthRepository)
		expectedValue bool
		expectedError error
	}{
		{
			nameTest: "First success request",
			config:   defaultConfig,
			mockAuthRep: func(m *mockAuthRep.AuthRepository) {
				m.On("CheckLimit", mock.Anything, repositoryDto.RateLimiterConfig{
					UserIP: defaultConfig.UserIP,
					Action: defaultConfig.Action,
					Window: defaultConfig.Window,
				}).Return(int64(1), nil)
			},
			expectedValue: false,
			expectedError: nil,
		},
		{
			nameTest: "Exceeded limit requests",
			config:   defaultConfig,
			mockAuthRep: func(m *mockAuthRep.AuthRepository) {
				m.On("CheckLimit", mock.Anything, repositoryDto.RateLimiterConfig{
					UserIP: defaultConfig.UserIP,
					Action: defaultConfig.Action,
					Window: defaultConfig.Window,
				}).Return(int64(6), nil)
			},
			expectedValue: true,
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockAuthRep != nil {
				test.mockAuthRep(mockRepo)
			}

			srv := NewService(mockRepo, nil, nil, nil, nil, nil, nil, nil)

			isFull, err := srv.UpdateCountRequests(context.Background(), test.config)

			assert.Equal(t, test.expectedValue, isFull, fmt.Sprintf("wait %t, get %t", test.expectedValue, test.expectedError))
			assert.NoError(t, err, "not wait error")
		})
	}
}

func TestUpdateCountRequestsError(t *testing.T) {
	errCheckLimitFailed := errors.New("fail checkLimit")

	defaultConfig := dto.RateLimiterConfig{
		UserIP: common.FixedUserIP,
		Limit:  5,
		Action: "action",
		Window: 1 * time.Minute,
	}
	tests := []struct {
		nameTest      string
		config        dto.RateLimiterConfig
		mockAuthRep   func(m *mockAuthRep.AuthRepository)
		expectedError error
	}{
		{
			nameTest: "Error check limit",
			config:   defaultConfig,
			mockAuthRep: func(m *mockAuthRep.AuthRepository) {
				m.On("CheckLimit", mock.Anything, repositoryDto.RateLimiterConfig{
					UserIP: defaultConfig.UserIP,
					Action: defaultConfig.Action,
					Window: defaultConfig.Window,
				}).Return(int64(0), errCheckLimitFailed)
			},
			expectedError: errCheckLimitFailed,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockAuthRep != nil {
				test.mockAuthRep(mockRepo)
			}

			srv := NewService(mockRepo, nil, nil, nil, nil, nil, nil, nil)

			_, err := srv.UpdateCountRequests(context.Background(), test.config)

			assert.ErrorIs(t, err, test.expectedError, fmt.Sprintf("wait %s", test.expectedError))
		})
	}
}

func TestCheckCoolDown(t *testing.T) {
	defaultConfig := dto.CoolDownConfig{
		Name:       "recovery_email",
		Email:      "test@mail.ru",
		Expiration: 1 * time.Minute,
	}

	expectedRepoConfig := repositoryDto.CoolDownConfig{
		Name:       defaultConfig.Name,
		Email:      defaultConfig.Email,
		Expiration: defaultConfig.Expiration,
	}

	tests := []struct {
		nameTest        string
		config          dto.CoolDownConfig
		mockBehavior    func(m *mockAuthRep.AuthRepository)
		expectedAllowed bool
		expectedTTL     time.Duration
	}{
		{
			nameTest: "Success cooldown allowed",
			config:   defaultConfig,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("SetCooldown", mock.Anything, expectedRepoConfig).
					Return(true, time.Duration(0), nil)
			},
			expectedAllowed: true,
			expectedTTL:     0,
		},
		{
			nameTest: "Success cooldown not allowed",
			config:   defaultConfig,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("SetCooldown", mock.Anything, expectedRepoConfig).
					Return(false, 30*time.Second, nil)
			},
			expectedAllowed: false,
			expectedTTL:     30 * time.Second,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			srv := NewService(mockRepo, nil, nil, nil, nil, nil, nil, nil)

			isAllowed, ttl, err := srv.CheckCoolDown(context.Background(), test.config)

			assert.NoError(t, err, "expected no error")
			assert.Equal(t, test.expectedAllowed, isAllowed, "incorrect isAllowed result")
			assert.Equal(t, test.expectedTTL, ttl, "incorrect TTL result")
		})
	}
}

func TestCheckCoolDownError(t *testing.T) {
	defaultConfig := dto.CoolDownConfig{
		Name:       "recovery_email",
		Email:      "test@mail.ru",
		Expiration: 1 * time.Minute,
	}

	expectedRepoConfig := repositoryDto.CoolDownConfig{
		Name:       defaultConfig.Name,
		Email:      defaultConfig.Email,
		Expiration: defaultConfig.Expiration,
	}

	errRepo := errors.New("repository error")

	tests := []struct {
		nameTest      string
		config        dto.CoolDownConfig
		mockBehavior  func(m *mockAuthRep.AuthRepository)
		expectedError error
	}{
		{
			nameTest: "Error from repository",
			config:   defaultConfig,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("SetCooldown", mock.Anything, expectedRepoConfig).
					Return(false, time.Duration(0), errRepo)
			},
			expectedError: errRepo,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			srv := NewService(mockRepo, nil, nil, nil, nil, nil, nil, nil)

			isAllowed, ttl, err := srv.CheckCoolDown(context.Background(), test.config)

			assert.ErrorIs(t, err, test.expectedError, "incorrect error type")
			assert.False(t, isAllowed, "isAllowed should be false on error")
			assert.Equal(t, time.Duration(0), ttl, "ttl should be 0 on error")
		})
	}
}

func TestRefreshSessionError(t *testing.T) {
	t.Run("Error refresh session", func(t *testing.T) {
		newErr := errors.New("error refresh")

		mockRep := mockAuthRep.NewAuthRepository(t)
		mockRep.On("ExtendSession", mock.Anything, common.FixedSessionID, time.Hour*24).Return(newErr)

		srv := NewService(mockRep, nil, nil, nil, nil, nil, nil, nil)

		err := srv.RefreshSession(context.Background(), common.FixedSessionID)

		assert.Error(t, err, "wait error")

		mockRep.AssertExpectations(t)
	})
}

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

			serviceLogOut := NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil, CreaterResetKey, CreaterSessionKey)

			err := serviceLogOut.LogOut(ctx, test.sessionID)
			assert.NoError(t, err, "not expected error")
		})
	}
}

func TestLogOutError(t *testing.T) {
	errInternalDB := errors.New("internal database failure")

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
			expectedError: common.ErrorNotExistingSession,
		},
		{
			nameTest:  "Error internal database failure",
			sessionID: common.FixedSessionID,
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				m.On("DeleteSession", ctx, common.FixedSessionID).Return(errInternalDB)
			},
			expectedError: errInternalDB,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			ctx := context.Background()

			serviceLogOut := NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil, CreaterResetKey, CreaterSessionKey)

			err := serviceLogOut.LogOut(ctx, test.sessionID)

			require.Error(t, err, "expected error to be returned")

			assert.ErrorIs(t, err, test.expectedError, "incorrect error returned")
		})
	}
}

func TestGetUserLink(t *testing.T) {
	expectedUUID := common.FixedUserUuiD

	tests := []struct {
		nameTest     string
		sessionID    string
		hasher       func(string) (string, error)
		checker      func(string, string) error
		generator    func() (string, error)
		mockBehavior func(m *mockAuthRep.AuthRepository)
		expectedID   uuid.UUID
	}{
		{
			nameTest:   "Success get user link",
			sessionID:  common.FixedSessionID,
			checker:    spyChecker,
			hasher:     spyHasher,
			generator:  spyGenerator,
			expectedID: expectedUUID,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				m.On("GetUserIDBySession", ctx, common.FixedSessionID).Return(expectedUUID.String(), nil)
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

			service := NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil, CreaterResetKey, CreaterSessionKey)

			userID, err := service.GetUserLink(ctx, test.sessionID)
			assert.NoError(t, err, "not expected error")
			assert.Equal(t, test.expectedID, userID, "incorrect user id")
		})
	}
}

func TestGetUserLinkError(t *testing.T) {
	tests := []struct {
		nameTest      string
		sessionID     string
		hasher        func(string) (string, error)
		checker       func(string, string) error
		generator     func() (string, error)
		mockBehavior  func(m *mockAuthRep.AuthRepository)
		expectedError error
		errorContains string
	}{
		{
			nameTest:  "Error session not found",
			sessionID: common.FixedSessionID,
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				m.On("GetUserIDBySession", ctx, common.FixedSessionID).Return("", common.ErrorNotExistingSession)
			},
			expectedError: common.ErrorNotExistingSession,
		},
		{
			nameTest:  "Error invalid UUID in DB",
			sessionID: common.FixedSessionID,
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				m.On("GetUserIDBySession", ctx, common.FixedSessionID).Return("not valid uuid", nil)
			},
			errorContains: "uuid.Parse",
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			ctx := context.Background()

			service := NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil, CreaterResetKey, CreaterSessionKey)

			userID, err := service.GetUserLink(ctx, test.sessionID)

			require.Error(t, err, "expected error")
			assert.Equal(t, uuid.Nil, userID, "expected nil uuid")

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError, "incorrect error type")
			}
		})
	}
}

func TestGetUserByEmail(t *testing.T) {
	expectedEmail := "test@example.com"
	expectedUser := repositoryDto.UserEntity{
		Link:  uuid.New(),
		Email: expectedEmail,
	}

	tests := []struct {
		nameTest     string
		email        string
		hasher       func(string) (string, error)
		checker      func(string, string) error
		generator    func() (string, error)
		mockBehavior func(m *mockAuthRep.AuthRepository)
		expectedUser dto.UserInfo
	}{
		{
			nameTest:  "Success get user by email",
			email:     expectedEmail,
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			expectedUser: dto.UserInfo{
				Link:  expectedUser.Link,
				Email: expectedUser.Email,
			},
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				m.On("GetUser", ctx, expectedEmail).Return(expectedUser, nil)
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

			service := NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil, nil, nil)

			user, err := service.GetUserByEmail(ctx, test.email)
			assert.NoError(t, err, "not expected error")
			assert.Equal(t, test.expectedUser, user, "incorrect user data")
		})
	}
}

func TestGetUserByEmailError(t *testing.T) {
	expectedEmail := "test@example.com"
	mockErr := errors.New("user not found")

	tests := []struct {
		nameTest      string
		email         string
		hasher        func(string) (string, error)
		checker       func(string, string) error
		generator     func() (string, error)
		mockBehavior  func(m *mockAuthRep.AuthRepository)
		expectedError error
	}{
		{
			nameTest:  "Error getting user",
			email:     expectedEmail,
			checker:   spyChecker,
			hasher:    spyHasher,
			generator: spyGenerator,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				m.On("GetUser", ctx, expectedEmail).Return(repositoryDto.UserEntity{}, mockErr)
			},
			expectedError: fmt.Errorf("rep.GetUser: %w", mockErr),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			ctx := context.Background()

			service := NewService(mockRepo, nil, test.hasher, test.checker, test.generator, nil, CreaterResetKey, CreaterSessionKey)

			user, err := service.GetUserByEmail(ctx, test.email)
			assert.Error(t, err, "expected error")
			assert.EqualError(t, err, test.expectedError.Error(), "incorrect error message")
			assert.Equal(t, dto.UserInfo{}, user, "expected empty user struct")
		})
	}
}

func TestSendRecoveryCode(t *testing.T) {
	targetEmail := "test@mail.ru"

	tests := []struct {
		nameTest      string
		email         string
		generator     func() (string, error)
		mockBehavior  func(m *mockAuthRep.AuthRepository)
		senderMock    func(m *mockSender.SenderLetters)
		expectedError error
	}{
		{
			nameTest:  "Success delivery code",
			email:     targetEmail,
			generator: func() (string, error) { return "123456", nil },
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("GetUserLink", mock.Anything, targetEmail).Return(common.FixedUserUuiD, nil)
				m.On("AddResetToken", mock.Anything, mock.Anything).Return(nil)
			},
			senderMock: func(m *mockSender.SenderLetters) {
				m.On("SendLetter", targetEmail, "Code to create a new password", mock.AnythingOfType("string")).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest:  "Error user not found",
			email:     "testing@mail.ru",
			generator: func() (string, error) { return "123456", nil },
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				m.On("GetUserLink", mock.Anything, "testing@mail.ru").Return(uuid.Nil, common.ErrorNonexistentUser)
			},
			senderMock:    func(m *mockSender.SenderLetters) {},
			expectedError: fmt.Errorf("rep.GetUser: %w", common.ErrorNonexistentUser),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			mockMail := mockSender.NewSenderLetters(t)

			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}
			if test.senderMock != nil {
				test.senderMock(mockMail)
			}

			service := NewService(mockRepo, mockMail, nil, nil, test.generator, test.generator, CreaterResetKey, CreaterSessionKey)

			err := service.SendRecoveryCode(context.Background(), test.email)

			time.Sleep(10 * time.Millisecond)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

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

				m.On("GetUserLinkByResetToken", mock.Anything, validToken).Return(common.FixedUserUuiD.String(), nil)
			},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := mockAuthRep.NewAuthRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepo)
			}

			service := NewService(mockRepo, nil, nil, nil, nil, nil, nil, nil)
			err := service.CheckRecoveryCode(context.Background(), test.tokenID)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

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

				m.On("GetUserLinkByResetToken", ctx, common.FixedResetTokenID).
					Return(common.FixedUserUuiD.String(), nil)

				m.On("UpdatePassword", ctx, common.FixedUserUuiD, "hash_new_password").
					Return(nil)
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

			serviceAuth := NewService(mockRepo, nil, test.hasher, nil, nil, nil, nil, nil)

			err := serviceAuth.ResetPassword(ctx, test.tokenID, test.newPassword)

			assert.NoError(t, err, "expected no error")
		})
	}
}

func TestResetPasswordError(t *testing.T) {
	targetUserID := uuid.New()

	_, invalidUUIDErr := uuid.Parse("invalid-uuid-string")

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
				m.On("GetUserLinkByResetToken", ctx, common.FixedResetTokenID).
					Return("", common.ErrorResetTokenExpired)
			},
			expectedError: fmt.Errorf("rep.GetResetToken: %w", common.ErrorResetTokenExpired),
		},
		{
			nameTest:    "Error invalid UUID format from DB",
			tokenID:     common.FixedResetTokenID,
			newPassword: "new_password",
			hasher:      spyHasher,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				m.On("GetUserLinkByResetToken", ctx, common.FixedResetTokenID).
					Return("invalid-uuid-string", nil)
			},
			expectedError: fmt.Errorf("uuid.Parse: %w", invalidUUIDErr),
		},
		{
			nameTest:    "Error hasher fails",
			tokenID:     common.FixedResetTokenID,
			newPassword: "new_password",
			hasher:      spyHasherError,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				m.On("GetUserLinkByResetToken", ctx, common.FixedResetTokenID).
					Return(targetUserID.String(), nil)
			},
			expectedError: fmt.Errorf("hasher: %w", errors.New("failed to create hash: \"error bcrypt\"")),
		},
		{
			nameTest:    "Error update password in DB",
			tokenID:     common.FixedResetTokenID,
			newPassword: "new_password",
			hasher:      spyHasher,
			mockBehavior: func(m *mockAuthRep.AuthRepository) {
				ctx := context.Background()
				m.On("GetUserLinkByResetToken", ctx, common.FixedResetTokenID).
					Return(targetUserID.String(), nil)

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

			serviceAuth := NewService(mockRepo, nil, test.hasher, nil, nil, nil, nil, nil)

			err := serviceAuth.ResetPassword(ctx, test.tokenID, test.newPassword)

			assert.EqualError(t, err, test.expectedError.Error(), "incorrect error message")
		})
	}
}

func TestEnsureUserByEmail(t *testing.T) {
	testUserInfo := dto.RegistrationUser{
		Email: "info@mail.ru",
	}

	tests := []struct {
		Name         string
		ExpectError  bool
		MockBehavior func(r *mockAuthRep.AuthRepository)
	}{
		{
			Name:        "no error, user exists",
			ExpectError: false,
			MockBehavior: func(r *mockAuthRep.AuthRepository) {
				r.On("GetUser", mock.Anything, testUserInfo.Email).
					Return(repositoryDto.UserEntity{Email: testUserInfo.Email}, nil)
			},
		},
		{
			Name:        "no error, user does not exists",
			ExpectError: false,
			MockBehavior: func(r *mockAuthRep.AuthRepository) {
				r.On("GetUser", mock.Anything, testUserInfo.Email).
					Return(repositoryDto.UserEntity{}, common.ErrorNonexistentUser)

				r.On("AddUser", mock.Anything, mock.Anything).
					Return(nil)

				r.On("AddSession", mock.Anything, mock.Anything).
					Return(nil)
			},
		},
		{
			Name:        "cannot find user error",
			ExpectError: true,
			MockBehavior: func(r *mockAuthRep.AuthRepository) {
				r.On("GetUser", mock.Anything, testUserInfo.Email).
					Return(repositoryDto.UserEntity{}, errors.New("cannot find user"))
			},
		},
		{
			Name:        "cannot register error",
			ExpectError: true,
			MockBehavior: func(r *mockAuthRep.AuthRepository) {
				r.On("GetUser", mock.Anything, testUserInfo.Email).
					Return(repositoryDto.UserEntity{}, common.ErrorNonexistentUser)

				r.On("AddUser", mock.Anything, mock.Anything).
					Return(errors.New("cannot register user"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			authRepo := new(mockAuthRep.AuthRepository)
			if test.MockBehavior != nil {
				test.MockBehavior(authRepo)
			}

			service := NewService(authRepo, nil, spyHasher, nil, spyGenerator, nil, nil, nil)
			user, err := service.EnsureUserByEmail(context.Background(), testUserInfo)
			if test.ExpectError {
				require.Error(t, err, "must return error")
				return
			}

			assert.Equal(t, testUserInfo.Email, user.Email, "users must be equal")

			authRepo.AssertExpectations(t)
		})
	}
}
