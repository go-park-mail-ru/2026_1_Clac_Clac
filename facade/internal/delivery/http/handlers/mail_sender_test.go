package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Inline Mocks ---

type mockMailSenderUC struct{ mock.Mock }

func (m *mockMailSenderUC) SendRecoveryCode(ctx context.Context, userLink uuid.UUID, email string) error {
	return m.Called(ctx, userLink, email).Error(0)
}

func (m *mockMailSenderUC) CheckRecoveryCode(ctx context.Context, tokenID string) error {
	return m.Called(ctx, tokenID).Error(0)
}

func (m *mockMailSenderUC) ExchangeTokenForUser(ctx context.Context, resetToken domain.ResetToken) (uuid.UUID, error) {
	args := m.Called(ctx, resetToken)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

type mockCoolDownUC struct{ mock.Mock }

func (m *mockCoolDownUC) CheckCoolDown(ctx context.Context, cooldown domain.Cooldown) (domain.CooldownResult, error) {
	args := m.Called(ctx, cooldown)
	return args.Get(0).(domain.CooldownResult), args.Error(1)
}

type mockGeterUserLink struct{ mock.Mock }

func (m *mockGeterUserLink) GetUserLink(ctx context.Context, email string) (uuid.UUID, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

// --- Tests ---

func TestSendRecoveryEmail(t *testing.T) {
	fixedLink := uuid.New()
	validEmail := "test@mail.ru"
	cfg := MailSenderConfig{CoolDownExpirationSec: 60}
	cooldownArgs := domain.Cooldown{
		Name:        nameCoolDown,
		Email:       validEmail,
		ExpirationS: cfg.CoolDownExpirationSec,
	}

	tests := []struct {
		name               string
		requestBody        any
		mockBehavior       func(ms *mockMailSenderUC, cd *mockCoolDownUC, gul *mockGeterUserLink)
		expectedStatusCode int
	}{
		{
			name:        "Success",
			requestBody: dto.PasswordRecoveryRequest{Email: validEmail},
			mockBehavior: func(ms *mockMailSenderUC, cd *mockCoolDownUC, gul *mockGeterUserLink) {
				cd.On("CheckCoolDown", mock.Anything, cooldownArgs).Return(domain.CooldownResult{Allowed: true}, nil)
				gul.On("GetUserLink", mock.Anything, validEmail).Return(fixedLink, nil)
				ms.On("SendRecoveryCode", mock.Anything, fixedLink, validEmail).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "InvalidJSON",
			requestBody:        "{bad_json}",
			mockBehavior:       func(ms *mockMailSenderUC, cd *mockCoolDownUC, gul *mockGeterUserLink) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "InvalidEmail",
			requestBody:        dto.PasswordRecoveryRequest{Email: "invalid-email"},
			mockBehavior:       func(ms *mockMailSenderUC, cd *mockCoolDownUC, gul *mockGeterUserLink) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:        "CoolDownError",
			requestBody: dto.PasswordRecoveryRequest{Email: validEmail},
			mockBehavior: func(ms *mockMailSenderUC, cd *mockCoolDownUC, gul *mockGeterUserLink) {
				cd.On("CheckCoolDown", mock.Anything, cooldownArgs).Return(domain.CooldownResult{}, errors.New("redis down"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:        "TooManyRequests",
			requestBody: dto.PasswordRecoveryRequest{Email: validEmail},
			mockBehavior: func(ms *mockMailSenderUC, cd *mockCoolDownUC, gul *mockGeterUserLink) {
				cd.On("CheckCoolDown", mock.Anything, cooldownArgs).Return(domain.CooldownResult{Allowed: false, WaitS: 30}, nil)
			},
			expectedStatusCode: http.StatusTooManyRequests,
		},
		{
			name:        "UserNotFound",
			requestBody: dto.PasswordRecoveryRequest{Email: validEmail},
			mockBehavior: func(ms *mockMailSenderUC, cd *mockCoolDownUC, gul *mockGeterUserLink) {
				cd.On("CheckCoolDown", mock.Anything, cooldownArgs).Return(domain.CooldownResult{Allowed: true}, nil)
				gul.On("GetUserLink", mock.Anything, validEmail).Return(uuid.Nil, common.ErrorNonexistentUser)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:        "GetUserLinkInternalError",
			requestBody: dto.PasswordRecoveryRequest{Email: validEmail},
			mockBehavior: func(ms *mockMailSenderUC, cd *mockCoolDownUC, gul *mockGeterUserLink) {
				cd.On("CheckCoolDown", mock.Anything, cooldownArgs).Return(domain.CooldownResult{Allowed: true}, nil)
				gul.On("GetUserLink", mock.Anything, validEmail).Return(uuid.Nil, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:        "SendRecoveryCodeInternalError",
			requestBody: dto.PasswordRecoveryRequest{Email: validEmail},
			mockBehavior: func(ms *mockMailSenderUC, cd *mockCoolDownUC, gul *mockGeterUserLink) {
				cd.On("CheckCoolDown", mock.Anything, cooldownArgs).Return(domain.CooldownResult{Allowed: true}, nil)
				gul.On("GetUserLink", mock.Anything, validEmail).Return(fixedLink, nil)
				ms.On("SendRecoveryCode", mock.Anything, fixedLink, validEmail).Return(errors.New("smtp error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := new(mockMailSenderUC)
			cd := new(mockCoolDownUC)
			gul := new(mockGeterUserLink)
			tc.mockBehavior(ms, cd, gul)

			handler := NewMailSender(ms, cd, gul, cfg)

			var bodyReader *bytes.Reader
			if strBody, ok := tc.requestBody.(string); ok {
				bodyReader = bytes.NewReader([]byte(strBody))
			} else {
				b, _ := json.Marshal(tc.requestBody)
				bodyReader = bytes.NewReader(b)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/forgot-password", bodyReader)
			rr := httptest.NewRecorder()

			handler.SendRecoveryEmail(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestCheckRecoveryCode(t *testing.T) {
	validCode := "123456"

	tests := []struct {
		name               string
		requestBody        any
		mockBehavior       func(ms *mockMailSenderUC)
		expectedStatusCode int
	}{
		{
			name:        "Success",
			requestBody: dto.RecoveryCodeRequest{Code: validCode},
			mockBehavior: func(ms *mockMailSenderUC) {
				ms.On("CheckRecoveryCode", mock.Anything, validCode).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "InvalidJSON",
			requestBody:        "{bad_json}",
			mockBehavior:       func(ms *mockMailSenderUC) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:        "TokenNotFound",
			requestBody: dto.RecoveryCodeRequest{Code: validCode},
			mockBehavior: func(ms *mockMailSenderUC) {
				ms.On("CheckRecoveryCode", mock.Anything, validCode).Return(common.ErrorResetTokenNotFound)
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:        "InternalServerError",
			requestBody: dto.RecoveryCodeRequest{Code: validCode},
			mockBehavior: func(ms *mockMailSenderUC) {
				ms.On("CheckRecoveryCode", mock.Anything, validCode).Return(errors.New("redis down"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := new(mockMailSenderUC)
			tc.mockBehavior(ms)

			handler := NewMailSender(ms, nil, nil, MailSenderConfig{})

			var bodyReader *bytes.Reader
			if strBody, ok := tc.requestBody.(string); ok {
				bodyReader = bytes.NewReader([]byte(strBody))
			} else {
				b, _ := json.Marshal(tc.requestBody)
				bodyReader = bytes.NewReader(b)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/check-code", bodyReader)
			rr := httptest.NewRecorder()

			handler.CheckRecoveryCode(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}
