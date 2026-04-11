package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/card/handler/dto"
	mockCardSrv "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/card/handler/mock_card_srv"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/card/service/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestGetCard(t *testing.T) {
	targetCardLink := uuid.New()
	targetDataDeadLine := time.Now().Add(24 * time.Hour)
	targetExecuterName := "John Doe"

	serviceCardInfo := serviceDto.InfoCard{
		Title:        "TestTitle",
		Description:  "Test Desc",
		NameExecuter: &targetExecuterName,
		DataDeadLine: &targetDataDeadLine,
	}

	expectedResponseInfo := dto.InfoCard{
		LinkCard:     targetCardLink,
		Title:        "TestTitle",
		Description:  "Test Desc",
		NameExecuter: &targetExecuterName,
		DataDeadLine: &targetDataDeadLine,
	}

	tests := []struct {
		nameTest           string
		pathVars           map[string]string
		mockBehavior       func(m *mockCardSrv.CardService)
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			nameTest: "Success get card",
			pathVars: map[string]string{"link": targetCardLink.String()},
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("GetCard", mock.Anything, targetCardLink).Return(serviceCardInfo, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   newOkResponse(api.StatusOK, expectedResponseInfo),
		},
		{
			nameTest:           "Error invalid uuid format",
			pathVars:           map[string]string{"link": "invalid-uuid"},
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, common.IncorrectPath),
		},
		{
			nameTest: "Error card not found",
			pathVars: map[string]string{"link": targetCardLink.String()},
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("GetCard", mock.Anything, targetCardLink).Return(serviceDto.InfoCard{}, common.ErrorNotExistingCard)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   newErrorResponse(http.StatusNotFound, failFindCard),
		},
		{
			nameTest: "Error internal server",
			pathVars: map[string]string{"link": targetCardLink.String()},
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("GetCard", mock.Anything, targetCardLink).Return(serviceDto.InfoCard{}, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   newErrorResponse(http.StatusInternalServerError, failGetCard),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockCardSrv.NewCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(Deps{Srv: mockService})
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request = mux.SetURLVars(request, test.pathVars)

			response := httptest.NewRecorder()
			handler.GetCard(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code, "incorrect status code")

			if test.expectedResponse != nil {
				responseJson, err := json.Marshal(test.expectedResponse)
				assert.NoError(t, err, "response marshal should not return error")
				assert.Equal(t, string(responseJson), response.Body.String(), "incorrect response body")
			}
		})
	}
}

func TestDeleteCard(t *testing.T) {
	targetCardLink := uuid.New()

	tests := []struct {
		nameTest           string
		pathVars           map[string]string
		mockBehavior       func(m *mockCardSrv.CardService)
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			nameTest: "Success delete card",
			pathVars: map[string]string{"link": targetCardLink.String()},
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("DeleteCard", mock.Anything, targetCardLink).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   newOkResponse(api.StatusOK, api.StatusOK),
		},
		{
			nameTest:           "Error invalid uuid format",
			pathVars:           map[string]string{"link": "invalid"},
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, common.IncorrectPath),
		},
		{
			nameTest: "Error card not found",
			pathVars: map[string]string{"link": targetCardLink.String()},
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("DeleteCard", mock.Anything, targetCardLink).Return(common.ErrorNotExistingCard)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   newErrorResponse(http.StatusNotFound, failFindCard),
		},
		{
			nameTest: "Error internal server",
			pathVars: map[string]string{"link": targetCardLink.String()},
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("DeleteCard", mock.Anything, targetCardLink).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   newErrorResponse(http.StatusInternalServerError, failDeleteCard),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockCardSrv.NewCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(Deps{Srv: mockService})
			request := httptest.NewRequest(http.MethodDelete, "/", nil)
			request = mux.SetURLVars(request, test.pathVars)

			response := httptest.NewRecorder()
			handler.DeleteCard(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code, "incorrect status code")

			if test.expectedResponse != nil {
				responseJson, err := json.Marshal(test.expectedResponse)
				assert.NoError(t, err, "response marshal should not return error")
				assert.Equal(t, string(responseJson), response.Body.String(), "incorrect response body")
			}
		})
	}
}

func TestUpdateCardDetails(t *testing.T) {
	targetCardLink := uuid.New()
	targetExecuterLink := uuid.New()
	targetDataDeadLine := time.Now().Add(24 * time.Hour)

	validRequest := dto.UpdatingCardDetails{
		Title:        "UpdTitle",
		Description:  "Updated Desc",
		LinkExecuter: &targetExecuterLink,
		DataDeadLine: &targetDataDeadLine,
	}

	tests := []struct {
		nameTest           string
		pathVars           map[string]string
		requestBody        any
		mockBehavior       func(m *mockCardSrv.CardService)
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			nameTest:    "Success update card details",
			pathVars:    map[string]string{"link": targetCardLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("UpdateCardDetails", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   newOkResponse(api.StatusOK, api.StatusOK),
		},
		{
			nameTest:           "Error invalid uuid format",
			pathVars:           map[string]string{"link": "invalid"},
			requestBody:        validRequest,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, common.IncorrectPath),
		},
		{
			nameTest:           "Error invalid json",
			pathVars:           map[string]string{"link": targetCardLink.String()},
			requestBody:        "invalid json",
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, common.IncorrectRequest),
		},
		{
			nameTest: "Error max len title",
			pathVars: map[string]string{"link": targetCardLink.String()},
			requestBody: dto.UpdatingCardDetails{
				Title:       "very long title exceeding the limit",
				Description: "desc",
			},
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, "max len title is 10"),
		},
		{
			nameTest:    "Error card not found",
			pathVars:    map[string]string{"link": targetCardLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("UpdateCardDetails", mock.Anything, mock.Anything).Return(common.ErrorNotExistingCard)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   newErrorResponse(http.StatusNotFound, failFindCard),
		},
		{
			nameTest:    "Error internal server",
			pathVars:    map[string]string{"link": targetCardLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("UpdateCardDetails", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   newErrorResponse(http.StatusInternalServerError, failUpdateCard),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockCardSrv.NewCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(Deps{
				Srv:               mockService,
				MaxLenTitle:       10,
				MaxLenDescription: 1000,
			})

			var bodyBytes []byte
			if strBody, ok := test.requestBody.(string); ok {
				bodyBytes = []byte(strBody)
			} else {
				bodyBytes, _ = json.Marshal(test.requestBody)
			}

			request := httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(bodyBytes))
			request = mux.SetURLVars(request, test.pathVars)
			response := httptest.NewRecorder()
			handler.UpdateCardDetails(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code, "incorrect status code")

			if test.expectedResponse != nil {
				responseJson, err := json.Marshal(test.expectedResponse)
				assert.NoError(t, err, "response marshal should not return error")
				assert.Equal(t, string(responseJson), response.Body.String(), "incorrect response body")
			}
		})
	}
}

func TestReorderCard(t *testing.T) {
	targetCardLink := uuid.New()
	targetSectionLink := uuid.New()

	validRequest := dto.PlaceCard{
		LinkSection: targetSectionLink,
		Position:    3,
	}

	tests := []struct {
		nameTest           string
		pathVars           map[string]string
		requestBody        any
		mockBehavior       func(m *mockCardSrv.CardService)
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			nameTest:    "Success reorder card",
			pathVars:    map[string]string{"link": targetCardLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockCardSrv.CardService) {

				m.On("ReorderCard", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   newOkResponse(api.StatusOK, api.StatusOK),
		},
		{
			nameTest:           "Error invalid uuid format",
			pathVars:           map[string]string{"link": "invalid"},
			requestBody:        validRequest,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, common.IncorrectPath),
		},
		{
			nameTest:    "Error card not found",
			pathVars:    map[string]string{"link": targetCardLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything).Return(common.ErrorNotExistingCard)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   newErrorResponse(http.StatusNotFound, failFindCard),
		},
		{
			nameTest:    "Error skip mandatory section",
			pathVars:    map[string]string{"link": targetCardLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything).Return(common.ErrorSkipMandatorySection)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, incorectMoveCard),
		},
		{
			nameTest:    "Error internal server",
			pathVars:    map[string]string{"link": targetCardLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("ReorderCard", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   newErrorResponse(http.StatusInternalServerError, failReorderCard),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockCardSrv.NewCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(Deps{Srv: mockService})

			var bodyBytes []byte
			if strBody, ok := test.requestBody.(string); ok {
				bodyBytes = []byte(strBody)
			} else {
				bodyBytes, _ = json.Marshal(test.requestBody)
			}

			request := httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(bodyBytes))
			request = mux.SetURLVars(request, test.pathVars)
			response := httptest.NewRecorder()

			handler.ReorderCard(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code, "incorrect status code")

			if test.expectedResponse != nil {
				responseJson, err := json.Marshal(test.expectedResponse)
				assert.NoError(t, err, "response marshal should not return error")
				assert.Equal(t, string(responseJson), response.Body.String(), "incorrect response body")
			}
		})
	}
}
func TestCreateCard(t *testing.T) {
	targetSectionLink := uuid.New()
	targetAuthorLink := uuid.New()
	targetCardLink := uuid.New()

	validRequest := dto.NewCard{
		Title:       "New Task",
		Description: "Task desc",
		LinkSection: targetSectionLink,
	}

	serviceResult := serviceDto.PlaceCard{
		LinkCard:    targetCardLink,
		LinkSection: targetSectionLink,
		Position:    1,
	}

	expectedResult := dto.PlaceCard{
		LinkCard:    targetCardLink,
		LinkSection: targetSectionLink,
		Position:    1,
	}

	tests := []struct {
		nameTest           string
		requestBody        any
		withAuth           bool
		mockBehavior       func(m *mockCardSrv.CardService)
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			nameTest:    "Success create card",
			requestBody: validRequest,
			withAuth:    true,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(serviceResult, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   newOkResponse(api.StatusOK, expectedResult),
		},
		{
			nameTest:           "Error unauthorized",
			requestBody:        validRequest,
			withAuth:           false,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   newErrorResponse(http.StatusUnauthorized, common.FailAuthorized),
		},
		{
			nameTest:           "Error invalid json",
			requestBody:        "invalid json",
			withAuth:           true,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, common.IncorrectRequest),
		},
		{
			nameTest: "Error max len title",
			requestBody: dto.NewCard{
				Title:       "very long title exceeding the limit",
				Description: "Task desc",
				LinkSection: targetSectionLink,
			},
			withAuth:           true,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, "max len title is 10"),
		},
		{
			nameTest:    "Error section not found",
			requestBody: validRequest,
			withAuth:    true,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(serviceDto.PlaceCard{}, common.ErrorNotExistingSection)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   newErrorResponse(http.StatusNotFound, failFindSection),
		},
		{
			nameTest:    "Error internal server",
			requestBody: validRequest,
			withAuth:    true,
			mockBehavior: func(m *mockCardSrv.CardService) {
				m.On("CreateCard", mock.Anything, mock.Anything).Return(serviceDto.PlaceCard{}, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   newErrorResponse(http.StatusInternalServerError, failCreateCard),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockCardSrv.NewCardService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(Deps{
				Srv:               mockService,
				MaxLenTitle:       10,
				MaxLenDescription: 1000,
			})

			var bodyBytes []byte
			if strBody, ok := test.requestBody.(string); ok {
				bodyBytes = []byte(strBody)
			} else {
				bodyBytes, _ = json.Marshal(test.requestBody)
			}

			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(bodyBytes))

			if test.withAuth {
				ctx := context.WithValue(request.Context(), middleware.UserContextLink{}, targetAuthorLink)
				request = request.WithContext(ctx)
			}

			response := httptest.NewRecorder()
			handler.CreateCard(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code, "incorrect status code")

			if test.expectedResponse != nil {
				responseJson, err := json.Marshal(test.expectedResponse)
				assert.NoError(t, err, "response marshal should not return error")
				assert.Equal(t, string(responseJson), response.Body.String(), "incorrect response body")
			}
		})
	}
}
