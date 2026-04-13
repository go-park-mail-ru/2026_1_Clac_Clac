package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/section/handler/dto"
	mockSectionService "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/section/handler/mock_section_service"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/section/service/dto"
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

func TestGetSection(t *testing.T) {
	targetSectionLink := common.FixedSectionUuiD
	maxTasks := 50

	serviceSectionInfo := serviceDto.FullSectionInfo{
		SectionLink: targetSectionLink,
		SectionName: "To Do",
		Position:    1,
		IsMandatory: true,
		Color:       "white",
		MaxTasks:    &maxTasks,
	}

	expectedResponseInfo := dto.FullSectionInfo{
		SectionLink: targetSectionLink,
		SectionName: "To Do",
		Position:    1,
		IsMandatory: true,
		Color:       "white",
		MaxTasks:    &maxTasks,
	}

	tests := []struct {
		nameTest           string
		pathVars           map[string]string
		mockBehavior       func(m *mockSectionService.SectionService)
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			nameTest: "Success get section",
			pathVars: map[string]string{"link": targetSectionLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("GetSectionInfo", mock.Anything, targetSectionLink).Return(serviceSectionInfo, nil)
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
			nameTest: "Error section not found",
			pathVars: map[string]string{"link": targetSectionLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("GetSectionInfo", mock.Anything, targetSectionLink).Return(serviceDto.FullSectionInfo{}, common.ErrorNotExistingSection)
			},
			expectedStatusCode: http.StatusNotFound, // Исправлено на 404
			expectedResponse:   newErrorResponse(http.StatusNotFound, failFindSection),
		},
		{
			nameTest: "Error internal server",
			pathVars: map[string]string{"link": targetSectionLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("GetSectionInfo", mock.Anything, targetSectionLink).Return(serviceDto.FullSectionInfo{}, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   newErrorResponse(http.StatusInternalServerError, failGetSection),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockSectionService.NewSectionService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{})
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request = mux.SetURLVars(request, test.pathVars)

			response := httptest.NewRecorder()
			handler.GetSection(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code, "incorrect status code")

			if test.expectedResponse != nil {
				responseJson, err := json.Marshal(test.expectedResponse)
				assert.NoError(t, err, "response marshal should not return error")
				assert.Equal(t, string(responseJson), response.Body.String(), "incorrect response body")
			}
		})
	}
}

func TestCreateSection(t *testing.T) {
	boardLink := common.FixedBoardUuiD
	sectionLink := common.FixedSectionUuiD
	validMaxTasks := 50
	invalidMaxTasks := 101

	validRequest := dto.CreatingSection{
		BoardLink:   boardLink,
		SectionName: "In Progress",
		IsMandatory: false,
		Color:       "blue",
		MaxTasks:    &validMaxTasks,
	}

	invalidMaxTasksRequest := dto.CreatingSection{
		BoardLink:   boardLink,
		SectionName: "Done",
		MaxTasks:    &invalidMaxTasks,
	}

	serviceCreatingSection := serviceDto.CreatingSection{
		BoardLink:   boardLink,
		SectionName: "In Progress",
		IsMandatory: false,
		Color:       "blue",
		MaxTasks:    &validMaxTasks,
	}

	serviceResultSection := serviceDto.EntitySection{
		SectionLink: sectionLink,
		SectionName: "In Progress",
		IsMandatory: false,
		Position:    2,
		Color:       "blue",
		MaxTasks:    &validMaxTasks,
	}

	expectedResponse := dto.FullSectionInfo{
		SectionLink: sectionLink,
		SectionName: "In Progress",
		IsMandatory: false,
		Position:    2,
		Color:       "blue",
		MaxTasks:    &validMaxTasks,
	}

	tests := []struct {
		nameTest           string
		requestBody        any
		mockBehavior       func(m *mockSectionService.SectionService)
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			nameTest:    "Success create section",
			requestBody: validRequest,
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("CreateSection", mock.Anything, serviceCreatingSection).Return(serviceResultSection, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   newOkResponse(api.StatusOK, expectedResponse),
		},
		{
			nameTest:           "Error decode request",
			requestBody:        "invalid json",
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, common.IncorrectPath),
		},
		{
			nameTest:           "Error validation max tasks exceeded",
			requestBody:        invalidMaxTasksRequest,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, common.IncorrectRequest),
		},
		{
			nameTest:    "Error section already exist",
			requestBody: validRequest,
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("CreateSection", mock.Anything, serviceCreatingSection).Return(serviceDto.EntitySection{}, common.ErrorSectionAlreadyExist)
			},
			expectedStatusCode: http.StatusConflict,
			expectedResponse:   newErrorResponse(http.StatusConflict, incorrectUniqSection),
		},
		{
			nameTest:    "Error invalid reference data",
			requestBody: validRequest,
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("CreateSection", mock.Anything, serviceCreatingSection).Return(serviceDto.EntitySection{}, common.ErrorInvalidReferenceSectionData) // Исправлено на SectionData
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, incorrectReferences),
		},
		{
			nameTest:    "Error invalid section data",
			requestBody: validRequest,
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("CreateSection", mock.Anything, serviceCreatingSection).Return(serviceDto.EntitySection{}, common.ErrorInvalidSectionData)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, invalidSectionData),
		},
		{
			nameTest:    "Error missing required field",
			requestBody: validRequest,
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("CreateSection", mock.Anything, serviceCreatingSection).Return(serviceDto.EntitySection{}, common.ErrorMissingRequiredField)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, failNullValue),
		},
		{
			nameTest:    "Error internal server",
			requestBody: validRequest,
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("CreateSection", mock.Anything, serviceCreatingSection).Return(serviceDto.EntitySection{}, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   newErrorResponse(http.StatusInternalServerError, failCreateSection),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockSectionService.NewSectionService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{
				MaxQuantityTasks: 100,
				MinQuantityTasks: 0,
			})

			var bodyBytes []byte
			if strBody, ok := test.requestBody.(string); ok {
				bodyBytes = []byte(strBody)
			} else {
				bodyBytes, _ = json.Marshal(test.requestBody)
			}

			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(bodyBytes))
			response := httptest.NewRecorder()
			handler.CreateSection(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code, "incorrect status code")

			if test.expectedResponse != nil {
				responseJson, err := json.Marshal(test.expectedResponse)
				assert.NoError(t, err, "response marshal should not return error")
				assert.Equal(t, string(responseJson), response.Body.String(), "incorrect response body")
			}
		})
	}
}

func TestDeleteSection(t *testing.T) {
	targetSectionLink := common.FixedSectionUuiD

	tests := []struct {
		nameTest           string
		pathVars           map[string]string
		mockBehavior       func(m *mockSectionService.SectionService)
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			nameTest: "Success delete section",
			pathVars: map[string]string{sectionLinkKey: targetSectionLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("DeleteSection", mock.Anything, targetSectionLink).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   newOkResponse(api.StatusOK, api.StatusOK),
		},
		{
			nameTest:           "Error invalid path format",
			pathVars:           map[string]string{sectionLinkKey: "invalid"},
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, common.IncorrectPath),
		},
		{
			nameTest: "Error section not found",
			pathVars: map[string]string{sectionLinkKey: targetSectionLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("DeleteSection", mock.Anything, targetSectionLink).Return(common.ErrorNotExistingSection)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   newErrorResponse(http.StatusNotFound, failFindSection),
		},
		{
			nameTest: "Error delete backlog",
			pathVars: map[string]string{sectionLinkKey: targetSectionLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("DeleteSection", mock.Anything, targetSectionLink).Return(common.ErrorDeleteBacklog)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, failDeleteBacklog),
		},
		{
			nameTest: "Error invalid reference data",
			pathVars: map[string]string{sectionLinkKey: targetSectionLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("DeleteSection", mock.Anything, targetSectionLink).Return(common.ErrorInvalidReferenceSectionData)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, incorrectReferences),
		},
		{
			nameTest: "Error invalid card data",
			pathVars: map[string]string{sectionLinkKey: targetSectionLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("DeleteSection", mock.Anything, targetSectionLink).Return(common.ErrorInvalidCardData)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, invalidCardData),
		},
		{
			nameTest: "Error missing required field",
			pathVars: map[string]string{sectionLinkKey: targetSectionLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("DeleteSection", mock.Anything, targetSectionLink).Return(common.ErrorMissingRequiredField)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, failNullValue),
		},
		{
			nameTest: "Error internal server",
			pathVars: map[string]string{sectionLinkKey: targetSectionLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("DeleteSection", mock.Anything, targetSectionLink).Return(errors.New("db disconnect"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   newErrorResponse(http.StatusInternalServerError, failDeleteSection),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockSectionService.NewSectionService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{})
			request := httptest.NewRequest(http.MethodDelete, "/", nil)
			request = mux.SetURLVars(request, test.pathVars)

			response := httptest.NewRecorder()
			handler.DeleteSection(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code, "incorrect status code")

			if test.expectedResponse != nil {
				responseJson, err := json.Marshal(test.expectedResponse)
				assert.NoError(t, err, "response marshal should not return error")
				assert.Equal(t, string(responseJson), response.Body.String(), "incorrect response body")
			}
		})
	}
}

func TestReorderSection(t *testing.T) {
	boardLink := common.FixedBoardUuiD
	section1 := common.FixedSectionUuiD
	section2 := common.FixedSectionUuiD

	validRequest := dto.ListSectionLink{
		List: []uuid.UUID{section1, section2},
	}

	tests := []struct {
		nameTest           string
		pathVars           map[string]string
		requestBody        any
		mockBehavior       func(m *mockSectionService.SectionService)
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			nameTest:    "Success reorder sections",
			pathVars:    map[string]string{boardLinkKey: boardLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("ReorderSection", mock.Anything, boardLink, validRequest.List).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   newOkResponse(api.StatusOK, api.StatusOK),
		},
		{
			nameTest:           "Error invalid board link",
			pathVars:           map[string]string{boardLinkKey: "invalid"},
			requestBody:        validRequest,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, common.IncorrectPath),
		},
		{
			nameTest:           "Error decode request body",
			pathVars:           map[string]string{boardLinkKey: boardLink.String()},
			requestBody:        "{ bad json }",
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, common.IncorrectRequest),
		},
		{
			nameTest:    "Error not find all links",
			pathVars:    map[string]string{boardLinkKey: boardLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("ReorderSection", mock.Anything, boardLink, validRequest.List).Return(common.ErrorNotFindAllLinks)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   newErrorResponse(http.StatusNotFound, failFindSection),
		},
		{
			nameTest:    "Error invalid section data",
			pathVars:    map[string]string{boardLinkKey: boardLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("ReorderSection", mock.Anything, boardLink, validRequest.List).Return(common.ErrorInvalidSectionData)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, invalidSectionData),
		},
		{
			nameTest:    "Error invalid reference data",
			pathVars:    map[string]string{boardLinkKey: boardLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("ReorderSection", mock.Anything, boardLink, validRequest.List).Return(common.ErrorInvalidReferenceSectionData)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, incorrectReferences),
		},
		{
			nameTest:    "Error missing required field",
			pathVars:    map[string]string{boardLinkKey: boardLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("ReorderSection", mock.Anything, boardLink, validRequest.List).Return(common.ErrorMissingRequiredField)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, failNullValue),
		},
		{
			nameTest:    "Error internal server",
			pathVars:    map[string]string{boardLinkKey: boardLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("ReorderSection", mock.Anything, boardLink, validRequest.List).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   newErrorResponse(http.StatusInternalServerError, failReorderSections),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockSectionService.NewSectionService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{})

			var bodyBytes []byte
			if strBody, ok := test.requestBody.(string); ok {
				bodyBytes = []byte(strBody)
			} else {
				bodyBytes, _ = json.Marshal(test.requestBody)
			}

			request := httptest.NewRequest(http.MethodPatch, "/", bytes.NewBuffer(bodyBytes)) // Исправлено на PATCH
			request = mux.SetURLVars(request, test.pathVars)

			response := httptest.NewRecorder()
			handler.ReorderSection(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code, "incorrect status code")

			if test.expectedResponse != nil {
				responseJson, err := json.Marshal(test.expectedResponse)
				assert.NoError(t, err, "response marshal should not return error")
				assert.Equal(t, string(responseJson), response.Body.String(), "incorrect response body")
			}
		})
	}
}

func TestUpdateSection(t *testing.T) {
	sectionLink := common.FixedBoardUuiD
	maxTasks := 50
	invalidMaxTasks := 101

	validRequest := dto.FullSectionInfo{
		SectionName: "Updated Name",
		Position:    3,
		IsMandatory: true,
		Color:       "red",
		MaxTasks:    &maxTasks,
	}

	serviceUpdateInfo := serviceDto.FullSectionInfo{
		SectionLink: sectionLink,
		SectionName: "Updated Name",
		Position:    3,
		IsMandatory: true,
		Color:       "red",
		MaxTasks:    &maxTasks,
	}

	invalidNameRequest := dto.FullSectionInfo{
		SectionName: string(make([]byte, 129)),
		Color:       "red",
	}

	invalidColorRequest := dto.FullSectionInfo{
		SectionName: "Valid Name",
		Color:       "unsupported_color",
	}

	tests := []struct {
		nameTest           string
		pathVars           map[string]string
		requestBody        any
		mockBehavior       func(m *mockSectionService.SectionService)
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			nameTest:    "Success update section",
			pathVars:    map[string]string{sectionLinkKey: sectionLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("UpdateSection", mock.Anything, serviceUpdateInfo).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   newOkResponse(api.StatusOK, api.StatusOK),
		},
		{
			nameTest:           "Error invalid section link",
			pathVars:           map[string]string{sectionLinkKey: "invalid-uuid"},
			requestBody:        validRequest,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, common.IncorrectPath),
		},
		{
			nameTest:           "Error invalid json body",
			pathVars:           map[string]string{sectionLinkKey: sectionLink.String()},
			requestBody:        "invalid json",
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, common.IncorrectPath),
		},
		{
			nameTest:           "Error invalid section name length",
			pathVars:           map[string]string{sectionLinkKey: sectionLink.String()},
			requestBody:        invalidNameRequest,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, "max len name is 128"),
		},
		{
			nameTest:           "Error invalid color",
			pathVars:           map[string]string{sectionLinkKey: sectionLink.String()},
			requestBody:        invalidColorRequest,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, incorrectTypeColor),
		},
		{
			nameTest: "Error invalid max tasks",
			pathVars: map[string]string{sectionLinkKey: sectionLink.String()},
			requestBody: dto.FullSectionInfo{
				SectionName: "Updated Name",
				Color:       "red",
				MaxTasks:    &invalidMaxTasks,
			},
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, fmt.Sprintf("max quantity tasks is %d", 100)),
		},
		{
			nameTest:    "Error section not found",
			pathVars:    map[string]string{sectionLinkKey: sectionLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("UpdateSection", mock.Anything, serviceUpdateInfo).Return(common.ErrorNotExistingSection)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   newErrorResponse(http.StatusNotFound, failFindSection),
		},
		{
			nameTest:    "Error update backlog",
			pathVars:    map[string]string{sectionLinkKey: sectionLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("UpdateSection", mock.Anything, serviceUpdateInfo).Return(common.ErrorUpdateBacklog)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, failUpdateBacklog),
		},
		{
			nameTest:    "Error invalid reference data",
			pathVars:    map[string]string{sectionLinkKey: sectionLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("UpdateSection", mock.Anything, serviceUpdateInfo).Return(common.ErrorInvalidReferenceSectionData)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, incorrectReferences),
		},
		{
			nameTest:    "Error invalid section data",
			pathVars:    map[string]string{sectionLinkKey: sectionLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("UpdateSection", mock.Anything, serviceUpdateInfo).Return(common.ErrorInvalidSectionData)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, invalidSectionData),
		},
		{
			nameTest:    "Error missing required field",
			pathVars:    map[string]string{sectionLinkKey: sectionLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("UpdateSection", mock.Anything, serviceUpdateInfo).Return(common.ErrorMissingRequiredField)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, failNullValue),
		},
		{
			nameTest:    "Error internal server",
			pathVars:    map[string]string{sectionLinkKey: sectionLink.String()},
			requestBody: validRequest,
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("UpdateSection", mock.Anything, serviceUpdateInfo).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   newErrorResponse(http.StatusInternalServerError, failUpdateSection),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockSectionService.NewSectionService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{
				MaxLenNameSection: 128,
				MaxQuantityTasks:  100,
				MinQuantityTasks:  0,
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
			handler.UpdateSection(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code, "incorrect status code")

			if test.expectedResponse != nil {
				responseJson, err := json.Marshal(test.expectedResponse)
				assert.NoError(t, err, "response marshal should not return error")
				assert.Equal(t, string(responseJson), response.Body.String(), "incorrect response body")
			}
		})
	}
}

func TestGetAllSections(t *testing.T) {
	boardLink := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	sectionLink := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	maxTasks := 50

	serviceSections := []serviceDto.FullSectionInfo{
		{
			SectionLink: sectionLink,
			SectionName: "To Do",
			Position:    1,
			IsMandatory: true,
			Color:       "white",
			MaxTasks:    &maxTasks,
		},
	}

	expectedResponseInfo := dto.SectionsResponse{
		Sections: []dto.FullSectionInfo{
			{
				SectionLink: sectionLink,
				SectionName: "To Do",
				Position:    1,
				IsMandatory: true,
				Color:       "white",
				MaxTasks:    &maxTasks,
			},
		},
	}

	tests := []struct {
		nameTest           string
		pathVars           map[string]string
		mockBehavior       func(m *mockSectionService.SectionService)
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			nameTest: "Success get all sections",
			pathVars: map[string]string{boardLinkKey: boardLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("GetAllSections", mock.Anything, boardLink).Return(serviceSections, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   newOkResponse(api.StatusOK, expectedResponseInfo),
		},
		{
			nameTest:           "Error invalid board link",
			pathVars:           map[string]string{boardLinkKey: "invalid-uuid"},
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, common.IncorrectPath),
		},
		{
			nameTest: "Error internal server",
			pathVars: map[string]string{boardLinkKey: boardLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("GetAllSections", mock.Anything, boardLink).Return(nil, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   newErrorResponse(http.StatusInternalServerError, failGetAllSections),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockSectionService.NewSectionService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{})
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request = mux.SetURLVars(request, test.pathVars)

			response := httptest.NewRecorder()
			handler.GetAllSections(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code, "incorrect status code")

			if test.expectedResponse != nil {
				responseJson, err := json.Marshal(test.expectedResponse)
				assert.NoError(t, err, "response marshal should not return error")
				assert.Equal(t, string(responseJson), response.Body.String(), "incorrect response body")
			}
		})
	}
}

func TestGetCards(t *testing.T) {
	targetSectionLink := uuid.New()
	targetExecuterName := "John Doe"
	targetDeadLine := time.Now().Add(24 * time.Hour)

	serviceCards := []serviceDto.Card{
		{
			CardLink:     uuid.New(),
			ExecuterName: &targetExecuterName,
			Title:        "Task 1",
			DeadLine:     &targetDeadLine,
		},
	}

	expectedResponseInfo := dto.CardsSection{
		Cards: []dto.Card{
			{
				CardLink:     serviceCards[0].CardLink,
				ExecuterName: serviceCards[0].ExecuterName,
				Title:        serviceCards[0].Title,
				DeadLine:     serviceCards[0].DeadLine,
			},
		},
	}

	tests := []struct {
		nameTest           string
		pathVars           map[string]string
		mockBehavior       func(m *mockSectionService.SectionService)
		expectedStatusCode int
		expectedResponse   any
	}{
		{
			nameTest: "Success get cards",
			pathVars: map[string]string{sectionLinkKey: targetSectionLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("GetCards", mock.Anything, targetSectionLink).Return(serviceCards, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   newOkResponse(api.StatusOK, expectedResponseInfo),
		},
		{
			nameTest:           "Error invalid uuid format",
			pathVars:           map[string]string{sectionLinkKey: "invalid-uuid"},
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   newErrorResponse(http.StatusBadRequest, common.IncorrectPath),
		},
		{
			nameTest: "Error section not found",
			pathVars: map[string]string{sectionLinkKey: targetSectionLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("GetCards", mock.Anything, targetSectionLink).Return([]serviceDto.Card{}, common.ErrorNotExistingSection)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   newErrorResponse(http.StatusNotFound, failFindSection),
		},
		{
			nameTest: "Error internal server",
			pathVars: map[string]string{sectionLinkKey: targetSectionLink.String()},
			mockBehavior: func(m *mockSectionService.SectionService) {
				m.On("GetCards", mock.Anything, targetSectionLink).Return([]serviceDto.Card{}, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   newErrorResponse(http.StatusInternalServerError, failGetCards),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockService := mockSectionService.NewSectionService(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockService)
			}

			handler := NewHandler(mockService, Config{})
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request = mux.SetURLVars(request, test.pathVars)

			response := httptest.NewRecorder()
			handler.GetCards(response, request)

			assert.Equal(t, test.expectedStatusCode, response.Code, "incorrect status code")

			if test.expectedResponse != nil {
				responseJson, err := json.Marshal(test.expectedResponse)
				assert.NoError(t, err, "response marshal should not return error")
				assert.Equal(t, string(responseJson), response.Body.String(), "incorrect response body")
			}
		})
	}
}
