package profile

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mockProfileHand "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handler/profile/mock_profile_hand"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUserProfile(t *testing.T) {
	targetUserID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	expectedUser := models.User{
		ID:          targetUserID,
		DisplayName: "Artem",
		Email:       "test@mail.ru",
	}

	tests := []struct {
		nameTest           string
		ctxValue           interface{}
		mockBehavior       func(m *mockProfileHand.ProfileService)
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			nameTest: "Success get profile",
			ctxValue: targetUserID,
			mockBehavior: func(m *mockProfileHand.ProfileService) {
				m.On("GetProfileUser", mock.Anything, targetUserID).Return(expectedUser, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   "{\"id\":\"11111111-1111-1111-1111-111111111111\",\"display_name\":\"Artem\",\"email\":\"test@mail.ru\",\"boards\":null}\n",
		},
		{
			nameTest:           "Error user not authorized",
			ctxValue:           nil,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "{\"error\":\"user was not authorized\"}\n",
		},
		{
			nameTest:           "Error context value is not UUID",
			ctxValue:           "invalid-uuid-string",
			mockBehavior:       nil,
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "{\"error\":\"user was not authorized\"}\n",
		},
		{
			nameTest: "Error from service",
			ctxValue: targetUserID,
			mockBehavior: func(m *mockProfileHand.ProfileService) {
				m.On("GetProfileUser", mock.Anything, targetUserID).Return(models.User{}, errors.New("database connection lost"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   "{\"error\":\"database connection lost\"}\n",
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockProfileService := mockProfileHand.NewProfileService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockProfileService)
			}

			handler := NewProfileHandler(mockProfileService)
			request := httptest.NewRequest(http.MethodGet, "/profile", nil)

			if test.ctxValue != nil {
				ctx := context.WithValue(request.Context(), middleware.UserIDKey, test.ctxValue)
				request = request.WithContext(ctx)
			}

			response := httptest.NewRecorder()
			handler.GetProfile(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code)

			if test.expectedResponse != "" {
				assert.Equal(t, test.expectedResponse, response.Body.String())
			}
		})
	}
}
