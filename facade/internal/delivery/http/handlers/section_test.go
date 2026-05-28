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
	mockSectionUC "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/mock_section_use_case"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	testUserLink    = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	testSectionLink = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	testBoardLink   = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

func newTestSectionHandler(srv SectionUsecase) *Section {
	return NewSection(srv, SectionConfig{MaxLenDisplayName: 128})
}

func sectionRequest(t *testing.T, method, url string, body any, withCtx bool) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		buf.Write(b)
	}
	req := httptest.NewRequest(method, url, &buf)
	if withCtx {
		ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, testUserLink)
		req = req.WithContext(ctx)
	}
	return req
}

func TestHandlerGetSections(t *testing.T) {
	sections := []domain.SectionInfo{{Link: testSectionLink, Name: "To Do"}}

	tests := []struct {
		name               string
		setContext         bool
		boardLinkParam     string
		mockBehavior       func(m *mockSectionUC.SectionUsecase)
		expectedStatusCode int
	}{
		{
			name:           "Success",
			setContext:     true,
			boardLinkParam: testBoardLink.String(),
			mockBehavior: func(m *mockSectionUC.SectionUsecase) {
				m.On("GetSections", mock.Anything, domain.GetSectionsRequest{
					UserLink:  testUserLink,
					BoardLink: testBoardLink,
				}).Return(sections, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			boardLinkParam:     testBoardLink.String(),
			mockBehavior:       func(m *mockSectionUC.SectionUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidUUID",
			setContext:         true,
			boardLinkParam:     "not-a-uuid",
			mockBehavior:       func(m *mockSectionUC.SectionUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:           "InternalError",
			setContext:     true,
			boardLinkParam: testBoardLink.String(),
			mockBehavior: func(m *mockSectionUC.SectionUsecase) {
				m.On("GetSections", mock.Anything, mock.Anything).Return([]domain.SectionInfo(nil), errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockSectionUC.NewSectionUsecase(t)
			tc.mockBehavior(m)

			req := sectionRequest(t, http.MethodGet, "/boards/"+tc.boardLinkParam+"/sections", nil, tc.setContext)
			req = mux.SetURLVars(req, map[string]string{sectionBoardLinkKey: tc.boardLinkParam})
			rr := httptest.NewRecorder()

			newTestSectionHandler(m).GetSections(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestHandlerGetSection(t *testing.T) {
	section := domain.SectionInfo{Link: testSectionLink, Name: "To Do"}

	tests := []struct {
		name               string
		setContext         bool
		sectionLinkParam   string
		mockBehavior       func(m *mockSectionUC.SectionUsecase)
		expectedStatusCode int
	}{
		{
			name:             "Success",
			setContext:       true,
			sectionLinkParam: testSectionLink.String(),
			mockBehavior: func(m *mockSectionUC.SectionUsecase) {
				m.On("GetSection", mock.Anything, domain.GetSectionRequest{
					UserLink:    testUserLink,
					SectionLink: testSectionLink,
				}).Return(section, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			sectionLinkParam:   testSectionLink.String(),
			mockBehavior:       func(m *mockSectionUC.SectionUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidUUID",
			setContext:         true,
			sectionLinkParam:   "not-a-uuid",
			mockBehavior:       func(m *mockSectionUC.SectionUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:             "NotFound",
			setContext:       true,
			sectionLinkParam: testSectionLink.String(),
			mockBehavior: func(m *mockSectionUC.SectionUsecase) {
				m.On("GetSection", mock.Anything, mock.Anything).Return(domain.SectionInfo{}, common.ErrorSectionNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:             "InternalError",
			setContext:       true,
			sectionLinkParam: testSectionLink.String(),
			mockBehavior: func(m *mockSectionUC.SectionUsecase) {
				m.On("GetSection", mock.Anything, mock.Anything).Return(domain.SectionInfo{}, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockSectionUC.NewSectionUsecase(t)
			tc.mockBehavior(m)

			req := sectionRequest(t, http.MethodGet, "/sections/"+tc.sectionLinkParam, nil, tc.setContext)
			req = mux.SetURLVars(req, map[string]string{sectionLinkKey: tc.sectionLinkParam})
			rr := httptest.NewRecorder()

			newTestSectionHandler(m).GetSection(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestHandlerGetCards(t *testing.T) {
	cards := []domain.CardInfo{{CardLink: uuid.New(), Title: "Task 1"}}

	tests := []struct {
		name               string
		setContext         bool
		sectionLinkParam   string
		mockBehavior       func(m *mockSectionUC.SectionUsecase)
		expectedStatusCode int
	}{
		{
			name:             "Success",
			setContext:       true,
			sectionLinkParam: testSectionLink.String(),
			mockBehavior: func(m *mockSectionUC.SectionUsecase) {
				m.On("GetCards", mock.Anything, domain.GetCardsRequest{
					UserLink:    testUserLink,
					SectionLink: testSectionLink,
				}).Return(cards, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			sectionLinkParam:   testSectionLink.String(),
			mockBehavior:       func(m *mockSectionUC.SectionUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidUUID",
			setContext:         true,
			sectionLinkParam:   "not-a-uuid",
			mockBehavior:       func(m *mockSectionUC.SectionUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:             "NotFound",
			setContext:       true,
			sectionLinkParam: testSectionLink.String(),
			mockBehavior: func(m *mockSectionUC.SectionUsecase) {
				m.On("GetCards", mock.Anything, mock.Anything).Return([]domain.CardInfo(nil), common.ErrorSectionNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:             "InternalError",
			setContext:       true,
			sectionLinkParam: testSectionLink.String(),
			mockBehavior: func(m *mockSectionUC.SectionUsecase) {
				m.On("GetCards", mock.Anything, mock.Anything).Return([]domain.CardInfo(nil), errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockSectionUC.NewSectionUsecase(t)
			tc.mockBehavior(m)

			req := sectionRequest(t, http.MethodGet, "/sections/"+tc.sectionLinkParam+"/cards", nil, tc.setContext)
			req = mux.SetURLVars(req, map[string]string{sectionLinkKey: tc.sectionLinkParam})
			rr := httptest.NewRecorder()

			newTestSectionHandler(m).GetCards(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestHandlerCreateSection(t *testing.T) {
	createReq := dto.CreateSectionRequest{
		BoardLink:   testBoardLink,
		Name:        "New Section",
		IsMandatory: false,
		Color:       "red",
	}
	createdSection := domain.SectionInfo{Link: testSectionLink, Name: "New Section"}

	tests := []struct {
		name               string
		setContext         bool
		request            any
		mockBehavior       func(m *mockSectionUC.SectionUsecase)
		expectedStatusCode int
	}{
		{
			name:       "Success",
			setContext: true,
			request:    createReq,
			mockBehavior: func(m *mockSectionUC.SectionUsecase) {
				m.On("CreateSection", mock.Anything, domain.CreateSectionRequest{
					UserLink:    testUserLink,
					BoardLink:   testBoardLink,
					Name:        "New Section",
					IsMandatory: false,
					Color:       "red",
				}).Return(createdSection, nil)
			},
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			request:            createReq,
			mockBehavior:       func(m *mockSectionUC.SectionUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidJSON",
			setContext:         true,
			request:            "{bad}",
			mockBehavior:       func(m *mockSectionUC.SectionUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "EmptyName",
			setContext:         true,
			request:            dto.CreateSectionRequest{Name: ""},
			mockBehavior:       func(m *mockSectionUC.SectionUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "EmptyBoardLink",
			setContext:         true,
			request:            dto.CreateSectionRequest{Name: "Test"},
			mockBehavior:       func(m *mockSectionUC.SectionUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:       "InternalError",
			setContext: true,
			request:    createReq,
			mockBehavior: func(m *mockSectionUC.SectionUsecase) {
				m.On("CreateSection", mock.Anything, mock.Anything).Return(domain.SectionInfo{}, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockSectionUC.NewSectionUsecase(t)
			tc.mockBehavior(m)

			var bodyBytes []byte
			if s, ok := tc.request.(string); ok {
				bodyBytes = []byte(s)
			} else {
				var err error
				bodyBytes, err = json.Marshal(tc.request)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/sections", bytes.NewReader(bodyBytes))
			if tc.setContext {
				ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, testUserLink)
				req = req.WithContext(ctx)
			}
			rr := httptest.NewRecorder()

			newTestSectionHandler(m).CreateSection(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestHandlerDeleteSection(t *testing.T) {
	tests := []struct {
		name               string
		setContext         bool
		sectionLinkParam   string
		mockBehavior       func(m *mockSectionUC.SectionUsecase)
		expectedStatusCode int
	}{
		{
			name:             "Success",
			setContext:       true,
			sectionLinkParam: testSectionLink.String(),
			mockBehavior: func(m *mockSectionUC.SectionUsecase) {
				m.On("DeleteSection", mock.Anything, domain.DeleteSectionRequest{
					UserLink:    testUserLink,
					SectionLink: testSectionLink,
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			sectionLinkParam:   testSectionLink.String(),
			mockBehavior:       func(m *mockSectionUC.SectionUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidUUID",
			setContext:         true,
			sectionLinkParam:   "bad-uuid",
			mockBehavior:       func(m *mockSectionUC.SectionUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:             "NotFound",
			setContext:       true,
			sectionLinkParam: testSectionLink.String(),
			mockBehavior: func(m *mockSectionUC.SectionUsecase) {
				m.On("DeleteSection", mock.Anything, mock.Anything).Return(common.ErrorSectionNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:             "InternalError",
			setContext:       true,
			sectionLinkParam: testSectionLink.String(),
			mockBehavior: func(m *mockSectionUC.SectionUsecase) {
				m.On("DeleteSection", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockSectionUC.NewSectionUsecase(t)
			tc.mockBehavior(m)

			req := sectionRequest(t, http.MethodDelete, "/sections/"+tc.sectionLinkParam, nil, tc.setContext)
			req = mux.SetURLVars(req, map[string]string{sectionLinkKey: tc.sectionLinkParam})
			rr := httptest.NewRecorder()

			newTestSectionHandler(m).DeleteSection(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestHandlerReorderSections(t *testing.T) {
	reorderReq := dto.ListSectionLink{List: []uuid.UUID{testSectionLink}}

	tests := []struct {
		name               string
		setContext         bool
		boardLinkParam     string
		request            any
		mockBehavior       func(m *mockSectionUC.SectionUsecase)
		expectedStatusCode int
	}{
		{
			name:           "Success",
			setContext:     true,
			boardLinkParam: testBoardLink.String(),
			request:        reorderReq,
			mockBehavior: func(m *mockSectionUC.SectionUsecase) {
				m.On("ReorderSection", mock.Anything, domain.ReorderSectionRequest{
					UserLink:  testUserLink,
					BoardLink: testBoardLink,
					LinksList: []uuid.UUID{testSectionLink},
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			boardLinkParam:     testBoardLink.String(),
			request:            reorderReq,
			mockBehavior:       func(m *mockSectionUC.SectionUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidUUID",
			setContext:         true,
			boardLinkParam:     "bad-uuid",
			request:            reorderReq,
			mockBehavior:       func(m *mockSectionUC.SectionUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "InvalidJSON",
			setContext:         true,
			boardLinkParam:     testBoardLink.String(),
			request:            "{bad}",
			mockBehavior:       func(m *mockSectionUC.SectionUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:           "InternalError",
			setContext:     true,
			boardLinkParam: testBoardLink.String(),
			request:        reorderReq,
			mockBehavior: func(m *mockSectionUC.SectionUsecase) {
				m.On("ReorderSection", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockSectionUC.NewSectionUsecase(t)
			tc.mockBehavior(m)

			var bodyBytes []byte
			if s, ok := tc.request.(string); ok {
				bodyBytes = []byte(s)
			} else {
				var err error
				bodyBytes, err = json.Marshal(tc.request)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPatch, "/boards/"+tc.boardLinkParam+"/sections/reorder", bytes.NewReader(bodyBytes))
			if tc.setContext {
				ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, testUserLink)
				req = req.WithContext(ctx)
			}
			req = mux.SetURLVars(req, map[string]string{sectionBoardLinkKey: tc.boardLinkParam})
			rr := httptest.NewRecorder()

			newTestSectionHandler(m).ReorderSections(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}

func TestHandlerUpdateSection(t *testing.T) {
	updateReq := dto.SectionInfo{
		Name:        "Updated",
		IsMandatory: true,
		Color:       "blue",
	}

	tests := []struct {
		name               string
		setContext         bool
		sectionLinkParam   string
		request            any
		mockBehavior       func(m *mockSectionUC.SectionUsecase)
		expectedStatusCode int
	}{
		{
			name:             "Success",
			setContext:       true,
			sectionLinkParam: testSectionLink.String(),
			request:          updateReq,
			mockBehavior: func(m *mockSectionUC.SectionUsecase) {
				m.On("UpdateSection", mock.Anything, domain.UpdateSectionRequest{
					UserLink:    testUserLink,
					SectionLink: testSectionLink,
					Name:        "Updated",
					IsMandatory: true,
					Color:       "blue",
				}).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unauthorized",
			setContext:         false,
			sectionLinkParam:   testSectionLink.String(),
			request:            updateReq,
			mockBehavior:       func(m *mockSectionUC.SectionUsecase) {},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "InvalidUUID",
			setContext:         true,
			sectionLinkParam:   "bad-uuid",
			request:            updateReq,
			mockBehavior:       func(m *mockSectionUC.SectionUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "InvalidJSON",
			setContext:         true,
			sectionLinkParam:   testSectionLink.String(),
			request:            "{bad}",
			mockBehavior:       func(m *mockSectionUC.SectionUsecase) {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:             "NotFound",
			setContext:       true,
			sectionLinkParam: testSectionLink.String(),
			request:          updateReq,
			mockBehavior: func(m *mockSectionUC.SectionUsecase) {
				m.On("UpdateSection", mock.Anything, mock.Anything).Return(common.ErrorSectionNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:             "InternalError",
			setContext:       true,
			sectionLinkParam: testSectionLink.String(),
			request:          updateReq,
			mockBehavior: func(m *mockSectionUC.SectionUsecase) {
				m.On("UpdateSection", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockSectionUC.NewSectionUsecase(t)
			tc.mockBehavior(m)

			var bodyBytes []byte
			if s, ok := tc.request.(string); ok {
				bodyBytes = []byte(s)
			} else {
				var err error
				bodyBytes, err = json.Marshal(tc.request)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, "/sections/"+tc.sectionLinkParam, bytes.NewReader(bodyBytes))
			if tc.setContext {
				ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, testUserLink)
				req = req.WithContext(ctx)
			}
			req = mux.SetURLVars(req, map[string]string{sectionLinkKey: tc.sectionLinkParam})
			rr := httptest.NewRecorder()

			newTestSectionHandler(m).UpdateSection(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}
