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
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	mockAppealUC "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/mock_appeal_use_case"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var defaultAppealCfg = AppealConfig{
	MultipartAttachmentFileKey: "attachment",
	MaxAttachmentSize:          5 << 20,
	MaxLenDisplayName:          128,
	MaxLenDescription:          500,
}

func newTestAppealHandler(svc AppealUsecase) *Appeal {
	return NewAppeal(svc, defaultAppealCfg)
}

func withUserCtx(req *http.Request) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserContextLink{}, fixedLink)
	return req.WithContext(ctx)
}


func TestAppealHandler_CreateAppeal(t *testing.T) {
	appealLink := uuid.New()
	validBody := map[string]string{
		"email":        "user@example.com",
		"display_name": "Alice",
		"description":  "some issue",
		"category":     "CATEGORY_TECHNICAL",
	}

	tests := []struct {
		name           string
		setCtx         bool
		body           any
		mockBehavior   func(m *mockAppealUC.AppealUsecase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			setCtx: true,
			body:   validBody,
			mockBehavior: func(m *mockAppealUC.AppealUsecase) {
				m.On("CreateAppeal", mock.Anything, mock.AnythingOfType("domain.CreateAppealInfo")).
					Return(appealLink, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "appeal_link",
		},
		{
			name:           "Unauthorized",
			setCtx:         false,
			body:           validBody,
			mockBehavior:   func(m *mockAppealUC.AppealUsecase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "user not authorized",
		},
		{
			name:           "InvalidJSON",
			setCtx:         true,
			body:           "not-json",
			mockBehavior:   func(m *mockAppealUC.AppealUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid schema",
		},
		{
			name:   "ExistingUser",
			setCtx: true,
			body:   validBody,
			mockBehavior: func(m *mockAppealUC.AppealUsecase) {
				m.On("CreateAppeal", mock.Anything, mock.AnythingOfType("domain.CreateAppealInfo")).
					Return(uuid.Nil, common.ErrorExistingUser)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   common.ErrorExistingUser.Error(),
		},
		{
			name:   "NotNullValue",
			setCtx: true,
			body:   validBody,
			mockBehavior: func(m *mockAppealUC.AppealUsecase) {
				m.On("CreateAppeal", mock.Anything, mock.AnythingOfType("domain.CreateAppealInfo")).
					Return(uuid.Nil, common.ErrorNotNullValue)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   common.ErrorNotNullValue.Error(),
		},
		{
			name:   "InternalError",
			setCtx: true,
			body:   validBody,
			mockBehavior: func(m *mockAppealUC.AppealUsecase) {
				m.On("CreateAppeal", mock.Anything, mock.AnythingOfType("domain.CreateAppealInfo")).
					Return(uuid.Nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   ErrCannotCreateAppeal.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raw, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/appeals", bytes.NewReader(raw))
			if tc.setCtx {
				req = withUserCtx(req)
			}
			rr := httptest.NewRecorder()

			m := mockAppealUC.NewAppealUsecase(t)
			tc.mockBehavior(m)
			newTestAppealHandler(m).CreateAppeal(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.expectedBody)
		})
	}
}

func TestAppealHandler_GetAppeals(t *testing.T) {
	appeals := []domain.AppealInfo{
		{
			AppealID:    1,
			AppealLink:  uuid.New(),
			Email:       "a@b.com",
			DisplayName: "Alice",
			Category:    "CATEGORY_TECHNICAL",
			Status:      "STATUS_OPEN",
			CreatedAt:   time.Now(),
		},
	}

	tests := []struct {
		name           string
		setCtx         bool
		mockBehavior   func(m *mockAppealUC.AppealUsecase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			setCtx: true,
			mockBehavior: func(m *mockAppealUC.AppealUsecase) {
				m.On("GetAppeal", mock.Anything, fixedLink).Return("user", appeals, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "a@b.com",
		},
		{
			name:           "Unauthorized",
			setCtx:         false,
			mockBehavior:   func(m *mockAppealUC.AppealUsecase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "user not authorized",
		},
		{
			name:   "InternalError",
			setCtx: true,
			mockBehavior: func(m *mockAppealUC.AppealUsecase) {
				m.On("GetAppeal", mock.Anything, fixedLink).Return("", []domain.AppealInfo{}, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   ErrCannotGetAppeals.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/appeals", nil)
			if tc.setCtx {
				req = withUserCtx(req)
			}
			rr := httptest.NewRecorder()

			m := mockAppealUC.NewAppealUsecase(t)
			tc.mockBehavior(m)
			newTestAppealHandler(m).GetAppeals(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.expectedBody)
		})
	}
}

func buildMultipart(fileKey, filename, content string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile(fileKey, filename)
	_, _ = strings.NewReader(content).WriteTo(part)
	_ = writer.Close()
	return body, writer.FormDataContentType()
}

func TestAppealHandler_UploadAttachment(t *testing.T) {
	appealLink := uuid.New()
	attachURL := "https://cdn.example.com/file.png"

	tests := []struct {
		name           string
		setCtx         bool
		setLinkVar     bool
		rawLink        string
		mockBehavior   func(m *mockAppealUC.AppealUsecase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "Success",
			setCtx:     true,
			setLinkVar: true,
			rawLink:    appealLink.String(),
			mockBehavior: func(m *mockAppealUC.AppealUsecase) {
				m.On("UploadAttachment", mock.Anything, mock.AnythingOfType("domain.UploadAttachmentInfo"), mock.Anything).
					Return(attachURL, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "attachment_url",
		},
		{
			name:           "Unauthorized",
			setCtx:         false,
			setLinkVar:     true,
			rawLink:        appealLink.String(),
			mockBehavior:   func(m *mockAppealUC.AppealUsecase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "user not authorized",
		},
		{
			name:           "MissingLinkVar",
			setCtx:         true,
			setLinkVar:     false,
			rawLink:        "",
			mockBehavior:   func(m *mockAppealUC.AppealUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrAppealLinkMissing.Error(),
		},
		{
			name:           "InvalidUUID",
			setCtx:         true,
			setLinkVar:     true,
			rawLink:        "not-a-uuid",
			mockBehavior:   func(m *mockAppealUC.AppealUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrInvalidAppealLink.Error(),
		},
		{
			name:       "ServiceError",
			setCtx:     true,
			setLinkVar: true,
			rawLink:    appealLink.String(),
			mockBehavior: func(m *mockAppealUC.AppealUsecase) {
				m.On("UploadAttachment", mock.Anything, mock.AnythingOfType("domain.UploadAttachmentInfo"), mock.Anything).
					Return("", errors.New("storage error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   ErrCannotUploadFile.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body, ct := buildMultipart(defaultAppealCfg.MultipartAttachmentFileKey, "file.png", "fake-image-data")
			req := httptest.NewRequest(http.MethodPut, "/appeals/"+tc.rawLink+"/attachment", body)
			req.Header.Set("Content-Type", ct)
			if tc.setCtx {
				req = withUserCtx(req)
			}
			if tc.setLinkVar {
				req = mux.SetURLVars(req, map[string]string{"link": tc.rawLink})
			}
			rr := httptest.NewRecorder()

			m := mockAppealUC.NewAppealUsecase(t)
			tc.mockBehavior(m)
			newTestAppealHandler(m).UploadAttachment(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.expectedBody)
		})
	}
}

// --- DeleteAppeal ---

func TestAppealHandler_DeleteAppeal(t *testing.T) {
	appealLink := uuid.New()

	tests := []struct {
		name           string
		setCtx         bool
		setLinkVar     bool
		rawLink        string
		mockBehavior   func(m *mockAppealUC.AppealUsecase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "Success",
			setCtx:     true,
			setLinkVar: true,
			rawLink:    appealLink.String(),
			mockBehavior: func(m *mockAppealUC.AppealUsecase) {
				m.On("DeleteAppeal", mock.Anything, domain.DeleteInfo{
					UserLink:   fixedLink,
					AppealLink: appealLink,
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
		{
			name:           "Unauthorized",
			setCtx:         false,
			setLinkVar:     true,
			rawLink:        appealLink.String(),
			mockBehavior:   func(m *mockAppealUC.AppealUsecase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "user not authorized",
		},
		{
			name:           "MissingLinkVar",
			setCtx:         true,
			setLinkVar:     false,
			rawLink:        "",
			mockBehavior:   func(m *mockAppealUC.AppealUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrAppealLinkMissing.Error(),
		},
		{
			name:           "InvalidUUID",
			setCtx:         true,
			setLinkVar:     true,
			rawLink:        "bad-uuid",
			mockBehavior:   func(m *mockAppealUC.AppealUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrInvalidAppealLink.Error(),
		},
		{
			name:       "ServiceError",
			setCtx:     true,
			setLinkVar: true,
			rawLink:    appealLink.String(),
			mockBehavior: func(m *mockAppealUC.AppealUsecase) {
				m.On("DeleteAppeal", mock.Anything, domain.DeleteInfo{
					UserLink:   fixedLink,
					AppealLink: appealLink,
				}).Return(errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   ErrCannotDeleteAppeal.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/appeals/"+tc.rawLink, nil)
			if tc.setCtx {
				req = withUserCtx(req)
			}
			if tc.setLinkVar {
				req = mux.SetURLVars(req, map[string]string{"link": tc.rawLink})
			}
			rr := httptest.NewRecorder()

			m := mockAppealUC.NewAppealUsecase(t)
			tc.mockBehavior(m)
			newTestAppealHandler(m).DeleteAppeal(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.expectedBody)
		})
	}
}

func TestAppealHandler_GetStats(t *testing.T) {
	stats := domain.AppealsStats{
		OpenAppeals:   3,
		InWorkAppeals: 1,
		CloseAppeals:  5,
	}

	tests := []struct {
		name           string
		setCtx         bool
		mockBehavior   func(m *mockAppealUC.AppealUsecase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			setCtx: true,
			mockBehavior: func(m *mockAppealUC.AppealUsecase) {
				m.On("GetStats", mock.Anything, fixedLink).Return(stats, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "open_appeals",
		},
		{
			name:           "Unauthorized",
			setCtx:         false,
			mockBehavior:   func(m *mockAppealUC.AppealUsecase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "user not authorized",
		},
		{
			name:   "ServiceError",
			setCtx: true,
			mockBehavior: func(m *mockAppealUC.AppealUsecase) {
				m.On("GetStats", mock.Anything, fixedLink).Return(domain.AppealsStats{}, errors.New("grpc error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   ErrCannotGetStats.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/appeals/stats", nil)
			if tc.setCtx {
				req = withUserCtx(req)
			}
			rr := httptest.NewRecorder()

			m := mockAppealUC.NewAppealUsecase(t)
			tc.mockBehavior(m)
			newTestAppealHandler(m).GetStats(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.expectedBody)
		})
	}
}

func TestAppealHandler_ChangeAppealStatus(t *testing.T) {
	appealLink := uuid.New()

	tests := []struct {
		name           string
		setCtx         bool
		setLinkVar     bool
		rawLink        string
		body           any
		mockBehavior   func(m *mockAppealUC.AppealUsecase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "Success",
			setCtx:     true,
			setLinkVar: true,
			rawLink:    appealLink.String(),
			body:       map[string]string{"new_status": "STATUS_CLOSED"},
			mockBehavior: func(m *mockAppealUC.AppealUsecase) {
				m.On("ChangeAppealStatus", mock.Anything, domain.ChangeAppealStatusInfo{
					UserLink:   fixedLink,
					AppealLink: appealLink,
					NewStatus:  "STATUS_CLOSED",
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
		{
			name:           "Unauthorized",
			setCtx:         false,
			setLinkVar:     true,
			rawLink:        appealLink.String(),
			body:           map[string]string{"new_status": "STATUS_CLOSED"},
			mockBehavior:   func(m *mockAppealUC.AppealUsecase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "user not authorized",
		},
		{
			name:           "MissingLinkVar",
			setCtx:         true,
			setLinkVar:     false,
			rawLink:        "",
			body:           map[string]string{"new_status": "STATUS_CLOSED"},
			mockBehavior:   func(m *mockAppealUC.AppealUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrAppealLinkMissing.Error(),
		},
		{
			name:           "InvalidUUID",
			setCtx:         true,
			setLinkVar:     true,
			rawLink:        "bad-uuid",
			body:           map[string]string{"new_status": "STATUS_CLOSED"},
			mockBehavior:   func(m *mockAppealUC.AppealUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   ErrInvalidAppealLink.Error(),
		},
		{
			name:           "InvalidJSONBody",
			setCtx:         true,
			setLinkVar:     true,
			rawLink:        appealLink.String(),
			body:           "not-json",
			mockBehavior:   func(m *mockAppealUC.AppealUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid schema",
		},
		{
			name:       "ServiceError",
			setCtx:     true,
			setLinkVar: true,
			rawLink:    appealLink.String(),
			body:       map[string]string{"new_status": "STATUS_CLOSED"},
			mockBehavior: func(m *mockAppealUC.AppealUsecase) {
				m.On("ChangeAppealStatus", mock.Anything, domain.ChangeAppealStatusInfo{
					UserLink:   fixedLink,
					AppealLink: appealLink,
					NewStatus:  "STATUS_CLOSED",
				}).Return(errors.New("grpc error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   ErrCannotChangeStatus.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raw, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPatch, "/appeals/"+tc.rawLink, bytes.NewReader(raw))
			if tc.setCtx {
				req = withUserCtx(req)
			}
			if tc.setLinkVar {
				req = mux.SetURLVars(req, map[string]string{"link": tc.rawLink})
			}
			rr := httptest.NewRecorder()

			m := mockAppealUC.NewAppealUsecase(t)
			tc.mockBehavior(m)
			newTestAppealHandler(m).ChangeAppealStatus(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.expectedBody)
		})
	}
}
