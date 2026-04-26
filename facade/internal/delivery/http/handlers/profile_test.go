package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/dto"
	handlerCommon "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/common"
	mockProfileUC "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/mock_profile_use_case"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"strings"

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
	SignatureTypeBytes:     512,
	MaxLenNameUser:        128,
	MaxLenDescriptionUser: 500,
	MaxReadBytes:          5 << 20,
}

func newProfileHandler(uc *mockProfileUC.ProfileUseCase) *Profile {
	return NewProfileHandler(uc, defaultProfileCfg)
}

func TestGetProfile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := mockProfileUC.NewProfileUseCase(t)
		user := domain.FullInfoUser{UserLink: fixedLink, Email: "t@mail.ru", DisplayName: "Test"}
		m.On("GetProfile", mock.Anything, fixedLink).Return(user, nil)

		req := httptest.NewRequest(http.MethodGet, "/profiles", nil)
		ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		newProfileHandler(m).GetProfile(rr, req)

		expected, _ := json.Marshal(newOkResponse(api.StatusOK, toProfileResponse(user)))
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, string(expected), rr.Body.String())
	})

	t.Run("Unauthorized", func(t *testing.T) {
		m := mockProfileUC.NewProfileUseCase(t)
		req := httptest.NewRequest(http.MethodGet, "/profiles", nil)
		rr := httptest.NewRecorder()
		newProfileHandler(m).GetProfile(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		m := mockProfileUC.NewProfileUseCase(t)
		m.On("GetProfile", mock.Anything, fixedLink).Return(domain.FullInfoUser{}, common.ErrorNonexistentUser)

		req := httptest.NewRequest(http.MethodGet, "/profiles", nil)
		ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		newProfileHandler(m).GetProfile(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestGetProfileByLink(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := mockProfileUC.NewProfileUseCase(t)
		user := domain.FullInfoUser{UserLink: fixedLink, Email: "t@mail.ru"}
		m.On("GetProfile", mock.Anything, fixedLink).Return(user, nil)

		req := httptest.NewRequest(http.MethodGet, "/profiles/"+fixedLink.String(), nil)
		req = mux.SetURLVars(req, map[string]string{"user_link": fixedLink.String()})
		rr := httptest.NewRecorder()

		newProfileHandler(m).GetProfileByLink(rr, req)

		expected, _ := json.Marshal(newOkResponse(api.StatusOK, toProfileResponse(user)))
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, string(expected), rr.Body.String())
	})

	t.Run("InvalidUUID", func(t *testing.T) {
		m := mockProfileUC.NewProfileUseCase(t)
		req := httptest.NewRequest(http.MethodGet, "/profiles/not-a-uuid", nil)
		req = mux.SetURLVars(req, map[string]string{"user_link": "not-a-uuid"})
		rr := httptest.NewRecorder()
		newProfileHandler(m).GetProfileByLink(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestUpdateProfile(t *testing.T) {
	type TestCase struct {
		Name               string
		Request            any
		SetContext         bool
		ExpectedStatusCode int
		ExpectedResponse   any
		MockBehavior       func(m *mockProfileUC.ProfileUseCase)
	}

	tests := []TestCase{
		{
			Name:               "Success",
			Request:            dto.UpdateProfileRequest{DisplayName: "NewName", Description: "desc"},
			SetContext:         true,
			ExpectedStatusCode: http.StatusOK,
			ExpectedResponse:   newOkResponse(api.StatusOK, api.StatusOK),
			MockBehavior: func(m *mockProfileUC.ProfileUseCase) {
				m.On("UpdateProfile", mock.Anything, domain.UpdatedInfo{
					UserLink:    fixedLink,
					DisplayName: "NewName",
					Description: "desc",
				}).Return(nil)
			},
		},
		{
			Name:               "Unauthorized",
			Request:            dto.UpdateProfileRequest{DisplayName: "N"},
			SetContext:         false,
			ExpectedStatusCode: http.StatusUnauthorized,
			MockBehavior:       func(m *mockProfileUC.ProfileUseCase) {},
		},
		{
			Name:               "InvalidJSON",
			SetContext:         true,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedResponse:   newErrorResponse(http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error()),
			MockBehavior:       func(m *mockProfileUC.ProfileUseCase) {},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			m := mockProfileUC.NewProfileUseCase(t)
			tc.MockBehavior(m)

			var bodyReader *bytes.Reader
			if tc.Request != nil {
				b, err := json.Marshal(tc.Request)
				require.NoError(t, err)
				bodyReader = bytes.NewReader(b)
			} else {
				bodyReader = bytes.NewReader([]byte("{bad}"))
			}

			req := httptest.NewRequest(http.MethodPost, "/profiles/info", bodyReader)
			if tc.SetContext {
				ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
				req = req.WithContext(ctx)
			}
			rr := httptest.NewRecorder()
			newProfileHandler(m).UpdateProfile(rr, req)

			assert.Equal(t, tc.ExpectedStatusCode, rr.Code)
			if tc.ExpectedResponse != nil {
				expectedJSON, err := json.Marshal(tc.ExpectedResponse)
				require.NoError(t, err)
				assert.Equal(t, string(expectedJSON), rr.Body.String())
			}
		})
	}
}

func buildAvatarRequest(t *testing.T, fileData []byte, withContext bool) *http.Request {
	t.Helper()
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

var jpegMagic = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}

func TestUpdateAvatar(t *testing.T) {
	t.Run("NoMultipart", func(t *testing.T) {
		m := mockProfileUC.NewProfileUseCase(t)
		req := httptest.NewRequest(http.MethodPut, "/profiles/avatar", strings.NewReader("plain text body"))
		ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		newProfileHandler(m).UpdateAvatar(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("InvalidMimeType", func(t *testing.T) {
		m := mockProfileUC.NewProfileUseCase(t)
		req := buildAvatarRequest(t, []byte("not-an-image-bytes"), true)
		rr := httptest.NewRecorder()
		newProfileHandler(m).UpdateAvatar(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Success", func(t *testing.T) {
		m := mockProfileUC.NewProfileUseCase(t)
		m.On("UpdateAvatar", mock.Anything, mock.MatchedBy(func(info domain.AvatarInfo) bool {
			return info.UserLink == fixedLink && info.ContentType == "image/jpeg"
		})).Return("https://cdn.example.com/avatar.jpg", nil)

		req := buildAvatarRequest(t, jpegMagic, true)
		rr := httptest.NewRecorder()
		newProfileHandler(m).UpdateAvatar(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		m := mockProfileUC.NewProfileUseCase(t)
		req := buildAvatarRequest(t, jpegMagic, false)
		rr := httptest.NewRecorder()
		newProfileHandler(m).UpdateAvatar(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		m := mockProfileUC.NewProfileUseCase(t)
		m.On("UpdateAvatar", mock.Anything, mock.Anything).Return("", common.ErrorNonexistentUser)

		req := buildAvatarRequest(t, jpegMagic, true)
		rr := httptest.NewRecorder()
		newProfileHandler(m).UpdateAvatar(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestGetProfileByLinkNotFound(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		m := mockProfileUC.NewProfileUseCase(t)
		m.On("GetProfile", mock.Anything, fixedLink).Return(domain.FullInfoUser{}, common.ErrorNonexistentUser)

		req := httptest.NewRequest(http.MethodGet, "/profiles/"+fixedLink.String(), nil)
		req = mux.SetURLVars(req, map[string]string{"user_link": fixedLink.String()})
		rr := httptest.NewRecorder()
		newProfileHandler(m).GetProfileByLink(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestUpdateProfileErrors(t *testing.T) {
	t.Run("NameTooLong", func(t *testing.T) {
		m := mockProfileUC.NewProfileUseCase(t)
		req := httptest.NewRequest(http.MethodPost, "/profiles/info",
			strings.NewReader(`{"display_name":"`+strings.Repeat("a", 200)+`"}`))
		ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		newProfileHandler(m).UpdateProfile(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("DescriptionTooLong", func(t *testing.T) {
		m := mockProfileUC.NewProfileUseCase(t)
		body, _ := json.Marshal(dto.UpdateProfileRequest{
			DisplayName: "ok",
			Description: strings.Repeat("d", 600),
		})
		req := httptest.NewRequest(http.MethodPost, "/profiles/info", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		newProfileHandler(m).UpdateProfile(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("InvalidProfileData", func(t *testing.T) {
		m := mockProfileUC.NewProfileUseCase(t)
		m.On("UpdateProfile", mock.Anything, mock.Anything).Return(common.ErrorInvalidProfileData)

		body, _ := json.Marshal(dto.UpdateProfileRequest{DisplayName: "ok", Description: "desc"})
		req := httptest.NewRequest(http.MethodPost, "/profiles/info", bytes.NewReader(body))
		ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		newProfileHandler(m).UpdateProfile(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestDeleteAvatar(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := mockProfileUC.NewProfileUseCase(t)
		m.On("DeleteAvatar", mock.Anything, fixedLink).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/profiles/avatar", nil)
		ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		newProfileHandler(m).DeleteAvatar(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		m := mockProfileUC.NewProfileUseCase(t)
		req := httptest.NewRequest(http.MethodDelete, "/profiles/avatar", nil)
		rr := httptest.NewRecorder()
		newProfileHandler(m).DeleteAvatar(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		m := mockProfileUC.NewProfileUseCase(t)
		m.On("DeleteAvatar", mock.Anything, fixedLink).Return(common.ErrorNonexistentUser)

		req := httptest.NewRequest(http.MethodDelete, "/profiles/avatar", nil)
		ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		newProfileHandler(m).DeleteAvatar(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}
