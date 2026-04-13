package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/handler/dto"
	mockProfileSrv "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/handler/mock_profile_srv"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/service/dto"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newOkResponse[T any](status string, data T) api.OkResponse[T] {
	return api.OkResponse[T]{
		Response: api.Response{
			Status: status,
		},
		Data: data,
	}
}

func newErrorResponse(code int, message string) api.ErrorResponse {
	return api.ErrorResponse{
		Response: api.Response{
			Status: api.StatusError,
		},
		Code:    code,
		Message: message,
	}
}

func TestGetUserProfile(t *testing.T) {
	targetUserLink := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	expectedUser := serviceDto.UserInfo{
		Link:        targetUserLink,
		DisplayName: "Artem",
		Email:       "test@mail.ru",
		AvatarURL:   "",
	}

	expectedHandlerResponse := dto.UserInfoResponse{
		Link:        targetUserLink,
		DisplayName: "Artem",
		Email:       "test@mail.ru",
		AvatarURL:   "",
	}

	vaildExtensions := map[string]struct{}{
		"image/jpg":  {},
		"image/jpeg": {},
		"image/png":  {},
		"image/webp": {},
	}

	tests := []struct {
		nameTest           string
		ctxValue           any
		mockBehavior       func(m *mockProfileSrv.ProfileService)
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			nameTest: "Success get profile",
			ctxValue: targetUserLink,
			mockBehavior: func(m *mockProfileSrv.ProfileService) {
				m.On("GetProfileUser", mock.Anything, targetUserLink).Return(expectedUser, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   newOkResponse(api.StatusOK, expectedHandlerResponse),
		},
		{
			nameTest:           "Error user not authorized",
			ctxValue:           nil,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   newErrorResponse(http.StatusUnauthorized, unauthorizedMessage),
		},
		{
			nameTest:           "Error context value is not UUID",
			ctxValue:           "invalid-uuid-string",
			mockBehavior:       nil,
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   newErrorResponse(http.StatusUnauthorized, unauthorizedMessage),
		},
		{
			nameTest: "Error user not found",
			ctxValue: targetUserLink,
			mockBehavior: func(m *mockProfileSrv.ProfileService) {
				m.On("GetProfileUser", mock.Anything, targetUserLink).Return(
					serviceDto.UserInfo{},
					common.ErrorNonexistentUser,
				)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   newErrorResponse(http.StatusNotFound, failFoundUser),
		},
		{
			nameTest: "Error from service",
			ctxValue: targetUserLink,
			mockBehavior: func(m *mockProfileSrv.ProfileService) {
				m.On("GetProfileUser", mock.Anything, targetUserLink).Return(
					serviceDto.UserInfo{},
					errors.New("database connection lost"),
				)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   newErrorResponse(http.StatusInternalServerError, failGetInfoUser),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockProfileService := mockProfileSrv.NewProfileService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockProfileService)
			}

			handler := NewHandler(mockProfileService, Config{
				ValidExtensions: vaildExtensions,
			})
			request := httptest.NewRequest(http.MethodGet, "/", nil)

			if test.ctxValue != nil {
				ctx := context.WithValue(request.Context(), middleware.UserContextLink{}, test.ctxValue)
				request = request.WithContext(ctx)
			}

			response := httptest.NewRecorder()
			handler.GetProfile(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code, "incorrect status code")

			if test.expectedResponse != nil {
				responseJson, err := json.Marshal(test.expectedResponse)
				require.NoError(t, err, "response marshal should not return error")

				assert.Equal(t, string(responseJson), response.Body.String(), "incorrect response body")
			}
		})
	}
}

func TestGetProfileByLink(t *testing.T) {
	targetUserLink := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	expectedUser := serviceDto.UserInfo{
		Link:        targetUserLink,
		DisplayName: "Artem",
		Email:       "test@mail.ru",
		AvatarURL:   "",
	}

	expectedHandlerResponse := dto.UserInfoResponse{
		Link:        targetUserLink,
		DisplayName: "Artem",
		Email:       "test@mail.ru",
		AvatarURL:   "",
	}

	tests := []struct {
		nameTest           string
		userLinkParam      string
		mockBehavior       func(m *mockProfileSrv.ProfileService)
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			nameTest:      "Success get profile by link",
			userLinkParam: targetUserLink.String(),
			mockBehavior: func(m *mockProfileSrv.ProfileService) {
				m.On("GetProfileByLink", mock.Anything, targetUserLink).Return(expectedUser, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   newOkResponse(api.StatusOK, expectedHandlerResponse),
		},
		{
			nameTest:           "Error invalid UUID",
			userLinkParam:      "invalid-uuid",
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, common.IncorrectPath),
		},
		{
			nameTest:      "Error user not found",
			userLinkParam: targetUserLink.String(),
			mockBehavior: func(m *mockProfileSrv.ProfileService) {
				m.On("GetProfileByLink", mock.Anything, targetUserLink).Return(
					serviceDto.UserInfo{},
					common.ErrorNonexistentUser,
				)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   newErrorResponse(http.StatusNotFound, failFoundUser),
		},
		{
			nameTest:      "Error from service",
			userLinkParam: targetUserLink.String(),
			mockBehavior: func(m *mockProfileSrv.ProfileService) {
				m.On("GetProfileByLink", mock.Anything, targetUserLink).Return(
					serviceDto.UserInfo{},
					errors.New("service error"),
				)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   newErrorResponse(http.StatusInternalServerError, failGetInfoUser),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockProfileService := mockProfileSrv.NewProfileService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockProfileService)
			}

			handler := NewHandler(mockProfileService, Config{})
			request := httptest.NewRequest(http.MethodGet, "/profiles/"+test.userLinkParam, nil)
			request = mux.SetURLVars(request, map[string]string{"user_link": test.userLinkParam})

			response := httptest.NewRecorder()
			handler.GetProfileByLink(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code)

			if test.expectedResponse != nil {
				responseJson, err := json.Marshal(test.expectedResponse)
				require.NoError(t, err)

				assert.Equal(t, string(responseJson), response.Body.String())
			}
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	vaildExtensions := map[string]struct{}{
		"image/jpg":  {},
		"image/jpeg": {},
		"image/png":  {},
		"image/webp": {},
	}

	ErrorIncorrectLengthName := errors.New("must contain maximum 128 symbols")
	ErrorIncorrectLengthDescription := errors.New("must contain maximum 500 symbols")

	targetUserLink := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tests := []struct {
		nameTest           string
		ctxValue           any
		requestBody        any
		mockBahavior       func(m *mockProfileSrv.ProfileService)
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			nameTest: "Success update profile",
			ctxValue: targetUserLink,
			requestBody: dto.UpdatedInfo{
				DisplayName:     "bobr",
				DescriptionUser: "ausdhakshdklashkdhaskldhklahsdh",
			},
			mockBahavior: func(m *mockProfileSrv.ProfileService) {
				m.On("UpdateProfile", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   newOkResponse(api.StatusOK, api.StatusOK),
		},
		{
			nameTest:           "Error user not authorized",
			ctxValue:           nil,
			requestBody:        dto.UpdatedInfo{DisplayName: "bobr", DescriptionUser: "desc"},
			mockBahavior:       nil,
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   newErrorResponse(http.StatusUnauthorized, unauthorizedMessage),
		},
		{
			nameTest:           "Error invalid JSON",
			ctxValue:           targetUserLink,
			requestBody:        "invalid-json-body",
			mockBahavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, common.IncorrectFormatRequest),
		},
		{
			nameTest: "Invalid len display name",
			ctxValue: targetUserLink,
			requestBody: dto.UpdatedInfo{
				DisplayName:     "bobrkkabdckjBZXKJCbzXcb.kZX>BCb ZMXNBCnbZXCb,mZBXCbZXcjkBZXJCBzjkxbCJbzxjkbckjzbxCJKBxjkbcjKBXJKCBZKJXBCjkzxbjcbZKJXBCjBZXJKCBJZKXBCjkZXBCKJZBXcjkbjzxkbCJKZBXCKBKJZXBCJKZKXJc",
				DescriptionUser: "ausdhakshdklashkdhaskldhklahsdh",
			},
			mockBahavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, fmt.Sprintf("incorrect name: %s", ErrorIncorrectLengthName)),
		},
		{
			nameTest: "Invaild len in description",
			ctxValue: targetUserLink,
			requestBody: dto.UpdatedInfo{
				DisplayName: "bobr",
				DescriptionUser: `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
			aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`,
			},
			mockBahavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, fmt.Sprintf("incorrect description: %s", ErrorIncorrectLengthDescription)),
		},
		{
			nameTest: "Error missing required field",
			ctxValue: targetUserLink,
			requestBody: dto.UpdatedInfo{
				DisplayName:     "bobr",
				DescriptionUser: "desc",
			},
			mockBahavior: func(m *mockProfileSrv.ProfileService) {
				m.On("UpdateProfile", mock.Anything, mock.Anything).Return(common.ErrorMissingRequiredField)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, failNullValue),
		},
		{
			nameTest: "Error invalid profile data",
			ctxValue: targetUserLink,
			requestBody: dto.UpdatedInfo{
				DisplayName:     "bobr",
				DescriptionUser: "desc",
			},
			mockBahavior: func(m *mockProfileSrv.ProfileService) {
				m.On("UpdateProfile", mock.Anything, mock.Anything).Return(common.ErrorInvalidProfileData)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, invalidProfileData),
		},
		{
			nameTest: "Error internal server",
			ctxValue: targetUserLink,
			requestBody: dto.UpdatedInfo{
				DisplayName:     "bobr",
				DescriptionUser: "desc",
			},
			mockBahavior: func(m *mockProfileSrv.ProfileService) {
				m.On("UpdateProfile", mock.Anything, mock.Anything).Return(errors.New(failGetInfoUser))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   newErrorResponse(http.StatusInternalServerError, failUpdateUserInfo),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockProfileService := mockProfileSrv.NewProfileService(t)
			if test.mockBahavior != nil {
				test.mockBahavior(mockProfileService)
			}

			var infoJSON []byte
			if strBody, ok := test.requestBody.(string); ok {
				infoJSON = []byte(strBody)
			} else {
				infoJSON, _ = json.Marshal(test.requestBody)
			}

			info := bytes.NewReader(infoJSON)
			request := httptest.NewRequest(http.MethodPut, "/", info)
			if test.ctxValue != nil {
				ctx := context.WithValue(context.Background(), middleware.UserContextLink{}, test.ctxValue)
				request = request.WithContext(ctx)
			}
			response := httptest.NewRecorder()

			handler := NewHandler(mockProfileService, Config{
				ValidExtensions:       vaildExtensions,
				MaxLenNameUser:        128,
				MaxLenDescriptionUser: 500,
			})
			handler.UpdateProfile(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code)

			if test.expectedResponse != nil {
				responseJSON, err := json.Marshal(test.expectedResponse)
				assert.NoError(t, err)

				assert.Equal(t, string(responseJSON), response.Body.String(), "incorrect response body")
			}
		})
	}
}

func TestUpdateAvatar(t *testing.T) {
	vaildExtensions := map[string]struct{}{
		"image/jpg":  {},
		"image/jpeg": {},
		"image/png":  {},
		"image/webp": {},
	}

	targetUserLink := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	validJpgBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}
	invalidJpgBytes := []byte("bobr bobr")

	expectedAvatarURL := "https://example.com/new_avatar.jpg"

	tests := []struct {
		nameTest           string
		ctxValue           any
		formFieldName      string
		fileContent        []byte
		invalidMultipart   bool
		mockBehavior       func(m *mockProfileSrv.ProfileService)
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			nameTest:      "Success update avatar",
			ctxValue:      targetUserLink,
			formFieldName: nameAvatarBlock,
			fileContent:   validJpgBytes,
			mockBehavior: func(m *mockProfileSrv.ProfileService) {
				m.On("UpdateAvatar", mock.Anything, mock.Anything).Return(expectedAvatarURL, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: newOkResponse(api.StatusOK, dto.AvatarResponse{
				AvatarURL: expectedAvatarURL,
			}),
		},
		{
			nameTest:           "Error parse multipart form",
			ctxValue:           targetUserLink,
			formFieldName:      nameAvatarBlock,
			fileContent:        validJpgBytes,
			invalidMultipart:   true, // Симулируем отсутствие boundary в заголовках
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, tooLargeAvatar),
		},
		{
			nameTest:           "Error invalid file field name",
			ctxValue:           targetUserLink,
			formFieldName:      "wrong_field_name",
			fileContent:        validJpgBytes,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, invalidFile),
		},
		{
			nameTest:           "Error incorrect type avatar",
			ctxValue:           targetUserLink,
			formFieldName:      nameAvatarBlock,
			fileContent:        invalidJpgBytes,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, incorrectTypeAvatar),
		},
		{
			nameTest:           "Error context lacks user UUID",
			ctxValue:           nil,
			formFieldName:      nameAvatarBlock,
			fileContent:        validJpgBytes,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   newErrorResponse(http.StatusUnauthorized, unauthorizedMessage),
		},
		{
			nameTest:      "Error user not found",
			ctxValue:      targetUserLink,
			formFieldName: nameAvatarBlock,
			fileContent:   validJpgBytes,
			mockBehavior: func(m *mockProfileSrv.ProfileService) {
				m.On("UpdateAvatar", mock.Anything, mock.Anything).Return("", common.ErrorNonexistentUser)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   newErrorResponse(http.StatusNotFound, failDeleteFile), // Согласно коду, он возвращает failDeleteFile
		},
		{
			nameTest:      "Error from service",
			ctxValue:      targetUserLink,
			formFieldName: nameAvatarBlock,
			fileContent:   validJpgBytes,
			mockBehavior: func(m *mockProfileSrv.ProfileService) {
				m.On("UpdateAvatar", mock.Anything, mock.Anything).
					Return("", errors.New("s3 bucket error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   newErrorResponse(http.StatusInternalServerError, failAvatarUrl),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockProfileService := mockProfileSrv.NewProfileService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockProfileService)
			}

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			part, err := writer.CreateFormFile(test.formFieldName, "avatar.jpg")
			assert.NoError(t, err, "not wait error")

			_, err = part.Write(test.fileContent)
			assert.NoError(t, err, "not wait error")

			err = writer.Close()
			assert.NoError(t, err, "not wait error")

			request := httptest.NewRequest(http.MethodPost, "/", body)

			if !test.invalidMultipart {
				request.Header.Set("Content-Type", writer.FormDataContentType())
			}

			if test.ctxValue != nil {
				ctx := context.WithValue(context.Background(), middleware.UserContextLink{}, targetUserLink)
				request = request.WithContext(ctx)
			}

			response := httptest.NewRecorder()

			handler := NewHandler(mockProfileService, Config{
				ValidExtensions:     vaildExtensions,
				MaxReadBytes:        5 * 1024 * 1024,
				SiganatureTypeBytes: 512,
			})

			handler.UpdateAvatar(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code)

			if test.expectedResponse != nil {
				responseJson, err := json.Marshal(test.expectedResponse)
				assert.NoError(t, err, "not wait error")

				assert.Equal(t, string(responseJson), response.Body.String())
			}
		})
	}
}

func TestDeleteAvatar(t *testing.T) {
	targetUserID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	tests := []struct {
		nameTest           string
		ctxValue           any
		mockBehavior       func(m *mockProfileSrv.ProfileService)
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			nameTest: "Success delete avatar",
			ctxValue: targetUserID,
			mockBehavior: func(m *mockProfileSrv.ProfileService) {
				m.On("DeleteAvatar", mock.Anything, targetUserID).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   newOkResponse(api.StatusOK, api.StatusOK),
		},
		{
			nameTest:           "Error user not authorized",
			ctxValue:           nil,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   newErrorResponse(http.StatusUnauthorized, unauthorizedMessage),
		},
		{
			nameTest: "Error user not found",
			ctxValue: targetUserID,
			mockBehavior: func(m *mockProfileSrv.ProfileService) {
				m.On("DeleteAvatar", mock.Anything, targetUserID).Return(common.ErrorNonexistentUser)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   newErrorResponse(http.StatusNotFound, failFoundUser),
		},
		{
			nameTest: "Error internal server",
			ctxValue: targetUserID,
			mockBehavior: func(m *mockProfileSrv.ProfileService) {
				m.On("DeleteAvatar", mock.Anything, targetUserID).Return(errors.New("database error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   newErrorResponse(http.StatusInternalServerError, failDeleteFile),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockProfileService := mockProfileSrv.NewProfileService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockProfileService)
			}

			handler := NewHandler(mockProfileService, Config{})
			request := httptest.NewRequest(http.MethodDelete, "/", nil)

			if test.ctxValue != nil {
				ctx := context.WithValue(request.Context(), middleware.UserContextLink{}, test.ctxValue)
				request = request.WithContext(ctx)
			}

			response := httptest.NewRecorder()
			handler.DeleteAvatar(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code, "incorrect status code")

			if test.expectedResponse != nil {
				responseJson, err := json.Marshal(test.expectedResponse)
				require.NoError(t, err, "not wait error")

				assert.Equal(t, string(responseJson), response.Body.String(), "incorrect response body")
			}
		})
	}
}
