package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/dto"
	mockProfileUC "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/mock_profile_use_case"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var defaultProfileCfg = ProfileConfig{
	ValidExtensions: map[string]struct{}{
		"image/jpeg": {},
		"image/png":  {},
		"image/webp": {},
	},
	SignatureTypeBytes:    512,
	MaxLenNameUser:        128,
	MaxLenDescriptionUser: 500,
	MaxReadBytes:          5 << 20,
	MaxLenPassword:        128,
	MinLenPassword:        8,
}

var jpegMagic = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}

func newTestProfileHandler(profile ProfileUseCase, mail MailSenderUsecase) *Profile {
	return NewProfileHandler(profile, mail, defaultProfileCfg)
}

func TestGetProfile(t *testing.T) {
	user := domain.FullInfoUser{UserLink: fixedLink, Email: "t@mail.ru", DisplayName: "Test"}

	tests := []struct {
		name               string
		setContext         bool
		mockBehavior       func(m *mockProfileUC.ProfileUseCase)
		expectedStatusCode int
		expectedContains   string
	}{
		{
			name:       "Success",
			setContext: true,
			mockBehavior: func(m *mockProfileUC.ProfileUseCase) {
				m.On("GetProfile", mock.Anything, fixedLink).Return(user, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedContains:   "t@mail.ru",
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			mockBehavior:       func(m *mockProfileUC.ProfileUseCase) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedContains:   msgUnauthorized,
		},
		{
			name:       "UserNotFound",
			setContext: true,
			mockBehavior: func(m *mockProfileUC.ProfileUseCase) {
				m.On("GetProfile", mock.Anything, fixedLink).Return(domain.FullInfoUser{}, common.ErrorNonexistentUser)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedContains:   msgUserNotFound,
		},
		{
			name:       "InternalError",
			setContext: true,
			mockBehavior: func(m *mockProfileUC.ProfileUseCase) {
				m.On("GetProfile", mock.Anything, fixedLink).Return(domain.FullInfoUser{}, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedContains:   msgFailGetProfile,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockProfileUC.NewProfileUseCase(t)
			tc.mockBehavior(m)

			req := httptest.NewRequest(http.MethodGet, "/profiles", nil)
			if tc.setContext {
				ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
				req = req.WithContext(ctx)
			}
			rr := httptest.NewRecorder()

			newTestProfileHandler(m, nil).GetProfile(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.expectedContains)
		})
	}
}

func TestGetProfileByLink(t *testing.T) {
	user := domain.FullInfoUser{UserLink: fixedLink, Email: "t@mail.ru"}

	tests := []struct {
		name               string
		linkParam          string
		mockBehavior       func(m *mockProfileUC.ProfileUseCase)
		expectedStatusCode int
	}{
		{
			name:      "Success",
			linkParam: fixedLink.String(),
			mockBehavior: func(m *mockProfileUC.ProfileUseCase) {
				m.On("GetProfile", mock.Anything, fixedLink).Return(user, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "InvalidUUID",
			linkParam:          "not-a-uuid",
			mockBehavior:       func(m *mockProfileUC.ProfileUseCase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:      "NotFound",
			linkParam: fixedLink.String(),
			mockBehavior: func(m *mockProfileUC.ProfileUseCase) {
				m.On("GetProfile", mock.Anything, fixedLink).Return(domain.FullInfoUser{}, common.ErrorNonexistentUser)
			},
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockProfileUC.NewProfileUseCase(t)
			tc.mockBehavior(m)

			req := httptest.NewRequest(http.MethodGet, "/profiles/"+tc.linkParam, nil)
			req = mux.SetURLVars(req, map[string]string{"user_link": tc.linkParam})
			rr := httptest.NewRecorder()

			newTestProfileHandler(m, nil).GetProfileByLink(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	tests := []struct {
		name               string
		request            any
		setContext         bool
		mockBehavior       func(m *mockProfileUC.ProfileUseCase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			request:    dto.UpdateProfileRequest{DisplayName: "NewName", Description: "desc"},
			setContext: true,
			mockBehavior: func(m *mockProfileUC.ProfileUseCase) {
				m.On("UpdateProfile", mock.Anything, domain.UpdatedInfo{
					UserLink:    fixedLink,
					DisplayName: "NewName",
					Description: "desc",
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			request:            dto.UpdateProfileRequest{DisplayName: "Name"},
			setContext:         false,
			mockBehavior:       func(m *mockProfileUC.ProfileUseCase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidJSON",
			request:            "{bad}",
			setContext:         true,
			mockBehavior:       func(m *mockProfileUC.ProfileUseCase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "NameTooLong",
			request:            dto.UpdateProfileRequest{DisplayName: strings.Repeat("a", 200)},
			setContext:         true,
			mockBehavior:       func(m *mockProfileUC.ProfileUseCase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "DescriptionTooLong",
			request:            dto.UpdateProfileRequest{DisplayName: "ok", Description: strings.Repeat("d", 600)},
			setContext:         true,
			mockBehavior:       func(m *mockProfileUC.ProfileUseCase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "MissingRequiredField",
			request:    dto.UpdateProfileRequest{DisplayName: "ok", Description: "desc"},
			setContext: true,
			mockBehavior: func(m *mockProfileUC.ProfileUseCase) {
				m.On("UpdateProfile", mock.Anything, mock.Anything).Return(common.ErrorMissingRequiredField)
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "InvalidProfileData",
			request:    dto.UpdateProfileRequest{DisplayName: "ok", Description: "desc"},
			setContext: true,
			mockBehavior: func(m *mockProfileUC.ProfileUseCase) {
				m.On("UpdateProfile", mock.Anything, mock.Anything).Return(common.ErrorInvalidProfileData)
			},
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockProfileUC.NewProfileUseCase(t)
			tc.mockBehavior(m)

			var bodyReader *bytes.Reader
			if strBody, ok := tc.request.(string); ok {
				bodyReader = bytes.NewReader([]byte(strBody))
			} else {
				b, _ := json.Marshal(tc.request)
				bodyReader = bytes.NewReader(b)
			}

			req := httptest.NewRequest(http.MethodPost, "/profiles/info", bodyReader)
			if tc.setContext {
				ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
				req = req.WithContext(ctx)
			}
			rr := httptest.NewRecorder()

			newTestProfileHandler(m, nil).UpdateProfile(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func buildAvatarRequest(t *testing.T, fileData []byte, withContext bool, isMultipart bool) *http.Request {
	t.Helper()
	if !isMultipart {
		req := httptest.NewRequest(http.MethodPut, "/profiles/avatar", strings.NewReader("plain text body"))
		if withContext {
			ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
			req = req.WithContext(ctx)
		}
		return req
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("avatar", "avatar.jpg")
	require.NoError(t, err)
	_, err = part.Write(fileData)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req := httptest.NewRequest(http.MethodPut, "/profiles/avatar", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if withContext {
		ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
		req = req.WithContext(ctx)
	}
	return req
}

func TestUpdateAvatar(t *testing.T) {
	tests := []struct {
		name               string
		fileData           []byte
		withContext        bool
		isMultipart        bool
		mockBehavior       func(m *mockProfileUC.ProfileUseCase)
		expectedStatusCode int
	}{
		{
			name:        "Success",
			fileData:    jpegMagic,
			withContext: true,
			isMultipart: true,
			mockBehavior: func(m *mockProfileUC.ProfileUseCase) {
				m.On("UpdateAvatar", mock.Anything, mock.MatchedBy(func(info domain.AvatarInfo) bool {
					return info.UserLink == fixedLink && info.ContentType == "image/jpeg"
				})).Return("https://cdn.example.com/avatar.jpg", nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "NoMultipart",
			fileData:           nil,
			withContext:        true,
			isMultipart:        false,
			mockBehavior:       func(m *mockProfileUC.ProfileUseCase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "InvalidMimeType",
			fileData:           []byte("not-an-image"),
			withContext:        true,
			isMultipart:        true,
			mockBehavior:       func(m *mockProfileUC.ProfileUseCase) {},
			expectedStatusCode: http.StatusUnsupportedMediaType,
		},
		{
			name:               "Unauthorized",
			fileData:           jpegMagic,
			withContext:        false,
			isMultipart:        true,
			mockBehavior:       func(m *mockProfileUC.ProfileUseCase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:        "UserNotFound",
			fileData:    jpegMagic,
			withContext: true,
			isMultipart: true,
			mockBehavior: func(m *mockProfileUC.ProfileUseCase) {
				m.On("UpdateAvatar", mock.Anything, mock.Anything).Return("", common.ErrorNonexistentUser)
			},
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockProfileUC.NewProfileUseCase(t)
			tc.mockBehavior(m)

			req := buildAvatarRequest(t, tc.fileData, tc.withContext, tc.isMultipart)
			rr := httptest.NewRecorder()

			newTestProfileHandler(m, nil).UpdateAvatar(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestDeleteAvatar(t *testing.T) {
	tests := []struct {
		name               string
		setContext         bool
		mockBehavior       func(m *mockProfileUC.ProfileUseCase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			setContext: true,
			mockBehavior: func(m *mockProfileUC.ProfileUseCase) {
				m.On("DeleteAvatar", mock.Anything, fixedLink).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			mockBehavior:       func(m *mockProfileUC.ProfileUseCase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:       "UserNotFound",
			setContext: true,
			mockBehavior: func(m *mockProfileUC.ProfileUseCase) {
				m.On("DeleteAvatar", mock.Anything, fixedLink).Return(common.ErrorNonexistentUser)
			},
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockProfileUC.NewProfileUseCase(t)
			tc.mockBehavior(m)

			req := httptest.NewRequest(http.MethodDelete, "/profiles/avatar", nil)
			if tc.setContext {
				ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
				req = req.WithContext(ctx)
			}
			rr := httptest.NewRecorder()

			newTestProfileHandler(m, nil).DeleteAvatar(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestResetUserPassword(t *testing.T) {
	validReq := dto.NewPasswordRequest{
		TokenID:          "valid_token",
		Password:         "NewPassword123",
		RepeatedPassword: "NewPassword123",
	}

	tests := []struct {
		name               string
		request            any
		mockBehavior       func(p *mockProfileUC.ProfileUseCase, m *mockMailSenderUC)
		expectedStatusCode int
	}{
		{
			name:    "Success",
			request: validReq,
			mockBehavior: func(p *mockProfileUC.ProfileUseCase, m *mockMailSenderUC) {
				m.On("CheckRecoveryCode", mock.Anything, validReq.TokenID).Return(nil)
				m.On("ExchangeTokenForUser", mock.Anything, domain.ResetToken{Token: validReq.TokenID}).Return(fixedLink, nil)
				p.On("ResetPassword", mock.Anything, domain.UpdatedPassword{UserLink: fixedLink, Password: validReq.Password}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "InvalidJSON",
			request:            "{bad_json}",
			mockBehavior:       func(p *mockProfileUC.ProfileUseCase, m *mockMailSenderUC) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "PasswordsMismatch",
			request: dto.NewPasswordRequest{
				TokenID:          "valid_token",
				Password:         "Pass12345",
				RepeatedPassword: "Pass54321",
			},
			mockBehavior:       func(p *mockProfileUC.ProfileUseCase, m *mockMailSenderUC) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:    "TokenNotFoundOnCheck",
			request: validReq,
			mockBehavior: func(p *mockProfileUC.ProfileUseCase, m *mockMailSenderUC) {
				m.On("CheckRecoveryCode", mock.Anything, validReq.TokenID).Return(common.ErrorResetTokenNotFound)
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:    "InternalErrorOnCheck",
			request: validReq,
			mockBehavior: func(p *mockProfileUC.ProfileUseCase, m *mockMailSenderUC) {
				m.On("CheckRecoveryCode", mock.Anything, validReq.TokenID).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:    "TokenNotFoundOnExchange",
			request: validReq,
			mockBehavior: func(p *mockProfileUC.ProfileUseCase, m *mockMailSenderUC) {
				m.On("CheckRecoveryCode", mock.Anything, validReq.TokenID).Return(nil)
				m.On("ExchangeTokenForUser", mock.Anything, domain.ResetToken{Token: validReq.TokenID}).Return(uuid.Nil, common.ErrorResetTokenNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:    "InternalErrorOnExchange",
			request: validReq,
			mockBehavior: func(p *mockProfileUC.ProfileUseCase, m *mockMailSenderUC) {
				m.On("CheckRecoveryCode", mock.Anything, validReq.TokenID).Return(nil)
				m.On("ExchangeTokenForUser", mock.Anything, domain.ResetToken{Token: validReq.TokenID}).Return(uuid.Nil, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:    "ResetPasswordNotNullError",
			request: validReq,
			mockBehavior: func(p *mockProfileUC.ProfileUseCase, m *mockMailSenderUC) {
				m.On("CheckRecoveryCode", mock.Anything, validReq.TokenID).Return(nil)
				m.On("ExchangeTokenForUser", mock.Anything, domain.ResetToken{Token: validReq.TokenID}).Return(fixedLink, nil)
				p.On("ResetPassword", mock.Anything, domain.UpdatedPassword{UserLink: fixedLink, Password: validReq.Password}).Return(common.ErrorNotNullValue)
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:    "ResetPasswordUserNotFound",
			request: validReq,
			mockBehavior: func(p *mockProfileUC.ProfileUseCase, m *mockMailSenderUC) {
				m.On("CheckRecoveryCode", mock.Anything, validReq.TokenID).Return(nil)
				m.On("ExchangeTokenForUser", mock.Anything, domain.ResetToken{Token: validReq.TokenID}).Return(fixedLink, nil)
				p.On("ResetPassword", mock.Anything, domain.UpdatedPassword{UserLink: fixedLink, Password: validReq.Password}).Return(common.ErrorNonexistentUser)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:    "ResetPasswordInternalError",
			request: validReq,
			mockBehavior: func(p *mockProfileUC.ProfileUseCase, m *mockMailSenderUC) {
				m.On("CheckRecoveryCode", mock.Anything, validReq.TokenID).Return(nil)
				m.On("ExchangeTokenForUser", mock.Anything, domain.ResetToken{Token: validReq.TokenID}).Return(fixedLink, nil)
				p.On("ResetPassword", mock.Anything, domain.UpdatedPassword{UserLink: fixedLink, Password: validReq.Password}).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := mockProfileUC.NewProfileUseCase(t)
			m := new(mockMailSenderUC)
			tc.mockBehavior(p, m)

			var bodyReader *bytes.Reader
			if strBody, ok := tc.request.(string); ok {
				bodyReader = bytes.NewReader([]byte(strBody))
			} else {
				b, _ := json.Marshal(tc.request)
				bodyReader = bytes.NewReader(b)
			}

			req := httptest.NewRequest(http.MethodPost, "/reset-password", bodyReader)
			rr := httptest.NewRecorder()

			newTestProfileHandler(p, m).ResetUserPassword(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}
