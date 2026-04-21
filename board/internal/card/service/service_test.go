package service

import (
	"context"
	"errors"
	"testing"
	"time"

	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/service/dto"
	mockCardRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/service/mock_card_rep"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCard(t *testing.T) {
	targetCardLink := uuid.New()
	targetExecuterName := "John Doe"
	targetDataDeadLine := time.Now()

	repResponse := repositoryDto.InfoCard{
		Title:        "Title",
		Description:  "Desc",
		NameExecuter: &targetExecuterName,
		DataDeadLine: &targetDataDeadLine,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockCardRep.CardRepository)
		expectedError bool
		expectedRes   dto.InfoCard
	}{
		{
			nameTest: "Success get card",
			mockBehavior: func(m *mockCardRep.CardRepository) {
				m.On("GetCard", mock.Anything, targetCardLink).Return(repResponse, nil)
			},
			expectedError: false,
			expectedRes: dto.InfoCard{
				Title:        "Title",
				Description:  "Desc",
				NameExecuter: &targetExecuterName,
				DataDeadLine: &targetDataDeadLine,
			},
		},
		{
			nameTest: "Error from repository",
			mockBehavior: func(m *mockCardRep.CardRepository) {
				m.On("GetCard", mock.Anything, targetCardLink).Return(repositoryDto.InfoCard{}, errors.New("db error"))
			},
			expectedError: true,
			expectedRes:   dto.InfoCard{},
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := mockCardRep.NewCardRepository(t)
			test.mockBehavior(mockRep)

			service := NewService(mockRep)
			res, err := service.GetCard(context.Background(), targetCardLink)

			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedRes, res)
			}
		})
	}
}

func TestDeleteCard(t *testing.T) {
	targetCardLink := uuid.New()

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockCardRep.CardRepository)
		expectedError bool
	}{
		{
			nameTest: "Success delete card",
			mockBehavior: func(m *mockCardRep.CardRepository) {
				m.On("DeleteCard", mock.Anything, targetCardLink).Return(nil)
			},
			expectedError: false,
		},
		{
			nameTest: "Error from repository",
			mockBehavior: func(m *mockCardRep.CardRepository) {
				m.On("DeleteCard", mock.Anything, targetCardLink).Return(errors.New("db error"))
			},
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := mockCardRep.NewCardRepository(t)
			test.mockBehavior(mockRep)

			service := NewService(mockRep)
			err := service.DeleteCard(context.Background(), targetCardLink)

			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateCardDetails(t *testing.T) {
	targetCardLink := uuid.New()
	targetExecuterLink := uuid.New()
	targetDataDeadLine := time.Now()

	updateDto := dto.UpdatingCardDetails{
		LinkCard:     targetCardLink,
		Title:        "Upd Title",
		Description:  "Upd Desc",
		LinkExecuter: &targetExecuterLink,
		DataDeadLine: &targetDataDeadLine,
	}

	repUpdateDto := repositoryDto.UpdatingCardDetails{
		LinkCard:     targetCardLink,
		Title:        "Upd Title",
		Description:  "Upd Desc",
		LinkExecuter: &targetExecuterLink,
		DataDeadLine: &targetDataDeadLine,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockCardRep.CardRepository)
		expectedError bool
	}{
		{
			nameTest: "Success update card",
			mockBehavior: func(m *mockCardRep.CardRepository) {
				m.On("UpdateCardDetails", mock.Anything, repUpdateDto).Return(nil)
			},
			expectedError: false,
		},
		{
			nameTest: "Error from repository",
			mockBehavior: func(m *mockCardRep.CardRepository) {
				m.On("UpdateCardDetails", mock.Anything, repUpdateDto).Return(errors.New("db error"))
			},
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := mockCardRep.NewCardRepository(t)
			test.mockBehavior(mockRep)

			service := NewService(mockRep)
			err := service.UpdateCardDetails(context.Background(), updateDto)

			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReordredCard(t *testing.T) {
	targetCardLink := uuid.New()
	targetSectionLink := uuid.New()

	placeDto := dto.PlaceCard{
		LinkCard:    targetCardLink,
		LinkSection: targetSectionLink,
		Position:    2,
	}

	repPlaceDto := repositoryDto.PlaceCard{
		LinkCard:    targetCardLink,
		LinkSection: targetSectionLink,
		Position:    2,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockCardRep.CardRepository)
		expectedError bool
	}{
		{
			nameTest: "Success reorder card",
			mockBehavior: func(m *mockCardRep.CardRepository) {
				m.On("ReorderCard", mock.Anything, repPlaceDto).Return(nil)
			},
			expectedError: false,
		},
		{
			nameTest: "Error from repository",
			mockBehavior: func(m *mockCardRep.CardRepository) {
				m.On("ReorderCard", mock.Anything, repPlaceDto).Return(errors.New("db error"))
			},
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := mockCardRep.NewCardRepository(t)
			test.mockBehavior(mockRep)

			service := NewService(mockRep)
			err := service.ReorderCard(context.Background(), placeDto)

			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateCard(t *testing.T) {
	targetAuthorLink := uuid.New()
	targetSectionLink := uuid.New()
	targetExecuterLink := uuid.New()
	targetDataDeadLine := time.Now()

	newCardDto := dto.NewCard{
		LinkAuthor:   targetAuthorLink,
		LinkSection:  targetSectionLink,
		Title:        "Title",
		Description:  "Desc",
		LinkExecuter: &targetExecuterLink,
		DataDeadLine: &targetDataDeadLine,
	}

	tests := []struct {
		nameTest      string
		mockBehavior  func(m *mockCardRep.CardRepository)
		expectedError bool
	}{
		{
			nameTest: "Success create card",
			mockBehavior: func(m *mockCardRep.CardRepository) {
				m.On("CreateCard", mock.Anything, mock.AnythingOfType("dto.NewCard")).Return(5, nil)
			},
			expectedError: false,
		},
		{
			nameTest: "Error from repository",
			mockBehavior: func(m *mockCardRep.CardRepository) {
				m.On("CreateCard", mock.Anything, mock.AnythingOfType("dto.NewCard")).Return(-1, errors.New("db error"))
			},
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			mockRep := mockCardRep.NewCardRepository(t)
			test.mockBehavior(mockRep)

			service := NewService(mockRep)
			res, err := service.CreateCard(context.Background(), newCardDto)

			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 5, res.Position)
				assert.Equal(t, targetSectionLink, res.LinkSection)
				assert.NotEqual(t, uuid.Nil, res.LinkCard)
			}
		})
	}
}
