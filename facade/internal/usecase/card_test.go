package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	fixedCardLink    = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	fixedSectionLink = uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	fixedCommentLink = uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	fixedSubtaskLink = uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	fixedUserLink    = uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	cardTestError    = errors.New("card client error")
)

type mockCardClient struct {
	mock.Mock
}

func (m *mockCardClient) GetCard(ctx context.Context, infoCard domain.GetCardRequest) (domain.CardInfo, error) {
	args := m.Called(ctx, infoCard)
	return args.Get(0).(domain.CardInfo), args.Error(1)
}

func (m *mockCardClient) DeleteCard(ctx context.Context, infoCard domain.DeleteCardRequest) error {
	args := m.Called(ctx, infoCard)
	return args.Error(0)
}

func (m *mockCardClient) UpdateCard(ctx context.Context, infoCard domain.UpdateCardRequest) error {
	args := m.Called(ctx, infoCard)
	return args.Error(0)
}

func (m *mockCardClient) ReorderCards(ctx context.Context, infoCard domain.ReorderCardsRequest) error {
	args := m.Called(ctx, infoCard)
	return args.Error(0)
}

func (m *mockCardClient) CreateCard(ctx context.Context, infoCard domain.CreateCardRequest) (domain.CreateCardResponse, error) {
	args := m.Called(ctx, infoCard)
	return args.Get(0).(domain.CreateCardResponse), args.Error(1)
}

func (m *mockCardClient) GetComments(ctx context.Context, infoComments domain.GetCommentsRequest) (domain.GetCommentsResponse, error) {
	args := m.Called(ctx, infoComments)
	return args.Get(0).(domain.GetCommentsResponse), args.Error(1)
}

func (m *mockCardClient) CreateComment(ctx context.Context, infoComment domain.CreateCommentRequest) (domain.CreateCommentResponse, error) {
	args := m.Called(ctx, infoComment)
	return args.Get(0).(domain.CreateCommentResponse), args.Error(1)
}

func (m *mockCardClient) DeleteComment(ctx context.Context, infoComment domain.DeleteCommentRequest) error {
	args := m.Called(ctx, infoComment)
	return args.Error(0)
}

func (m *mockCardClient) UpdateComment(ctx context.Context, infoComment domain.UpdateCommentRequest) error {
	args := m.Called(ctx, infoComment)
	return args.Error(0)
}

func (m *mockCardClient) CreateSubtask(ctx context.Context, infoSubtask domain.CreateSubtaskRequest) (domain.SubtaskInfo, error) {
	args := m.Called(ctx, infoSubtask)
	return args.Get(0).(domain.SubtaskInfo), args.Error(1)
}

func (m *mockCardClient) UpdateSubtask(ctx context.Context, infoSubtask domain.UpdateSubtaskRequest) error {
	args := m.Called(ctx, infoSubtask)
	return args.Error(0)
}

func (m *mockCardClient) DeleteSubtask(ctx context.Context, infoSubtask domain.DeleteSubtask) error {
	args := m.Called(ctx, infoSubtask)
	return args.Error(0)
}

func TestCardUsecase_GetCard(t *testing.T) {
	req := domain.GetCardRequest{UserLink: fixedUserLink, CardLink: fixedCardLink}
	expectedCard := domain.CardInfo{CardLink: fixedCardLink, Title: "Test Card"}

	tests := []struct {
		name         string
		mockBehavior func(m *mockCardClient)
		expectedCard domain.CardInfo
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockCardClient) {
				m.On("GetCard", mock.Anything, req).Return(expectedCard, nil)
			},
			expectedCard: expectedCard,
			expectError:  false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockCardClient) {
				m.On("GetCard", mock.Anything, req).Return(domain.CardInfo{}, cardTestError)
			},
			expectedCard: domain.CardInfo{},
			expectError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardClient)
			tc.mockBehavior(m)

			card, err := NewCard(m).GetCard(context.Background(), req)

			assert.Equal(t, tc.expectedCard, card)
			if tc.expectError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, cardTestError))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCardUsecase_DeleteCard(t *testing.T) {
	req := domain.DeleteCardRequest{UserLink: fixedUserLink, CardLink: fixedCardLink}

	tests := []struct {
		name         string
		mockBehavior func(m *mockCardClient)
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockCardClient) {
				m.On("DeleteCard", mock.Anything, req).Return(nil)
			},
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockCardClient) {
				m.On("DeleteCard", mock.Anything, req).Return(cardTestError)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardClient)
			tc.mockBehavior(m)

			err := NewCard(m).DeleteCard(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, cardTestError))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCardUsecase_UpdateCard(t *testing.T) {
	req := domain.UpdateCardRequest{
		UserLink:    fixedUserLink,
		CardLink:    fixedCardLink,
		Title:       "New Title",
		Description: "New Desc",
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mockCardClient)
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockCardClient) {
				m.On("UpdateCard", mock.Anything, req).Return(nil)
			},
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockCardClient) {
				m.On("UpdateCard", mock.Anything, req).Return(cardTestError)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardClient)
			tc.mockBehavior(m)

			err := NewCard(m).UpdateCard(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, cardTestError))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCardUsecase_ReorderCards(t *testing.T) {
	req := domain.ReorderCardsRequest{
		UserLink:    fixedUserLink,
		CardLink:    fixedCardLink,
		SectionLink: fixedSectionLink,
		Position:    2,
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mockCardClient)
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockCardClient) {
				m.On("ReorderCards", mock.Anything, req).Return(nil)
			},
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockCardClient) {
				m.On("ReorderCards", mock.Anything, req).Return(cardTestError)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardClient)
			tc.mockBehavior(m)

			err := NewCard(m).ReorderCards(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, cardTestError))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCardUsecase_CreateCard(t *testing.T) {
	req := domain.CreateCardRequest{
		UserLink:    fixedUserLink,
		SectionLink: fixedSectionLink,
		Title:       "New Card",
		Description: "Desc",
	}
	expectedResp := domain.CreateCardResponse{
		CardLink:    fixedCardLink,
		SectionLink: fixedSectionLink,
		Position:    1,
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mockCardClient)
		expectedResp domain.CreateCardResponse
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockCardClient) {
				m.On("CreateCard", mock.Anything, req).Return(expectedResp, nil)
			},
			expectedResp: expectedResp,
			expectError:  false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockCardClient) {
				m.On("CreateCard", mock.Anything, req).Return(domain.CreateCardResponse{}, cardTestError)
			},
			expectedResp: domain.CreateCardResponse{},
			expectError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardClient)
			tc.mockBehavior(m)

			resp, err := NewCard(m).CreateCard(context.Background(), req)

			assert.Equal(t, tc.expectedResp, resp)
			if tc.expectError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, cardTestError))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCardUsecase_GetComments(t *testing.T) {
	req := domain.GetCommentsRequest{UserLink: fixedUserLink, CardLink: fixedCardLink}
	expectedResp := domain.GetCommentsResponse{
		CommentsInfo: []domain.CommentInfo{
			{CommentLink: fixedCommentLink, AuthorLink: fixedUserLink, Text: "hello"},
		},
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mockCardClient)
		expectedResp domain.GetCommentsResponse
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockCardClient) {
				m.On("GetComments", mock.Anything, req).Return(expectedResp, nil)
			},
			expectedResp: expectedResp,
			expectError:  false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockCardClient) {
				m.On("GetComments", mock.Anything, req).Return(domain.GetCommentsResponse{}, cardTestError)
			},
			expectedResp: domain.GetCommentsResponse{},
			expectError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardClient)
			tc.mockBehavior(m)

			resp, err := NewCard(m).GetComments(context.Background(), req)

			assert.Equal(t, tc.expectedResp, resp)
			if tc.expectError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, cardTestError))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCardUsecase_CreateComment(t *testing.T) {
	req := domain.CreateCommentRequest{
		UserLink: fixedUserLink,
		CardLink: fixedCardLink,
		Text:     "comment text",
	}
	expectedResp := domain.CreateCommentResponse{CommentLink: fixedCommentLink}

	tests := []struct {
		name         string
		mockBehavior func(m *mockCardClient)
		expectedResp domain.CreateCommentResponse
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockCardClient) {
				m.On("CreateComment", mock.Anything, req).Return(expectedResp, nil)
			},
			expectedResp: expectedResp,
			expectError:  false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockCardClient) {
				m.On("CreateComment", mock.Anything, req).Return(domain.CreateCommentResponse{}, cardTestError)
			},
			expectedResp: domain.CreateCommentResponse{},
			expectError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardClient)
			tc.mockBehavior(m)

			resp, err := NewCard(m).CreateComment(context.Background(), req)

			assert.Equal(t, tc.expectedResp, resp)
			if tc.expectError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, cardTestError))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCardUsecase_DeleteComment(t *testing.T) {
	req := domain.DeleteCommentRequest{UserLink: fixedUserLink, CommentLink: fixedCommentLink}

	tests := []struct {
		name         string
		mockBehavior func(m *mockCardClient)
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockCardClient) {
				m.On("DeleteComment", mock.Anything, req).Return(nil)
			},
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockCardClient) {
				m.On("DeleteComment", mock.Anything, req).Return(cardTestError)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardClient)
			tc.mockBehavior(m)

			err := NewCard(m).DeleteComment(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, cardTestError))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCardUsecase_UpdateComment(t *testing.T) {
	req := domain.UpdateCommentRequest{
		UserLink:    fixedUserLink,
		CommentLink: fixedCommentLink,
		Text:        "updated text",
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mockCardClient)
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockCardClient) {
				m.On("UpdateComment", mock.Anything, req).Return(nil)
			},
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockCardClient) {
				m.On("UpdateComment", mock.Anything, req).Return(cardTestError)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardClient)
			tc.mockBehavior(m)

			err := NewCard(m).UpdateComment(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, cardTestError))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCardUsecase_CreateSubtask(t *testing.T) {
	req := domain.CreateSubtaskRequest{
		UserLink:    fixedUserLink,
		CardLink:    fixedCardLink,
		Description: "subtask desc",
	}
	expectedSubtask := domain.SubtaskInfo{
		SubtaskLink: fixedSubtaskLink,
		Description: "subtask desc",
		IsDone:      false,
		Position:    1,
	}

	tests := []struct {
		name            string
		mockBehavior    func(m *mockCardClient)
		expectedSubtask domain.SubtaskInfo
		expectError     bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockCardClient) {
				m.On("CreateSubtask", mock.Anything, req).Return(expectedSubtask, nil)
			},
			expectedSubtask: expectedSubtask,
			expectError:     false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockCardClient) {
				m.On("CreateSubtask", mock.Anything, req).Return(domain.SubtaskInfo{}, cardTestError)
			},
			expectedSubtask: domain.SubtaskInfo{},
			expectError:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardClient)
			tc.mockBehavior(m)

			subtask, err := NewCard(m).CreateSubtask(context.Background(), req)

			assert.Equal(t, tc.expectedSubtask, subtask)
			if tc.expectError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, cardTestError))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCardUsecase_UpdateSubtask(t *testing.T) {
	req := domain.UpdateSubtaskRequest{
		UserLink:    fixedUserLink,
		SubtaskLink: fixedSubtaskLink,
		IsDone:      true,
		Description: "updated desc",
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mockCardClient)
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockCardClient) {
				m.On("UpdateSubtask", mock.Anything, req).Return(nil)
			},
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockCardClient) {
				m.On("UpdateSubtask", mock.Anything, req).Return(cardTestError)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardClient)
			tc.mockBehavior(m)

			err := NewCard(m).UpdateSubtask(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, cardTestError))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCardUsecase_DeleteSubtask(t *testing.T) {
	req := domain.DeleteSubtask{UserLink: fixedUserLink, SubtaskLink: fixedSubtaskLink}

	tests := []struct {
		name         string
		mockBehavior func(m *mockCardClient)
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockCardClient) {
				m.On("DeleteSubtask", mock.Anything, req).Return(nil)
			},
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockCardClient) {
				m.On("DeleteSubtask", mock.Anything, req).Return(cardTestError)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := new(mockCardClient)
			tc.mockBehavior(m)

			err := NewCard(m).DeleteSubtask(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, cardTestError))
			} else {
				require.NoError(t, err)
			}
		})
	}
}
