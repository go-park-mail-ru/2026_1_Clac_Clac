package rbac

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Мок репозитория
type MockRbacRepository struct {
	mock.Mock
}

func (m *MockRbacRepository) GetUserRoleByBoardLink(ctx context.Context, boardLink uuid.UUID, userLink uuid.UUID) (Role, error) {
	args := m.Called(ctx, boardLink, userLink)
	return args.Get(0).(Role), args.Error(1)
}

func (m *MockRbacRepository) GetUserRoleBySectionLink(ctx context.Context, sectionLink uuid.UUID, userLink uuid.UUID) (Role, uuid.UUID, error) {
	args := m.Called(ctx, sectionLink, userLink)
	return args.Get(0).(Role), args.Get(1).(uuid.UUID), args.Error(2)
}

func (m *MockRbacRepository) GetUserRoleByCardLink(ctx context.Context, cardLink uuid.UUID, userLink uuid.UUID) (Role, uuid.UUID, error) {
	args := m.Called(ctx, cardLink, userLink)
	return args.Get(0).(Role), args.Get(1).(uuid.UUID), args.Error(2)
}

func (m *MockRbacRepository) GetUserRoleByCommentLink(ctx context.Context, commentLink uuid.UUID, userLink uuid.UUID) (Role, uuid.UUID, error) {
	args := m.Called(ctx, commentLink, userLink)
	return args.Get(0).(Role), args.Get(1).(uuid.UUID), args.Error(2)
}

func (m *MockRbacRepository) GetUserRoleBySubtaskLink(ctx context.Context, subtaskLink uuid.UUID, userLink uuid.UUID) (Role, uuid.UUID, error) {
	args := m.Called(ctx, subtaskLink, userLink)
	return args.Get(0).(Role), args.Get(1).(uuid.UUID), args.Error(2)
}

func TestService_CheckPermission(t *testing.T) {
	ctx := context.Background()
	itemLink := uuid.New()
	userLink := uuid.New()
	boardLink := uuid.New()

	tests := []struct {
		nameTest      string
		action        Action
		mockRole      Role
		mockError     error
		expectedError error
	}{
		{
			nameTest:      "Success - Admin can edit",
			action:        Actions.Edit,
			mockRole:      Roles.Admin,
			mockError:     nil,
			expectedError: nil,
		},
		{
			nameTest:      "Denied - Viewer cannot edit",
			action:        Actions.Edit,
			mockRole:      Roles.Viewer,
			mockError:     nil,
			expectedError: ErrActionDenied,
		},
		{
			nameTest:      "Denied - None role cannot view",
			action:        Actions.View,
			mockRole:      Roles.None,
			mockError:     nil,
			expectedError: ErrActionDenied,
		},
		{
			nameTest:      "Repository error",
			action:        Actions.View,
			mockRole:      Roles.None,
			mockError:     errors.New("db error"),
			expectedError: errors.New("db error"),
		},
	}

	for _, test := range tests {
		t.Run("Board_"+test.nameTest, func(t *testing.T) {
			mockRepo := new(MockRbacRepository)
			mockRepo.On("GetUserRoleByBoardLink", ctx, itemLink, userLink).Return(test.mockRole, test.mockError)

			svc := NewService(mockRepo)
			err := svc.CheckPermissionOnBoard(ctx, itemLink, userLink, test.action)

			if test.expectedError != nil {
				if errors.Is(test.expectedError, ErrActionDenied) {
					assert.ErrorIs(t, err, ErrActionDenied)
				} else {
					assert.ErrorContains(t, err, test.expectedError.Error())
					assert.ErrorContains(t, err, "rep.GetUserRoleByBoardLink")
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})

		t.Run("Section_"+test.nameTest, func(t *testing.T) {
			mockRepo := new(MockRbacRepository)
			mockRepo.On("GetUserRoleBySectionLink", ctx, itemLink, userLink).Return(test.mockRole, boardLink, test.mockError)

			svc := NewService(mockRepo)
			err := svc.CheckPermissionOnSection(ctx, itemLink, userLink, test.action)

			if test.expectedError != nil {
				if errors.Is(test.expectedError, ErrActionDenied) {
					assert.ErrorIs(t, err, ErrActionDenied)
				} else {
					assert.ErrorContains(t, err, test.expectedError.Error())
					assert.ErrorContains(t, err, "rep.GetUserRoleBySectionLink")
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})

		t.Run("Card_"+test.nameTest, func(t *testing.T) {
			mockRepo := new(MockRbacRepository)
			mockRepo.On("GetUserRoleByCardLink", ctx, itemLink, userLink).Return(test.mockRole, boardLink, test.mockError)

			svc := NewService(mockRepo)
			err := svc.CheckPermissionOnCard(ctx, itemLink, userLink, test.action)

			if test.expectedError != nil {
				if errors.Is(test.expectedError, ErrActionDenied) {
					assert.ErrorIs(t, err, ErrActionDenied)
				} else {
					assert.ErrorContains(t, err, test.expectedError.Error())
					assert.ErrorContains(t, err, "rep.GetUserRoleByCardLink")
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})

		t.Run("Comment_"+test.nameTest, func(t *testing.T) {
			mockRepo := new(MockRbacRepository)
			mockRepo.On("GetUserRoleByCommentLink", ctx, itemLink, userLink).Return(test.mockRole, boardLink, test.mockError)

			svc := NewService(mockRepo)
			err := svc.CheckPermissionOnComment(ctx, itemLink, userLink, test.action)

			if test.expectedError != nil {
				if errors.Is(test.expectedError, ErrActionDenied) {
					assert.ErrorIs(t, err, ErrActionDenied)
				} else {
					assert.ErrorContains(t, err, test.expectedError.Error())
					assert.ErrorContains(t, err, "rep.GetUserRoleByCommentLink")
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})

		t.Run("Subtask_"+test.nameTest, func(t *testing.T) {
			mockRepo := new(MockRbacRepository)
			mockRepo.On("GetUserRoleBySubtaskLink", ctx, itemLink, userLink).Return(test.mockRole, boardLink, test.mockError)

			svc := NewService(mockRepo)
			err := svc.CheckPermissionOnSubtask(ctx, itemLink, userLink, test.action)

			if test.expectedError != nil {
				if errors.Is(test.expectedError, ErrActionDenied) {
					assert.ErrorIs(t, err, ErrActionDenied)
				} else {
					assert.ErrorContains(t, err, test.expectedError.Error())
					assert.ErrorContains(t, err, "rep.GetUserRoleBySubtaskLink")
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
