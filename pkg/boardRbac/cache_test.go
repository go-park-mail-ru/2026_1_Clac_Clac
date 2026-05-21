package rbac

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mr.Close)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return mr, client
}

func TestCachedService_CheckPermissionOnBoard(t *testing.T) {
	ctx := context.Background()
	boardLink := uuid.New()
	userLink := uuid.New()

	tests := []struct {
		nameTest      string
		setupCache    func(mr *miniredis.Miniredis)
		mockBehavior  func(m *MockRbacRepository)
		action        Action
		expectedError error
	}{
		{
			nameTest:   "Cache miss - repo success admin edit",
			setupCache: func(mr *miniredis.Miniredis) {},
			mockBehavior: func(m *MockRbacRepository) {
				m.On("GetUserRoleByBoardLink", ctx, boardLink, userLink).Return(Roles.Admin, nil)
			},
			action:        Actions.Edit,
			expectedError: nil,
		},
		{
			nameTest: "Cache hit - allowed",
			setupCache: func(mr *miniredis.Miniredis) {
				mr.Set(roleKey(userLink, boardLink), string(Roles.Editor))
			},
			mockBehavior:  func(m *MockRbacRepository) {},
			action:        Actions.Edit,
			expectedError: nil,
		},
		{
			nameTest: "Cache hit - denied",
			setupCache: func(mr *miniredis.Miniredis) {
				mr.Set(roleKey(userLink, boardLink), string(Roles.Viewer))
			},
			mockBehavior:  func(m *MockRbacRepository) {},
			action:        Actions.Edit,
			expectedError: ErrActionDenied,
		},
		{
			nameTest:   "Cache miss - repo error",
			setupCache: func(mr *miniredis.Miniredis) {},
			mockBehavior: func(m *MockRbacRepository) {
				m.On("GetUserRoleByBoardLink", ctx, boardLink, userLink).Return(Roles.None, errors.New("db error"))
			},
			action:        Actions.Edit,
			expectedError: errors.New("db error"),
		},
		{
			nameTest:   "Cache miss - role none denied",
			setupCache: func(mr *miniredis.Miniredis) {},
			mockBehavior: func(m *MockRbacRepository) {
				m.On("GetUserRoleByBoardLink", ctx, boardLink, userLink).Return(Roles.None, nil)
			},
			action:        Actions.View,
			expectedError: ErrActionDenied,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mr, client := newTestRedis(t)
			test.setupCache(mr)

			mockRepo := new(MockRbacRepository)
			test.mockBehavior(mockRepo)

			svc := NewCachedService(mockRepo, client)
			err := svc.CheckPermissionOnBoard(ctx, boardLink, userLink, test.action)

			if test.expectedError != nil {
				if errors.Is(test.expectedError, ErrActionDenied) {
					assert.ErrorIs(t, err, ErrActionDenied)
				} else {
					assert.ErrorContains(t, err, test.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCachedService_CheckPermissionOnSection(t *testing.T) {
	ctx := context.Background()
	sectionLink := uuid.New()
	boardLink := uuid.New()
	userLink := uuid.New()

	tests := []struct {
		nameTest      string
		setupCache    func(mr *miniredis.Miniredis)
		mockBehavior  func(m *MockRbacRepository)
		action        Action
		expectedError error
	}{
		{
			nameTest:   "Cache miss - repo success",
			setupCache: func(mr *miniredis.Miniredis) {},
			mockBehavior: func(m *MockRbacRepository) {
				m.On("GetUserRoleBySectionLink", ctx, sectionLink, userLink).Return(Roles.Editor, boardLink, nil)
			},
			action:        Actions.Edit,
			expectedError: nil,
		},
		{
			nameTest: "Cache hit for mapping and role - allowed",
			setupCache: func(mr *miniredis.Miniredis) {
				mr.Set(mappingKey("section", sectionLink), boardLink.String())
				mr.Set(roleKey(userLink, boardLink), string(Roles.Admin))
			},
			mockBehavior:  func(m *MockRbacRepository) {},
			action:        Actions.Edit,
			expectedError: nil,
		},
		{
			nameTest: "Cache mapping hit, role miss - repo success",
			setupCache: func(mr *miniredis.Miniredis) {
				mr.Set(mappingKey("section", sectionLink), boardLink.String())
			},
			mockBehavior: func(m *MockRbacRepository) {
				m.On("GetUserRoleBySectionLink", ctx, sectionLink, userLink).Return(Roles.Editor, boardLink, nil)
			},
			action:        Actions.Edit,
			expectedError: nil,
		},
		{
			nameTest:   "Repo error",
			setupCache: func(mr *miniredis.Miniredis) {},
			mockBehavior: func(m *MockRbacRepository) {
				m.On("GetUserRoleBySectionLink", ctx, sectionLink, userLink).Return(Roles.None, uuid.Nil, errors.New("db error"))
			},
			action:        Actions.Edit,
			expectedError: errors.New("db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mr, client := newTestRedis(t)
			test.setupCache(mr)

			mockRepo := new(MockRbacRepository)
			test.mockBehavior(mockRepo)

			svc := NewCachedService(mockRepo, client)
			err := svc.CheckPermissionOnSection(ctx, sectionLink, userLink, test.action)

			if test.expectedError != nil {
				if errors.Is(test.expectedError, ErrActionDenied) {
					assert.ErrorIs(t, err, ErrActionDenied)
				} else {
					assert.ErrorContains(t, err, test.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCachedService_CheckPermissionOnCard(t *testing.T) {
	ctx := context.Background()
	cardLink := uuid.New()
	boardLink := uuid.New()
	userLink := uuid.New()

	t.Run("Cache miss repo success", func(t *testing.T) {
		_, client := newTestRedis(t)
		mockRepo := new(MockRbacRepository)
		mockRepo.On("GetUserRoleByCardLink", ctx, cardLink, userLink).Return(Roles.Editor, boardLink, nil)

		svc := NewCachedService(mockRepo, client)
		err := svc.CheckPermissionOnCard(ctx, cardLink, userLink, Actions.Edit)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repo denied", func(t *testing.T) {
		_, client := newTestRedis(t)
		mockRepo := new(MockRbacRepository)
		mockRepo.On("GetUserRoleByCardLink", ctx, cardLink, userLink).Return(Roles.None, uuid.Nil, nil)

		svc := NewCachedService(mockRepo, client)
		err := svc.CheckPermissionOnCard(ctx, cardLink, userLink, Actions.Edit)
		assert.ErrorIs(t, err, ErrActionDenied)
		mockRepo.AssertExpectations(t)
	})
}

func TestCachedService_CheckPermissionOnComment(t *testing.T) {
	ctx := context.Background()
	commentLink := uuid.New()
	boardLink := uuid.New()
	userLink := uuid.New()

	t.Run("Cache miss repo success", func(t *testing.T) {
		_, client := newTestRedis(t)
		mockRepo := new(MockRbacRepository)
		mockRepo.On("GetUserRoleByCommentLink", ctx, commentLink, userLink).Return(Roles.Editor, boardLink, nil)

		svc := NewCachedService(mockRepo, client)
		err := svc.CheckPermissionOnComment(ctx, commentLink, userLink, Actions.Edit)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Cache hit mapping and role denied", func(t *testing.T) {
		mr, client := newTestRedis(t)
		mr.Set(mappingKey("comment", commentLink), boardLink.String())
		mr.Set(roleKey(userLink, boardLink), string(Roles.Viewer))

		mockRepo := new(MockRbacRepository)
		svc := NewCachedService(mockRepo, client)
		err := svc.CheckPermissionOnComment(ctx, commentLink, userLink, Actions.Edit)
		assert.ErrorIs(t, err, ErrActionDenied)
	})
}

func TestCachedService_CheckPermissionOnSubtask(t *testing.T) {
	ctx := context.Background()
	subtaskLink := uuid.New()
	boardLink := uuid.New()
	userLink := uuid.New()

	t.Run("Cache miss repo success", func(t *testing.T) {
		_, client := newTestRedis(t)
		mockRepo := new(MockRbacRepository)
		mockRepo.On("GetUserRoleBySubtaskLink", ctx, subtaskLink, userLink).Return(Roles.Editor, boardLink, nil)

		svc := NewCachedService(mockRepo, client)
		err := svc.CheckPermissionOnSubtask(ctx, subtaskLink, userLink, Actions.Edit)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Cache miss repo error", func(t *testing.T) {
		_, client := newTestRedis(t)
		mockRepo := new(MockRbacRepository)
		mockRepo.On("GetUserRoleBySubtaskLink", ctx, subtaskLink, userLink).Return(Roles.None, uuid.Nil, errors.New("db fail"))

		svc := NewCachedService(mockRepo, client)
		err := svc.CheckPermissionOnSubtask(ctx, subtaskLink, userLink, Actions.Delete)
		assert.ErrorContains(t, err, "db fail")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Cache hit mapping and role allowed", func(t *testing.T) {
		mr, client := newTestRedis(t)
		mr.Set(mappingKey("subtask", subtaskLink), boardLink.String())
		mr.Set(roleKey(userLink, boardLink), string(Roles.Creator))

		mockRepo := new(MockRbacRepository)
		svc := NewCachedService(mockRepo, client)
		err := svc.CheckPermissionOnSubtask(ctx, subtaskLink, userLink, Actions.Delete)
		assert.NoError(t, err)
	})
}

func TestCachedService_CheckPermissionOnAttachment(t *testing.T) {
	ctx := context.Background()
	attachmentLink := uuid.New()
	boardLink := uuid.New()
	userLink := uuid.New()

	t.Run("Cache miss repo success", func(t *testing.T) {
		_, client := newTestRedis(t)
		mockRepo := new(MockRbacRepository)
		mockRepo.On("GetUserRoleByAttachmentLink", ctx, attachmentLink, userLink).Return(Roles.Editor, boardLink, nil)

		svc := NewCachedService(mockRepo, client)
		err := svc.CheckPermissionOnAttachment(ctx, attachmentLink, userLink, Actions.Edit)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Cache miss repo error", func(t *testing.T) {
		_, client := newTestRedis(t)
		mockRepo := new(MockRbacRepository)
		mockRepo.On("GetUserRoleByAttachmentLink", ctx, attachmentLink, userLink).Return(Roles.None, uuid.Nil, errors.New("db fail"))

		svc := NewCachedService(mockRepo, client)
		err := svc.CheckPermissionOnAttachment(ctx, attachmentLink, userLink, Actions.Delete)
		assert.ErrorContains(t, err, "db fail")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Cache hit mapping and role allowed", func(t *testing.T) {
		mr, client := newTestRedis(t)
		mr.Set(mappingKey("attachment", attachmentLink), boardLink.String())
		mr.Set(roleKey(userLink, boardLink), string(Roles.Creator))

		mockRepo := new(MockRbacRepository)
		svc := NewCachedService(mockRepo, client)
		err := svc.CheckPermissionOnAttachment(ctx, attachmentLink, userLink, Actions.Delete)
		assert.NoError(t, err)
	})

	t.Run("Cache hit mapping role denied", func(t *testing.T) {
		mr, client := newTestRedis(t)
		mr.Set(mappingKey("attachment", attachmentLink), boardLink.String())
		mr.Set(roleKey(userLink, boardLink), string(Roles.Viewer))

		mockRepo := new(MockRbacRepository)
		svc := NewCachedService(mockRepo, client)
		err := svc.CheckPermissionOnAttachment(ctx, attachmentLink, userLink, Actions.Delete)
		assert.ErrorIs(t, err, ErrActionDenied)
	})
}

func TestCachedService_InvalidateUserBoardRole(t *testing.T) {
	ctx := context.Background()
	boardLink := uuid.New()
	userLink := uuid.New()

	t.Run("Key exists - deleted successfully", func(t *testing.T) {
		mr, client := newTestRedis(t)
		mr.Set(roleKey(userLink, boardLink), string(Roles.Admin))

		mockRepo := new(MockRbacRepository)
		svc := NewCachedService(mockRepo, client)
		err := svc.InvalidateUserBoardRole(ctx, userLink, boardLink)
		assert.NoError(t, err)
	})

	t.Run("Key does not exist - no error", func(t *testing.T) {
		_, client := newTestRedis(t)
		mockRepo := new(MockRbacRepository)
		svc := NewCachedService(mockRepo, client)
		err := svc.InvalidateUserBoardRole(ctx, userLink, boardLink)
		assert.NoError(t, err)
	})
}

func TestCachedService_NewCachedService(t *testing.T) {
	_, client := newTestRedis(t)
	mockRepo := new(MockRbacRepository)
	svc := NewCachedService(mockRepo, client)
	assert.NotNil(t, svc)
}

// Ensure MockRbacRepository satisfies the Repository interface (already defined in service_test.go)
var _ Repository = (*MockRbacRepository)(nil)

// suppress unused import warning
var _ = mock.Anything
