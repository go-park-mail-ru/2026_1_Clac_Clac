package service_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/common"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/service"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/service/dto"
	mocks "github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/service/mock_appeal_rep"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/appealRbac"
)

type mockRbacService struct {
	mock.Mock
}

func (m *mockRbacService) CheckPermission(ctx context.Context, userLink uuid.UUID, action rbac.Action) error {
	args := m.Called(ctx, userLink, action)
	return args.Error(0)
}

func (m *mockRbacService) GetUserRole(ctx context.Context, userLink uuid.UUID) (rbac.Role, error) {
	args := m.Called(ctx, userLink)
	return args.Get(0).(rbac.Role), args.Error(1)
}

func TestService_CreateAppeal(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	appealLink := uuid.New()
	dbErr := errors.New("db error")

	appeal := serviceDto.EntityAppeal{
		UserLink:    userLink,
		Mail:        "test@test.com",
		Category:    common.Categories.Bug,
		Description: "test description",
		DisplayName: "Test User",
	}

	tests := []struct {
		name          string
		setupMock     func(repo *mocks.AppealRepository, perm *mockRbacService)
		expectedLink  uuid.UUID
		expectedError error
	}{
		{
			name: "Success",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("CheckPermission", ctx, userLink, rbac.Actions.Create).Return(nil)
				repo.On("CreateAppeal", ctx, mock.AnythingOfType("dto.CreateAppealInfo")).Return(appealLink, nil)
			},
			expectedLink: appealLink,
		},
		{
			name: "Permission denied",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("CheckPermission", ctx, userLink, rbac.Actions.Create).Return(rbac.ErrActionDenied)
			},
			expectedLink:  uuid.Nil,
			expectedError: rbac.ErrActionDenied,
		},
		{
			name: "Permission check error",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("CheckPermission", ctx, userLink, rbac.Actions.Create).Return(dbErr)
			},
			expectedLink:  uuid.Nil,
			expectedError: dbErr,
		},
		{
			name: "Repository error",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("CheckPermission", ctx, userLink, rbac.Actions.Create).Return(nil)
				repo.On("CreateAppeal", ctx, mock.AnythingOfType("dto.CreateAppealInfo")).Return(uuid.UUID{}, dbErr)
			},
			expectedError: dbErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := new(mocks.AppealRepository)
			perm := new(mockRbacService)
			test.setupMock(repo, perm)

			svc := service.NewService(repo, perm)
			link, err := svc.CreateAppeal(ctx, appeal)

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedLink, link)
			}
			repo.AssertExpectations(t)
			perm.AssertExpectations(t)
		})
	}
}

func TestService_DeleteAppeal(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	appealLink := uuid.New()
	dbErr := errors.New("db error")

	tests := []struct {
		name          string
		setupMock     func(repo *mocks.AppealRepository, perm *mockRbacService)
		expectedError error
	}{
		{
			name: "Success",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("CheckPermission", ctx, userLink, rbac.Actions.Delete).Return(nil)
				repo.On("DeleteAppeal", ctx, appealLink).Return(nil)
			},
		},
		{
			name: "Permission denied",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("CheckPermission", ctx, userLink, rbac.Actions.Delete).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
		{
			name: "Repository error",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("CheckPermission", ctx, userLink, rbac.Actions.Delete).Return(nil)
				repo.On("DeleteAppeal", ctx, appealLink).Return(dbErr)
			},
			expectedError: dbErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := new(mocks.AppealRepository)
			perm := new(mockRbacService)
			test.setupMock(repo, perm)

			svc := service.NewService(repo, perm)
			err := svc.DeleteAppeal(ctx, appealLink, userLink)

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
			}
			repo.AssertExpectations(t)
			perm.AssertExpectations(t)
		})
	}
}

func TestService_ChangeAppealStatus(t *testing.T) {
	ctx := context.Background()
	supportLink := uuid.New()
	appealLink := uuid.New()
	dbErr := errors.New("db error")

	info := serviceDto.ChangeAppealStatusInfo{
		SupporterLink: supportLink,
		AppealLink:    appealLink,
		Status:        common.Statuses.InWork,
	}

	tests := []struct {
		name          string
		setupMock     func(repo *mocks.AppealRepository, perm *mockRbacService)
		expectedError error
	}{
		{
			name: "Success",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("CheckPermission", ctx, supportLink, rbac.Actions.ChangeStatus).Return(nil)
				repo.On("ChangeAppealStatus", ctx, mock.AnythingOfType("dto.ChangeAppealStatusInfo")).Return(nil)
			},
		},
		{
			name: "Permission denied",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("CheckPermission", ctx, supportLink, rbac.Actions.ChangeStatus).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
		{
			name: "Repository error",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("CheckPermission", ctx, supportLink, rbac.Actions.ChangeStatus).Return(nil)
				repo.On("ChangeAppealStatus", ctx, mock.AnythingOfType("dto.ChangeAppealStatusInfo")).Return(dbErr)
			},
			expectedError: dbErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := new(mocks.AppealRepository)
			perm := new(mockRbacService)
			test.setupMock(repo, perm)

			svc := service.NewService(repo, perm)
			err := svc.ChangeAppealStatus(ctx, info)

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
			}
			repo.AssertExpectations(t)
			perm.AssertExpectations(t)
		})
	}
}

func TestService_GetStats(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	dbErr := errors.New("db error")

	tests := []struct {
		name          string
		setupMock     func(repo *mocks.AppealRepository, perm *mockRbacService)
		expectedStats serviceDto.AppealStats
		expectedError error
	}{
		{
			name: "Success",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("CheckPermission", ctx, userLink, rbac.Actions.ViewStats).Return(nil)
				repo.On("GetStats", ctx).Return(repositoryDto.AppealStats{Open: 2, InWork: 1, Close: 5}, nil)
			},
			expectedStats: serviceDto.AppealStats{Open: 2, InWork: 1, Close: 5},
		},
		{
			name: "Permission denied",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("CheckPermission", ctx, userLink, rbac.Actions.ViewStats).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
		{
			name: "Repository error",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("CheckPermission", ctx, userLink, rbac.Actions.ViewStats).Return(nil)
				repo.On("GetStats", ctx).Return(repositoryDto.AppealStats{}, dbErr)
			},
			expectedError: dbErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := new(mocks.AppealRepository)
			perm := new(mockRbacService)
			test.setupMock(repo, perm)

			svc := service.NewService(repo, perm)
			stats, err := svc.GetStats(ctx, userLink)

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedStats, stats)
			}
			repo.AssertExpectations(t)
			perm.AssertExpectations(t)
		})
	}
}

func TestService_GetAppeals(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	dbErr := errors.New("db error")

	userAppeals := []repositoryDto.AppealEntry{
		{AppealID: 1, AppealLink: uuid.New(), Email: "user@test.com"},
	}
	openAppeals := []repositoryDto.AppealEntry{
		{AppealID: 2, AppealLink: uuid.New(), Email: "other@test.com"},
	}

	tests := []struct {
		name          string
		setupMock     func(repo *mocks.AppealRepository, perm *mockRbacService)
		expectedRole  rbac.Role
		expectedCount int
		expectedError error
	}{
		{
			name: "User sees own appeals",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("GetUserRole", ctx, userLink).Return(rbac.Roles.User, nil)
				repo.On("GetUserAppeals", ctx, userLink).Return(userAppeals, nil)
			},
			expectedRole:  rbac.Roles.User,
			expectedCount: 1,
		},
		{
			name: "Support sees open appeals",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("GetUserRole", ctx, userLink).Return(rbac.Roles.Support, nil)
				repo.On("GetOpenAppeals", ctx, userLink).Return(openAppeals, nil)
			},
			expectedRole:  rbac.Roles.Support,
			expectedCount: 1,
		},
		{
			name: "Admin sees open appeals",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("GetUserRole", ctx, userLink).Return(rbac.Roles.Admin, nil)
				repo.On("GetOpenAppeals", ctx, userLink).Return(openAppeals, nil)
			},
			expectedRole:  rbac.Roles.Admin,
			expectedCount: 1,
		},
		{
			name: "GetUserRole error",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("GetUserRole", ctx, userLink).Return(rbac.Roles.User, dbErr)
			},
			expectedError: dbErr,
		},
		{
			name: "GetUserAppeals error",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("GetUserRole", ctx, userLink).Return(rbac.Roles.User, nil)
				repo.On("GetUserAppeals", ctx, userLink).Return(nil, dbErr)
			},
			expectedError: dbErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := new(mocks.AppealRepository)
			perm := new(mockRbacService)
			test.setupMock(repo, perm)

			svc := service.NewService(repo, perm)
			appeals, err := svc.GetAppeals(ctx, userLink)

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedRole, appeals.Role)
				assert.Len(t, appeals.Appeals, test.expectedCount)
			}
			repo.AssertExpectations(t)
			perm.AssertExpectations(t)
		})
	}
}

func TestService_UploadAttachment(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	appealLink := uuid.New()
	dbErr := errors.New("db error")

	tests := []struct {
		name          string
		setupMock     func(repo *mocks.AppealRepository, perm *mockRbacService)
		expectedError error
	}{
		{
			name: "Success",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("CheckPermission", ctx, userLink, rbac.Actions.Edit).Return(nil)
				repo.On("UploadAttachment", ctx, mock.Anything, mock.AnythingOfType("string"), "image/jpeg").Return("s3-key", nil)
				repo.On("UpdateAttachmentKey", ctx, "s3-key", appealLink).Return(nil)
			},
		},
		{
			name: "Permission denied",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("CheckPermission", ctx, userLink, rbac.Actions.Edit).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
		{
			name: "Upload error",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("CheckPermission", ctx, userLink, rbac.Actions.Edit).Return(nil)
				repo.On("UploadAttachment", ctx, mock.Anything, mock.AnythingOfType("string"), "image/jpeg").Return("", dbErr)
			},
			expectedError: dbErr,
		},
		{
			name: "UpdateAttachmentKey error",
			setupMock: func(repo *mocks.AppealRepository, perm *mockRbacService) {
				perm.On("CheckPermission", ctx, userLink, rbac.Actions.Edit).Return(nil)
				repo.On("UploadAttachment", ctx, mock.Anything, mock.AnythingOfType("string"), "image/jpeg").Return("s3-key", nil)
				repo.On("UpdateAttachmentKey", ctx, "s3-key", appealLink).Return(dbErr)
			},
			expectedError: dbErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := new(mocks.AppealRepository)
			perm := new(mockRbacService)
			test.setupMock(repo, perm)

			svc := service.NewService(repo, perm)
			file := bytes.NewReader([]byte("fake image data"))
			_, err := svc.UploadAttachment(ctx, file, "image/jpeg", ".jpg", appealLink, userLink)

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
			}
			repo.AssertExpectations(t)
			perm.AssertExpectations(t)
		})
	}
}
