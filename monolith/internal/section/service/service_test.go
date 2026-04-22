package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/common"
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/section/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/section/service/dto"
	mockRepo "github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/section/service/mock_section_rep"
	mockSectionRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/section/service/mock_section_rep"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServiceGetSectionInfo(t *testing.T) {
	ctx := context.Background()
	targetLink := common.FixedUserUuiD
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
		SectionName: "To Do",
		Position:    2,
		IsMandatory: true,
		Color:       "white",
		MaxTasks:    &maxTasks,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockRepo.SectionRepository)
		expectedError error
		expectedData  dto.FullSectionInfo
	}{
		{
			nameTest: "Success get section info",
			mockBehavior: func(m *mockRepo.SectionRepository) {
				m.On("GetSectionInfo", ctx, targetLink).Return(repoResult, nil)
			},
			expectedError: nil,
			expectedData:  expectedResult,
		},
		{
			nameTest: "Error from repository",
			mockBehavior: func(m *mockRepo.SectionRepository) {
				m.On("GetSectionInfo", ctx, targetLink).Return(repositoryDto.FullSectionInfo{}, errors.New("db disconnect"))
			},
			expectedError: errors.New("rep.GetSectionInfo: db disconnect"),
			expectedData:  dto.FullSectionInfo{},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepository := mockRepo.NewSectionRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepository)
			}

			service := NewService(mockRepository)
			result, err := service.GetSectionInfo(ctx, targetLink)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					assert.EqualError(t, err, test.expectedError.Error())
				}
			} else {
				if assert.NoError(t, err) {
					assert.Equal(t, test.expectedData, result)
				}
			}
		})
	}
}

func TestServiceCreateSection(t *testing.T) {
	ctx := context.Background()
	boardLink := common.FixedBoardUuiD
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
		mockBehavior  func(m *mockRepo.SectionRepository)
		expectedError error
	}{
		{
			nameTest: "Success create section",
			mockBehavior: func(m *mockRepo.SectionRepository) {
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
			nameTest: "Error from repository",
			mockBehavior: func(m *mockRepo.SectionRepository) {
				m.On("CreateSection", ctx, mock.Anything).Return(repositoryDto.FullSectionInfo{}, errors.New("insert fail"))
			},
			expectedError: errors.New("rep.CreateSection: insert fail"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepository := mockRepo.NewSectionRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepository)
			}

			service := NewService(mockRepository)
			result, err := service.CreateSection(ctx, inputDto)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					assert.EqualError(t, err, test.expectedError.Error())
				}
			} else {
				if assert.NoError(t, err) {
					assert.NotEqual(t, uuid.Nil, result.SectionLink)
					assert.Equal(t, repoResult.SectionName, result.SectionName)
					assert.Equal(t, repoResult.Position, result.Position)
					assert.Equal(t, repoResult.Color, result.Color)
					assert.Equal(t, repoResult.IsMandatory, result.IsMandatory)
					assert.Equal(t, repoResult.MaxTasks, result.MaxTasks)
				}
			}
		})
	}
}

func TestServiceDeleteSection(t *testing.T) {
	ctx := context.Background()
	targetLink := common.FixedUserUuiD

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockRepo.SectionRepository)
		expectedError error
	}{
		{
			nameTest: "Success delete section",
			mockBehavior: func(m *mockRepo.SectionRepository) {
				m.On("GetSectionInfo", ctx, targetLink).Return(repositoryDto.FullSectionInfo{Position: 2}, nil)
				m.On("DeleteSection", ctx, targetLink).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error get section info fails",
			mockBehavior: func(m *mockRepo.SectionRepository) {
				m.On("GetSectionInfo", ctx, targetLink).Return(repositoryDto.FullSectionInfo{}, errors.New("not found"))
			},
			expectedError: errors.New("rep.GetSectionInfo: not found"),
		},
		{
			nameTest: "Error delete section repository fail",
			mockBehavior: func(m *mockRepo.SectionRepository) {
				m.On("GetSectionInfo", ctx, targetLink).Return(repositoryDto.FullSectionInfo{Position: 2}, nil)
				m.On("DeleteSection", ctx, targetLink).Return(errors.New("db timeout"))
			},
			expectedError: errors.New("rep.DeleteSection: db timeout"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepository := mockRepo.NewSectionRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepository)
			}

			service := NewService(mockRepository)
			err := service.DeleteSection(ctx, targetLink)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrorDeleteBacklog) {
						assert.ErrorIs(t, err, common.ErrorDeleteBacklog)
					} else {
						assert.EqualError(t, err, test.expectedError.Error())
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServiceReorderSection(t *testing.T) {
	ctx := context.Background()
	boardLink := common.FixedBoardUuiD
	section1 := common.FixedUserUuiD
	section2 := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	tests := []struct {
		nameTest      string
		sectionLinks  []uuid.UUID
		mockBehavior  func(m *mockRepo.SectionRepository)
		expectedError error
	}{
		{
			nameTest:      "Success empty list",
			sectionLinks:  []uuid.UUID{},
			mockBehavior:  nil,
			expectedError: nil,
		},
		{
			nameTest:     "Success reorder",
			sectionLinks: []uuid.UUID{section1, section2},
			mockBehavior: func(m *mockRepo.SectionRepository) {
				m.On("ReorderSection", ctx, boardLink, []uuid.UUID{section1, section2}).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest:     "Error from repository",
			sectionLinks: []uuid.UUID{section1, section2},
			mockBehavior: func(m *mockRepo.SectionRepository) {
				m.On("ReorderSection", ctx, boardLink, []uuid.UUID{section1, section2}).Return(errors.New("deadlock"))
			},
			expectedError: errors.New("rep.ReorderSection: deadlock"),
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepository := mockRepo.NewSectionRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepository)
			}

			service := NewService(mockRepository)
			err := service.ReorderSection(ctx, boardLink, test.sectionLinks)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					assert.EqualError(t, err, test.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServiceUpdateSection(t *testing.T) {
	ctx := context.Background()
	targetLink := common.FixedSectionUuiD
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
		mockBehavior  func(m *mockRepo.SectionRepository)
		expectedError error
	}{
		{
			nameTest: "Success update section",
			mockBehavior: func(m *mockRepo.SectionRepository) {
				m.On("GetSectionInfo", ctx, targetLink).Return(repositoryDto.FullSectionInfo{Position: 2}, nil)
				m.On("UpdateSection", ctx, expectedRepoDto).Return(nil)
			},
			expectedError: nil,
		},
		{
			nameTest: "Error get section info fails",
			mockBehavior: func(m *mockRepo.SectionRepository) {
				m.On("GetSectionInfo", ctx, targetLink).Return(repositoryDto.FullSectionInfo{}, errors.New("not found"))
			},
			expectedError: errors.New("rep.GetSectionInfo: not found"),
		},
		{
			nameTest: "Error update backlog section",
			mockBehavior: func(m *mockRepo.SectionRepository) {
				m.On("GetSectionInfo", ctx, targetLink).Return(repositoryDto.FullSectionInfo{Position: 1}, nil)
			},
			expectedError: common.ErrorUpdateBacklog,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepository := mockRepo.NewSectionRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepository)
			}

			service := NewService(mockRepository)
			err := service.UpdateSection(ctx, updateDto)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					if errors.Is(test.expectedError, common.ErrorUpdateBacklog) {
						assert.ErrorIs(t, err, common.ErrorUpdateBacklog)
					} else {
						assert.EqualError(t, err, test.expectedError.Error())
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServiceGetAllSections(t *testing.T) {
	ctx := context.Background()
	boardLink := common.FixedBoardUuiD
	sectionLink := common.FixedUserUuiD
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
		mockBehavior  func(m *mockRepo.SectionRepository)
		expectedError error
		expectedData  []dto.FullSectionInfo
	}{
		{
			nameTest: "Success get all sections",
			mockBehavior: func(m *mockRepo.SectionRepository) {
				m.On("GetAllSections", ctx, boardLink).Return(repoSections, nil)
			},
			expectedError: nil,
			expectedData:  expectedResult,
		},
		{
			nameTest: "Success get empty slice",
			mockBehavior: func(m *mockRepo.SectionRepository) {
				m.On("GetAllSections", ctx, boardLink).Return([]repositoryDto.FullSectionInfo{}, nil)
			},
			expectedError: nil,
			expectedData:  nil,
		},
		{
			nameTest: "Error from repository",
			mockBehavior: func(m *mockRepo.SectionRepository) {
				m.On("GetAllSections", ctx, boardLink).Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("rep.GetAllSections: db error"),
			expectedData:  []dto.FullSectionInfo{},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRepository := mockRepo.NewSectionRepository(t)
			if test.mockBehavior != nil {
				test.mockBehavior(mockRepository)
			}

			service := NewService(mockRepository)
			result, err := service.GetAllSections(ctx, boardLink)

			if test.expectedError != nil {
				if assert.Error(t, err) {
					assert.EqualError(t, err, test.expectedError.Error())
				}
			} else {
				if assert.NoError(t, err) {
					assert.Equal(t, test.expectedData, result)
				}
			}
		})
	}
}

func TestGetCards(t *testing.T) {
	targetSectionLink := uuid.New()
	targetExecuterName := "John Doe"
	targetDeadLine := time.Now()

	repCards := []repositoryDto.Card{
		{
			CardLink:     uuid.New(),
			ExecuterName: &targetExecuterName,
			Title:        "Task 1",
			DeadLine:     &targetDeadLine,
		},
	}

	expectedCards := []dto.Card{
		{
			CardLink:     repCards[0].CardLink,
			ExecuterName: repCards[0].ExecuterName,
			Title:        repCards[0].Title,
			DeadLine:     repCards[0].DeadLine,
		},
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockSectionRep.SectionRepository)
		expectedError bool
		expectedRes   []dto.Card
	}{
		{
			nameTest: "Success get cards",
			mockBehavior: func(m *mockSectionRep.SectionRepository) {
				m.On("GetCards", mock.Anything, targetSectionLink).Return(repCards, nil)
			},
			expectedError: false,
			expectedRes:   expectedCards,
		},
		{
			nameTest: "Error from repository",
			mockBehavior: func(m *mockSectionRep.SectionRepository) {
				m.On("GetCards", mock.Anything, targetSectionLink).Return([]repositoryDto.Card{}, errors.New("db error"))
			},
			expectedError: true,
			expectedRes:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := mockSectionRep.NewSectionRepository(t)
			test.mockBehavior(mockRep)

			service := NewService(mockRep)
			res, err := service.GetCards(context.Background(), targetSectionLink)

			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedRes, res)
			}
		})
	}
}
