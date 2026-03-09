package profile

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	mockProfileHand "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handler/profile/mock_profile_hand"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type GetProfileTestCase struct {
	Name               string
	CtxValue           any
	MockBehavior       func(m *mockProfileHand.ProfileService)
	ExpectedStatusCode int
	ExpectedResponse   any
}

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
	targetUserID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	expectedUser := models.User{
		ID:          targetUserID,
		DisplayName: "Artem",
		Email:       "test@mail.ru",
	}

	tests := []GetProfileTestCase{
		{
			Name:     "Success get profile",
			CtxValue: targetUserID,
			MockBehavior: func(m *mockProfileHand.ProfileService) {
				m.On("GetProfileUser", mock.Anything, targetUserID).Return(expectedUser, nil)
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedResponse:   newOkResponse(api.StatusOK, expectedUser),
		},
		{
			Name:               "Error user not authorized",
			CtxValue:           nil,
			MockBehavior:       nil,
			ExpectedStatusCode: http.StatusUnauthorized,
			ExpectedResponse:   newErrorResponse(http.StatusUnauthorized, unauthorizedMessage),
		},
		{
			Name:               "Error context value is not UUID",
			CtxValue:           "invalid-uuid-string",
			MockBehavior:       nil,
			ExpectedStatusCode: http.StatusUnauthorized,
			ExpectedResponse:   newErrorResponse(http.StatusUnauthorized, unauthorizedMessage),
		},
		{
			Name:     "Error from service",
			CtxValue: targetUserID,
			MockBehavior: func(m *mockProfileHand.ProfileService) {
				m.On("GetProfileUser", mock.Anything, targetUserID).Return(
					models.User{},
					errors.New("database connection lost"),
				)
			},
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedResponse:   newErrorResponse(http.StatusInternalServerError, somethingWentWrong),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			mockProfileService := mockProfileHand.NewProfileService(t)
			if test.MockBehavior != nil {
				test.MockBehavior(mockProfileService)
			}

			handler := NewProfileHandler(mockProfileService)
			request := httptest.NewRequest(http.MethodGet, "/", nil)

			if test.CtxValue != nil {
				ctx := context.WithValue(request.Context(), middleware.UserIDKey{}, test.CtxValue)
				request = request.WithContext(ctx)
			}

			response := httptest.NewRecorder()
			handler.GetProfile(response, request)

			assert.Equal(t, test.ExpectedStatusCode, response.Code, "incorrect status code")

			if test.ExpectedResponse != nil {
				responseJson, err := json.Marshal(test.ExpectedResponse)
				require.NoError(t, err, "response marshal should not return error")

				assert.Equal(t, string(responseJson), response.Body.String(), "incorrect response body")
			}
		})
	}
}
