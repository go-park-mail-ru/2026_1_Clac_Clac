package rbac

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRbacRepository struct {
	mock.Mock
}

func (m *MockRbacRepository) GetUserRoleByLink(ctx context.Context, userLink uuid.UUID) (Role, error) {
	args := m.Called(ctx, userLink)
	return args.Get(0).(Role), args.Error(1)
}

func TestService_CheckPermission(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	dbErr := errors.New("db error")

	tests := []struct {
		nameTest      string
		action        Action
		mockRole      Role
		mockError     error
		expectedError error
	}{
		{
			nameTest:      "Success - User can create",
			action:        Actions.Create,
			mockRole:      Roles.User,
			mockError:     nil,
			expectedError: nil,
		},
		{
			nameTest:      "Success - Admin can view stats",
			action:        Actions.ViewStats,
			mockRole:      Roles.Admin,
			mockError:     nil,
			expectedError: nil,
		},
		{
			nameTest:      "Success - Support can change status",
			action:        Actions.ChangeStatus,
			mockRole:      Roles.Support,
			mockError:     nil,
			expectedError: nil,
		},
		{
			nameTest:      "Denied - Support cannot create",
			action:        Actions.Create,
			mockRole:      Roles.Support,
			mockError:     nil,
			expectedError: ErrActionDenied,
		},
		{
			nameTest:      "Denied - User cannot view stats",
			action:        Actions.ViewStats,
			mockRole:      Roles.User,
			mockError:     nil,
			expectedError: ErrActionDenied,
		},
		{
			nameTest:      "Repository error",
			action:        Actions.View,
			mockRole:      Roles.User,
			mockError:     dbErr,
			expectedError: dbErr,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := new(MockRbacRepository)
			mockRepo.On("GetUserRoleByLink", ctx, userLink).Return(test.mockRole, test.mockError)

			svc := NewService(mockRepo)
			err := svc.CheckPermission(ctx, userLink, test.action)

			if test.expectedError != nil {
				if errors.Is(test.expectedError, ErrActionDenied) {
					assert.ErrorIs(t, err, ErrActionDenied)
				} else {
					assert.ErrorContains(t, err, test.expectedError.Error())
					assert.ErrorContains(t, err, "rep.GetUserRoleByLink")
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetUserRole(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	dbErr := errors.New("db error")

	tests := []struct {
		nameTest      string
		mockRole      Role
		mockError     error
		expectedRole  Role
		expectedError error
	}{
		{
			nameTest:      "Success - returns Support role",
			mockRole:      Roles.Support,
			mockError:     nil,
			expectedRole:  Roles.Support,
			expectedError: nil,
		},
		{
			nameTest:      "Success - returns Admin role",
			mockRole:      Roles.Admin,
			mockError:     nil,
			expectedRole:  Roles.Admin,
			expectedError: nil,
		},
		{
			nameTest:      "Repository error - returns User role",
			mockRole:      Roles.User,
			mockError:     dbErr,
			expectedRole:  Roles.User,
			expectedError: dbErr,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepo := new(MockRbacRepository)
			mockRepo.On("GetUserRoleByLink", ctx, userLink).Return(test.mockRole, test.mockError)

			svc := NewService(mockRepo)
			role, err := svc.GetUserRole(ctx, userLink)

			assert.Equal(t, test.expectedRole, role)
			if test.expectedError != nil {
				assert.ErrorContains(t, err, test.expectedError.Error())
				assert.ErrorContains(t, err, "rep.GetUserRoleByLink")
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
