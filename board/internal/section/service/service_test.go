package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/models"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/service/dto"
	mockSectionRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/service/mock_section_rep"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRbacService struct {
	mock.Mock
}

func (m *MockRbacService) CheckPermissionOnBoard(ctx context.Context, itemLink uuid.UUID, userLink uuid.UUID, action rbac.Action) error {
	args := m.Called(ctx, itemLink, userLink, action)
	return args.Error(0)
}

func (m *MockRbacService) CheckPermissionOnSection(ctx context.Context, itemLink uuid.UUID, userLink uuid.UUID, action rbac.Action) error {
	args := m.Called(ctx, itemLink, userLink, action)
	return args.Error(0)
}

func (m *MockRbacService) CheckPermissionOnCard(ctx context.Context, itemLink uuid.UUID, userLink uuid.UUID, action rbac.Action) error {
	args := m.Called(ctx, itemLink, userLink, action)
	return args.Error(0)
}

func (m *MockRbacService) CheckPermissionOnComment(ctx context.Context, itemLink uuid.UUID, userLink uuid.UUID, action rbac.Action) error {
	args := m.Called(ctx, itemLink, userLink, action)
	return args.Error(0)
}

func (m *MockRbacService) CheckPermissionOnSubtask(ctx context.Context, itemLink uuid.UUID, userLink uuid.UUID, action rbac.Action) error {
	args := m.Called(ctx, itemLink, userLink, action)
	return args.Error(0)
}

func TestServiceGetSection(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	targetLink := uuid.New()
	maxTasks := 50

	repoResult := repositoryDto.FullSectionInfo{
		SectionLink: targetLink,
		SectionName: "To Do",
		Position:    2,
		IsMandatory: true,
		Color:       "white",
		MaxTasks:    &maxTasks,
	}

	expectedResult := dto.FullSectionInfo{
		SectionLink: targetLink,
		SectionName: "To Do",
		Position:    2,
		IsMandatory: true,
		Color:       "white",
		MaxTasks:    &maxTasks,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockSectionRep.SectionRepository, r *MockRbacService)
		expectedError error
		expectedData  dto.FullSectionInfo
	}{
		{
			nameTest: "Success get section info",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", ctx, targetLink, userLink, mock.Anything).Return(nil)
				m.On("GetSection", ctx, targetLink).Return(repoResult, nil)
			},
			expectedError: nil,
			expectedData:  expectedResult,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", ctx, targetLink, userLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
			expectedData:  dto.FullSectionInfo{},
		},
		{
			nameTest: "Error RBAC internal",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", ctx, targetLink, userLink, mock.Anything).Return(errors.New("rbac fail"))
			},
			expectedError: errors.New("SectionService.CheckPermissionOnSection: rbac fail"),
			expectedData:  dto.FullSectionInfo{},
		},
		{
			nameTest: "Error from repository",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", ctx, targetLink, userLink, mock.Anything).Return(nil)
				m.On("GetSection", ctx, targetLink).Return(repositoryDto.FullSectionInfo{}, errors.New("db disconnect"))
			},
			expectedError: errors.New("SectionRepository.GetSectionInfo: db disconnect"),
			expectedData:  dto.FullSectionInfo{},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepository := mockSectionRep.NewSectionRepository(t)
			mockRbac := new(MockRbacService)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepository, mockRbac)
			}

			service := NewService(mockRepository, mockRbac)
			result, err := service.GetSection(ctx, targetLink, userLink)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedData, result)
			}
		})
	}
}

func TestServiceGetSections(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	boardLink := uuid.New()
	sectionLink := uuid.New()
	maxTasks := 50

	repoSections := []repositoryDto.FullSectionInfo{
		{
			SectionLink: sectionLink,
			SectionName: "To Do",
			Position:    1,
			IsMandatory: true,
			Color:       "white",
			MaxTasks:    &maxTasks,
		},
	}

	expectedResult := []dto.FullSectionInfo{
		{
			SectionLink: sectionLink,
			SectionName: "To Do",
			Position:    1,
			IsMandatory: true,
			Color:       "white",
			MaxTasks:    &maxTasks,
		},
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockSectionRep.SectionRepository, r *MockRbacService)
		expectedError error
		expectedData  []dto.FullSectionInfo
	}{
		{
			nameTest: "Success get all sections",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnBoard", ctx, boardLink, userLink, mock.Anything).Return(nil)
				m.On("GetSections", ctx, boardLink).Return(repoSections, nil)
			},
			expectedError: nil,
			expectedData:  expectedResult,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnBoard", ctx, boardLink, userLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
			expectedData:  []dto.FullSectionInfo{},
		},
		{
			nameTest: "Error from repository",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnBoard", ctx, boardLink, userLink, mock.Anything).Return(nil)
				m.On("GetSections", ctx, boardLink).Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("SectionRepository.GetAllSections: db error"),
			expectedData:  []dto.FullSectionInfo{},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepository := mockSectionRep.NewSectionRepository(t)
			mockRbac := new(MockRbacService)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepository, mockRbac)
			}

			service := NewService(mockRepository, mockRbac)
			result, err := service.GetSections(ctx, boardLink, userLink)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedData, result)
			}
		})
	}
}

func TestServiceGetCards(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	targetSectionLink := uuid.New()
	targetExecutorLink := uuid.New()
	targetDeadLine := time.Now()

	repCards := []repositoryDto.Card{
		{
			CardLink:      uuid.New(),
			ExecutorLink:  &targetExecutorLink,
			Title:         "Task 1",
			DeadLine:      &targetDeadLine,
			Subtasks: []models.SubtaskInfo{
				{
					SubtaskLink: uuid.New(),
					Description: "Subtask 1",
					IsDone:      true,
					Position:    1,
				},
			},
		},
	}

	expectedCards := []dto.Card{
		{
			CardLink:      repCards[0].CardLink,
			ExecutorLink:  repCards[0].ExecutorLink,
			Title:         repCards[0].Title,
			DeadLine:      repCards[0].DeadLine,
			Subtasks: []models.SubtaskInfo{
				{
					SubtaskLink: repCards[0].Subtasks[0].SubtaskLink,
					Description: repCards[0].Subtasks[0].Description,
					IsDone:      repCards[0].Subtasks[0].IsDone,
					Position:    repCards[0].Subtasks[0].Position,
				},
			},
		},
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockSectionRep.SectionRepository, r *MockRbacService)
		expectedError error
		expectedRes   []dto.Card
	}{
		{
			nameTest: "Success get cards",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", ctx, targetSectionLink, userLink, mock.Anything).Return(nil)
				m.On("GetCards", ctx, targetSectionLink).Return(repCards, nil)
			},
			expectedError: nil,
			expectedRes:   expectedCards,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", ctx, targetSectionLink, userLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
			expectedRes:   []dto.Card{},
		},
		{
			nameTest: "Error from repository",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", ctx, targetSectionLink, userLink, mock.Anything).Return(nil)
				m.On("GetCards", ctx, targetSectionLink).Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("SectionRepository.GetCards: db error"),
			expectedRes:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepository := mockSectionRep.NewSectionRepository(t)
			mockRbac := new(MockRbacService)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepository, mockRbac)
			}

			service := NewService(mockRepository, mockRbac)
			res, err := service.GetCards(ctx, targetSectionLink, userLink)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedRes, res)
			}
		})
	}
}
func TestServiceCreateSection(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	boardLink := uuid.New()
	maxTasks := 20

	inputDto := dto.CreatingSection{
		BoardLink:   boardLink,
		SectionName: "In Progress",
		IsMandatory: false,
		Color:       "blue",
		MaxTasks:    &maxTasks,
	}

	repoResult := repositoryDto.FullSectionInfo{
		SectionName: "In Progress",
		Position:    3,
		IsMandatory: false,
		Color:       "blue",
		MaxTasks:    &maxTasks,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockSectionRep.SectionRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest: "Success create section",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnBoard", ctx, boardLink, userLink, mock.Anything).Return(nil)
				m.On("CreateSection", ctx, mock.MatchedBy(func(req repositoryDto.CreatingSection) bool {
					return req.BoardLink == boardLink &&
						req.SectionName == "In Progress" &&
						req.Color == "blue" &&
						*req.MaxTasks == maxTasks &&
						req.SectionLink != uuid.Nil
				})).Return(repoResult, nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnBoard", ctx, boardLink, userLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
		{
			nameTest: "Error from repository",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnBoard", ctx, boardLink, userLink, mock.Anything).Return(nil)
				m.On("CreateSection", ctx, mock.Anything).Return(repositoryDto.FullSectionInfo{}, errors.New("insert fail"))
			},
			expectedError: errors.New("SectionRepository.CreateSection: insert fail"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepository := mockSectionRep.NewSectionRepository(t)
			mockRbac := new(MockRbacService)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepository, mockRbac)
			}

			service := NewService(mockRepository, mockRbac)
			result, err := service.CreateSection(ctx, inputDto, userLink)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, result.SectionLink)
				assert.Equal(t, repoResult.SectionName, result.SectionName)
				assert.Equal(t, repoResult.Position, result.Position)
				assert.Equal(t, repoResult.Color, result.Color)
				assert.Equal(t, repoResult.IsMandatory, result.IsMandatory)
				assert.Equal(t, repoResult.MaxTasks, result.MaxTasks)
			}
		})
	}
}

func TestServiceDeleteSection(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	targetLink := uuid.New()

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockSectionRep.SectionRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest: "Success delete section",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", ctx, targetLink, userLink, mock.Anything).Return(nil)
				m.On("GetSection", ctx, targetLink).Return(repositoryDto.FullSectionInfo{Position: 2}, nil)
				m.On("DeleteSection", ctx, targetLink).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", ctx, targetLink, userLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
		{
			nameTest: "Error cannot delete backlog",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", ctx, targetLink, userLink, mock.Anything).Return(nil)
				m.On("GetSection", ctx, targetLink).Return(repositoryDto.FullSectionInfo{Position: 1}, nil)
			},
			expectedError: common.ErrCannotDeleteBacklog,
		},
		{
			nameTest: "Error delete section repository fail",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", ctx, targetLink, userLink, mock.Anything).Return(nil)
				m.On("GetSection", ctx, targetLink).Return(repositoryDto.FullSectionInfo{Position: 2}, nil)
				m.On("DeleteSection", ctx, targetLink).Return(errors.New("db timeout"))
			},
			expectedError: errors.New("SectionRepository.DeleteSection: db timeout"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepository := mockSectionRep.NewSectionRepository(t)
			mockRbac := new(MockRbacService)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepository, mockRbac)
			}

			service := NewService(mockRepository, mockRbac)
			err := service.DeleteSection(ctx, targetLink, userLink)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServiceReorderSection(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	boardLink := uuid.New()
	section1 := uuid.New()
	section2 := uuid.New()

	tests := []struct {
		nameTest      string
		sectionLinks  []uuid.UUID
		mockBehavior  func(m *mockSectionRep.SectionRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest:     "Success empty list",
			sectionLinks: []uuid.UUID{},
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnBoard", ctx, boardLink, userLink, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest:     "Success reorder",
			sectionLinks: []uuid.UUID{section1, section2},
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnBoard", ctx, boardLink, userLink, mock.Anything).Return(nil)
				m.On("ReorderSection", ctx, boardLink, []uuid.UUID{section1, section2}).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest:     "Error permission denied",
			sectionLinks: []uuid.UUID{section1, section2},
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnBoard", ctx, boardLink, userLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
		{
			nameTest:     "Error from repository",
			sectionLinks: []uuid.UUID{section1, section2},
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnBoard", ctx, boardLink, userLink, mock.Anything).Return(nil)
				m.On("ReorderSection", ctx, boardLink, []uuid.UUID{section1, section2}).Return(errors.New("deadlock"))
			},
			expectedError: errors.New("SectionRepository.ReorderSection: deadlock"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepository := mockSectionRep.NewSectionRepository(t)
			mockRbac := new(MockRbacService)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepository, mockRbac)
			}

			service := NewService(mockRepository, mockRbac)
			err := service.ReorderSection(ctx, boardLink, test.sectionLinks, userLink)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServiceUpdateSection(t *testing.T) {
	ctx := context.Background()
	userLink := uuid.New()
	targetLink := uuid.New()
	maxTasks := 50

	updateDto := dto.FullSectionInfo{
		SectionLink: targetLink,
		SectionName: "Updated Name",
		Position:    3,
		IsMandatory: false,
		Color:       "red",
		MaxTasks:    &maxTasks,
	}

	expectedRepoDto := repositoryDto.FullSectionInfo{
		SectionLink: targetLink,
		SectionName: "Updated Name",
		Position:    3,
		IsMandatory: false,
		Color:       "red",
		MaxTasks:    &maxTasks,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockSectionRep.SectionRepository, r *MockRbacService)
		expectedError error
	}{
		{
			nameTest: "Success update section",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", ctx, targetLink, userLink, mock.Anything).Return(nil)
				m.On("GetSection", ctx, targetLink).Return(repositoryDto.FullSectionInfo{Position: 2}, nil)
				m.On("UpdateSection", ctx, expectedRepoDto).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error permission denied",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", ctx, targetLink, userLink, mock.Anything).Return(rbac.ErrActionDenied)
			},
			expectedError: rbac.ErrActionDenied,
		},
		{
			nameTest: "Error get section info fails",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", ctx, targetLink, userLink, mock.Anything).Return(nil)
				m.On("GetSection", ctx, targetLink).Return(repositoryDto.FullSectionInfo{}, errors.New("not found"))
			},
			expectedError: errors.New("SectionRepository.GetSectionInfo: not found"),
		},
		{
			nameTest: "Error update backlog section",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", ctx, targetLink, userLink, mock.Anything).Return(nil)
				m.On("GetSection", ctx, targetLink).Return(repositoryDto.FullSectionInfo{Position: 1}, nil)
			},
			expectedError: common.ErrCannotUpdateBacklog,
		},
		{
			nameTest: "Error update repository fail",
			mockBehavior: func(m *mockSectionRep.SectionRepository, r *MockRbacService) {
				r.On("CheckPermissionOnSection", ctx, targetLink, userLink, mock.Anything).Return(nil)
				m.On("GetSection", ctx, targetLink).Return(repositoryDto.FullSectionInfo{Position: 2}, nil)
				m.On("UpdateSection", ctx, expectedRepoDto).Return(errors.New("db error"))
			},
			expectedError: errors.New("SectionRepository.UpdateSection: db error"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepository := mockSectionRep.NewSectionRepository(t)
			mockRbac := new(MockRbacService)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepository, mockRbac)
			}

			service := NewService(mockRepository, mockRbac)
			err := service.UpdateSection(ctx, updateDto, userLink)

			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
