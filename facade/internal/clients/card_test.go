package clients

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/card/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockCardServiceClient struct {
	mock.Mock
}

func (m *mockCardServiceClient) GetCard(ctx context.Context, in *pb.GetCardRequest, opts ...grpc.CallOption) (*pb.GetCardResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetCardResponse), args.Error(1)
}

func (m *mockCardServiceClient) DeleteCard(ctx context.Context, in *pb.DeleteCardRequest, opts ...grpc.CallOption) (*pb.DeleteCardResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.DeleteCardResponse), args.Error(1)
}

func (m *mockCardServiceClient) UpdateCard(ctx context.Context, in *pb.UpdateCardRequest, opts ...grpc.CallOption) (*pb.UpdateCardResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UpdateCardResponse), args.Error(1)
}

func (m *mockCardServiceClient) ReorderCards(ctx context.Context, in *pb.ReorderCardsRequest, opts ...grpc.CallOption) (*pb.ReorderCardsResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ReorderCardsResponse), args.Error(1)
}

func (m *mockCardServiceClient) CreateCard(ctx context.Context, in *pb.CreateCardRequest, opts ...grpc.CallOption) (*pb.CreateCardResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CreateCardResponse), args.Error(1)
}

func (m *mockCardServiceClient) GetComments(ctx context.Context, in *pb.GetCommentsRequest, opts ...grpc.CallOption) (*pb.GetCommentsResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetCommentsResponse), args.Error(1)
}

func (m *mockCardServiceClient) CreateComment(ctx context.Context, in *pb.CreateCommentRequest, opts ...grpc.CallOption) (*pb.CreateCommentResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CreateCommentResponse), args.Error(1)
}

func (m *mockCardServiceClient) DeleteComment(ctx context.Context, in *pb.DeleteCommentRequest, opts ...grpc.CallOption) (*pb.DeleteCommentResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.DeleteCommentResponse), args.Error(1)
}

func (m *mockCardServiceClient) UpdateComment(ctx context.Context, in *pb.UpdateCommentRequest, opts ...grpc.CallOption) (*pb.UpdateCommentResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UpdateCommentResponse), args.Error(1)
}

func (m *mockCardServiceClient) CreateSubtask(ctx context.Context, in *pb.CreateSubtaskRequest, opts ...grpc.CallOption) (*pb.CreateSubtaskResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CreateSubtaskResponse), args.Error(1)
}

func (m *mockCardServiceClient) UpdateSubtask(ctx context.Context, in *pb.UpdateSubtaskRequest, opts ...grpc.CallOption) (*pb.UpdateSubtaskResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UpdateSubtaskResponse), args.Error(1)
}

func (m *mockCardServiceClient) DeleteSubtask(ctx context.Context, in *pb.DeleteSubtaskRequest, opts ...grpc.CallOption) (*pb.DeleteSubtaskResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.DeleteSubtaskResponse), args.Error(1)
}

func (m *mockCardServiceClient) CreateAttachment(ctx context.Context, in *pb.CreateAttachmentRequest, opts ...grpc.CallOption) (*pb.CreateAttachmentResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.CreateAttachmentResponse), args.Error(1)
}

func (m *mockCardServiceClient) DeleteAttachment(ctx context.Context, in *pb.DeleteAttachmentRequest, opts ...grpc.CallOption) (*pb.DeleteAttachmentResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.DeleteAttachmentResponse), args.Error(1)
}

var (
	cardClientCardLink       = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	cardClientSectionLink    = uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	cardClientCommentLink    = uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	cardClientSubtaskLink    = uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	cardClientUserLink       = uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	cardClientAttachmentLink = uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
)

func TestCardClient_GetCard(t *testing.T) {
	ctx := context.Background()
	req := domain.GetCardRequest{UserLink: cardClientUserLink, CardLink: cardClientCardLink}

	tests := []struct {
		name        string
		mockResp    *pb.GetCardResponse
		mockErr     error
		expectedErr error
	}{
		{
			name: "success",
			mockResp: &pb.GetCardResponse{
				CardInfo: &pb.CardInfo{
					Title:       "Test",
					Description: "Desc",
					Subtasks:    []*pb.SubtaskInfo{},
				},
			},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "grpc error",
			mockResp:    nil,
			mockErr:     status.Error(codes.NotFound, "card not found"),
			expectedErr: common.ErrorCardNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockCardServiceClient)
			mc.On("GetCard", ctx, &pb.GetCardRequest{
				UserLink: cardClientUserLink.String(),
				CardLink: cardClientCardLink.String(),
			}).Return(tt.mockResp, tt.mockErr)

			c := &Card{client: mc}
			_, err := c.GetCard(ctx, req)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCardClient_DeleteCard(t *testing.T) {
	ctx := context.Background()
	req := domain.DeleteCardRequest{UserLink: cardClientUserLink, CardLink: cardClientCardLink}

	tests := []struct {
		name        string
		mockResp    *pb.DeleteCardResponse
		mockErr     error
		expectedErr error
	}{
		{
			name:        "success",
			mockResp:    &pb.DeleteCardResponse{},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "grpc error",
			mockResp:    nil,
			mockErr:     status.Error(codes.NotFound, "card not found"),
			expectedErr: common.ErrorCardNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockCardServiceClient)
			mc.On("DeleteCard", ctx, &pb.DeleteCardRequest{
				UserLink: cardClientUserLink.String(),
				CardLink: cardClientCardLink.String(),
			}).Return(tt.mockResp, tt.mockErr)

			c := &Card{client: mc}
			err := c.DeleteCard(ctx, req)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCardClient_UpdateCard(t *testing.T) {
	ctx := context.Background()
	req := domain.UpdateCardRequest{
		UserLink:    cardClientUserLink,
		CardLink:    cardClientCardLink,
		Title:       "New Title",
		Description: "New Desc",
	}

	tests := []struct {
		name        string
		mockResp    *pb.UpdateCardResponse
		mockErr     error
		expectedErr error
	}{
		{
			name:        "success",
			mockResp:    &pb.UpdateCardResponse{},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "grpc error",
			mockResp:    nil,
			mockErr:     status.Error(codes.PermissionDenied, "permission denied"),
			expectedErr: common.ErrorPermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockCardServiceClient)
			mc.On("UpdateCard", ctx, &pb.UpdateCardRequest{
				UserLink:    cardClientUserLink.String(),
				CardLink:    cardClientCardLink.String(),
				Title:       "New Title",
				Description: "New Desc",
			}).Return(tt.mockResp, tt.mockErr)

			c := &Card{client: mc}
			err := c.UpdateCard(ctx, req)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCardClient_ReorderCards(t *testing.T) {
	ctx := context.Background()
	req := domain.ReorderCardsRequest{
		UserLink:    cardClientUserLink,
		CardLink:    cardClientCardLink,
		SectionLink: cardClientSectionLink,
		Position:    2,
	}

	tests := []struct {
		name        string
		mockResp    *pb.ReorderCardsResponse
		mockErr     error
		expectedErr error
	}{
		{
			name:        "success",
			mockResp:    &pb.ReorderCardsResponse{},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "grpc error",
			mockResp:    nil,
			mockErr:     status.Error(codes.NotFound, "section not found"),
			expectedErr: common.ErrorSectionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockCardServiceClient)
			mc.On("ReorderCards", ctx, &pb.ReorderCardsRequest{
				UserLink:    cardClientUserLink.String(),
				CardLink:    cardClientCardLink.String(),
				SectionLink: cardClientSectionLink.String(),
				Position:    2,
			}).Return(tt.mockResp, tt.mockErr)

			c := &Card{client: mc}
			err := c.ReorderCards(ctx, req)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCardClient_CreateCard(t *testing.T) {
	ctx := context.Background()
	req := domain.CreateCardRequest{
		UserLink:    cardClientUserLink,
		SectionLink: cardClientSectionLink,
		Title:       "New Card",
		Description: "Desc",
	}

	tests := []struct {
		name        string
		mockResp    *pb.CreateCardResponse
		mockErr     error
		expectedErr error
	}{
		{
			name: "success",
			mockResp: &pb.CreateCardResponse{
				CardLink:    cardClientCardLink.String(),
				SectionLink: cardClientSectionLink.String(),
				Position:    1,
			},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "grpc error",
			mockResp:    nil,
			mockErr:     status.Error(codes.InvalidArgument, "task limit reached"),
			expectedErr: common.ErrorTaskLimitReached,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockCardServiceClient)
			mc.On("CreateCard", ctx, &pb.CreateCardRequest{
				UserLink:    cardClientUserLink.String(),
				SectionLink: cardClientSectionLink.String(),
				Title:       "New Card",
				Description: "Desc",
			}).Return(tt.mockResp, tt.mockErr)

			c := &Card{client: mc}
			_, err := c.CreateCard(ctx, req)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCardClient_GetComments(t *testing.T) {
	ctx := context.Background()
	req := domain.GetCommentsRequest{UserLink: cardClientUserLink, CardLink: cardClientCardLink}

	tests := []struct {
		name        string
		mockResp    *pb.GetCommentsResponse
		mockErr     error
		expectedErr error
	}{
		{
			name: "success",
			mockResp: &pb.GetCommentsResponse{
				CommentsInfo: []*pb.CommentInfo{
					{
						CommentLink: cardClientCommentLink.String(),
						AuthorLink:  cardClientUserLink.String(),
						Text:        "hello",
					},
				},
			},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "grpc error",
			mockResp:    nil,
			mockErr:     status.Error(codes.NotFound, "card not found"),
			expectedErr: common.ErrorCardNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockCardServiceClient)
			mc.On("GetComments", ctx, &pb.GetCommentsRequest{
				UserLink: cardClientUserLink.String(),
				CardLink: cardClientCardLink.String(),
			}).Return(tt.mockResp, tt.mockErr)

			c := &Card{client: mc}
			_, err := c.GetComments(ctx, req)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCardClient_CreateComment(t *testing.T) {
	ctx := context.Background()
	req := domain.CreateCommentRequest{
		UserLink: cardClientUserLink,
		CardLink: cardClientCardLink,
		Text:     "test comment",
	}

	tests := []struct {
		name        string
		mockResp    *pb.CreateCommentResponse
		mockErr     error
		expectedErr error
	}{
		{
			name:        "success",
			mockResp:    &pb.CreateCommentResponse{CommentLink: cardClientCommentLink.String()},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "grpc error",
			mockResp:    nil,
			mockErr:     status.Error(codes.PermissionDenied, "permission denied"),
			expectedErr: common.ErrorPermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockCardServiceClient)
			mc.On("CreateComment", ctx, &pb.CreateCommentRequest{
				UserLink: cardClientUserLink.String(),
				CardLink: cardClientCardLink.String(),
				Text:     "test comment",
			}).Return(tt.mockResp, tt.mockErr)

			c := &Card{client: mc}
			_, err := c.CreateComment(ctx, req)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCardClient_DeleteComment(t *testing.T) {
	ctx := context.Background()
	req := domain.DeleteCommentRequest{UserLink: cardClientUserLink, CommentLink: cardClientCommentLink}

	tests := []struct {
		name        string
		mockResp    *pb.DeleteCommentResponse
		mockErr     error
		expectedErr error
	}{
		{
			name:        "success",
			mockResp:    &pb.DeleteCommentResponse{},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "grpc error",
			mockResp:    nil,
			mockErr:     status.Error(codes.NotFound, "comment not found"),
			expectedErr: common.ErrorCommentNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockCardServiceClient)
			mc.On("DeleteComment", ctx, &pb.DeleteCommentRequest{
				UserLink:    cardClientUserLink.String(),
				CommentLink: cardClientCommentLink.String(),
			}).Return(tt.mockResp, tt.mockErr)

			c := &Card{client: mc}
			err := c.DeleteComment(ctx, req)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCardClient_UpdateComment(t *testing.T) {
	ctx := context.Background()
	req := domain.UpdateCommentRequest{
		UserLink:    cardClientUserLink,
		CommentLink: cardClientCommentLink,
		Text:        "updated",
	}

	tests := []struct {
		name        string
		mockResp    *pb.UpdateCommentResponse
		mockErr     error
		expectedErr error
	}{
		{
			name:        "success",
			mockResp:    &pb.UpdateCommentResponse{},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "grpc error",
			mockResp:    nil,
			mockErr:     status.Error(codes.NotFound, "comment not found"),
			expectedErr: common.ErrorCommentNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockCardServiceClient)
			mc.On("UpdateComment", ctx, &pb.UpdateCommentRequest{
				UserLink:    cardClientUserLink.String(),
				CommentLink: cardClientCommentLink.String(),
				Text:        "updated",
			}).Return(tt.mockResp, tt.mockErr)

			c := &Card{client: mc}
			err := c.UpdateComment(ctx, req)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCardClient_CreateSubtask(t *testing.T) {
	ctx := context.Background()
	req := domain.CreateSubtaskRequest{
		UserLink:    cardClientUserLink,
		CardLink:    cardClientCardLink,
		Description: "subtask desc",
	}

	tests := []struct {
		name        string
		mockResp    *pb.CreateSubtaskResponse
		mockErr     error
		expectedErr error
	}{
		{
			name: "success",
			mockResp: &pb.CreateSubtaskResponse{
				SubtaskLink: cardClientSubtaskLink.String(),
				Description: "subtask desc",
				IsDone:      false,
				Position:    1,
			},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "grpc error",
			mockResp:    nil,
			mockErr:     status.Error(codes.AlreadyExists, "already exists"),
			expectedErr: common.ErrorCardAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockCardServiceClient)
			mc.On("CreateSubtask", ctx, &pb.CreateSubtaskRequest{
				UserLink:    cardClientUserLink.String(),
				CardLink:    cardClientCardLink.String(),
				Description: "subtask desc",
			}).Return(tt.mockResp, tt.mockErr)

			c := &Card{client: mc}
			_, err := c.CreateSubtask(ctx, req)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCardClient_UpdateSubtask(t *testing.T) {
	ctx := context.Background()
	req := domain.UpdateSubtaskRequest{
		UserLink:    cardClientUserLink,
		SubtaskLink: cardClientSubtaskLink,
		IsDone:      true,
		Description: "updated",
	}

	tests := []struct {
		name        string
		mockResp    *pb.UpdateSubtaskResponse
		mockErr     error
		expectedErr error
	}{
		{
			name:        "success",
			mockResp:    &pb.UpdateSubtaskResponse{},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "grpc error",
			mockResp:    nil,
			mockErr:     status.Error(codes.NotFound, "sub task not found"),
			expectedErr: common.ErrorSubtaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockCardServiceClient)
			mc.On("UpdateSubtask", ctx, &pb.UpdateSubtaskRequest{
				UserLink:    cardClientUserLink.String(),
				SubtaskLink: cardClientSubtaskLink.String(),
				IsDone:      true,
				Description: "updated",
			}).Return(tt.mockResp, tt.mockErr)

			c := &Card{client: mc}
			err := c.UpdateSubtask(ctx, req)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCardClient_DeleteSubtask(t *testing.T) {
	ctx := context.Background()
	req := domain.DeleteSubtaskRequest{UserLink: cardClientUserLink, SubtaskLink: cardClientSubtaskLink}

	tests := []struct {
		name        string
		mockResp    *pb.DeleteSubtaskResponse
		mockErr     error
		expectedErr error
	}{
		{
			name:        "success",
			mockResp:    &pb.DeleteSubtaskResponse{},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "grpc error",
			mockResp:    nil,
			mockErr:     status.Error(codes.NotFound, "sub task not found"),
			expectedErr: common.ErrorSubtaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockCardServiceClient)
			mc.On("DeleteSubtask", ctx, &pb.DeleteSubtaskRequest{
				UserLink:    cardClientUserLink.String(),
				SubtaskLink: cardClientSubtaskLink.String(),
			}).Return(tt.mockResp, tt.mockErr)

			c := &Card{client: mc}
			err := c.DeleteSubtask(ctx, req)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConvertCardGRPCError(t *testing.T) {
	tests := []struct {
		name        string
		inputErr    error
		expectedErr error
	}{
		{
			name:        "NotFound card not found",
			inputErr:    status.Error(codes.NotFound, "card not found"),
			expectedErr: common.ErrorCardNotFound,
		},
		{
			name:        "NotFound section not found",
			inputErr:    status.Error(codes.NotFound, "section not found"),
			expectedErr: common.ErrorSectionNotFound,
		},
		{
			name:        "NotFound comment not found",
			inputErr:    status.Error(codes.NotFound, "comment not found"),
			expectedErr: common.ErrorCommentNotFound,
		},
		{
			name:        "NotFound sub task not found",
			inputErr:    status.Error(codes.NotFound, "sub task not found"),
			expectedErr: common.ErrorSubtaskNotFound,
		},
		{
			name:        "PermissionDenied",
			inputErr:    status.Error(codes.PermissionDenied, "access denied"),
			expectedErr: common.ErrorPermissionDenied,
		},
		{
			name:        "AlreadyExists",
			inputErr:    status.Error(codes.AlreadyExists, "card already exists"),
			expectedErr: common.ErrorCardAlreadyExists,
		},
		{
			name:        "InvalidArgument task limit reached",
			inputErr:    status.Error(codes.InvalidArgument, "task limit reached"),
			expectedErr: common.ErrorTaskLimitReached,
		},
		{
			name:        "InvalidArgument other",
			inputErr:    status.Error(codes.InvalidArgument, "bad input"),
			expectedErr: common.ErrorInvalidInput,
		},
		{
			name:        "non-grpc error returned as-is",
			inputErr:    common.ErrorCardNotFound,
			expectedErr: common.ErrorCardNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertCardGRPCError(tt.inputErr)
			assert.ErrorIs(t, result, tt.expectedErr)
		})
	}
}

func TestCardClient_CreateAttachment(t *testing.T) {
	ctx := context.Background()
	req := domain.CreateAttachmentRequest{
		UserLink:   cardClientUserLink,
		TaskLink:   cardClientCardLink,
		Attachment: bytes.NewReader([]byte("fake image data")),
		Filename:   "photo.png",
	}

	tests := []struct {
		name        string
		mockResp    *pb.CreateAttachmentResponse
		mockErr     error
		expectedErr error
	}{
		{
			name: "success",
			mockResp: &pb.CreateAttachmentResponse{
				AttachmentLink: cardClientAttachmentLink.String(),
				Path:           "https://s3.example.com/file.png",
				Position:       1,
				Name:           "photo.png",
			},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "grpc permission denied",
			mockResp:    nil,
			mockErr:     status.Error(codes.PermissionDenied, "permission denied"),
			expectedErr: common.ErrorPermissionDenied,
		},
		{
			name:        "grpc invalid argument",
			mockResp:    nil,
			mockErr:     status.Error(codes.InvalidArgument, "missing required field"),
			expectedErr: common.ErrorInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockCardServiceClient)
			mc.On("CreateAttachment", ctx, mock.MatchedBy(func(r *pb.CreateAttachmentRequest) bool {
				return r.UserLink == cardClientUserLink.String() && r.TaskLink == cardClientCardLink.String()
			})).Return(tt.mockResp, tt.mockErr)

			c := &Card{client: mc}
			_, err := c.CreateAttachment(ctx, req)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCardClient_DeleteAttachment(t *testing.T) {
	ctx := context.Background()
	req := domain.DeleteAttachmentRequest{
		UserLink:       cardClientUserLink,
		AttachmentLink: cardClientAttachmentLink,
	}

	tests := []struct {
		name        string
		mockResp    *pb.DeleteAttachmentResponse
		mockErr     error
		expectedErr error
	}{
		{
			name:        "success",
			mockResp:    &pb.DeleteAttachmentResponse{},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "attachment not found",
			mockResp:    nil,
			mockErr:     status.Error(codes.NotFound, "attachment not found"),
			expectedErr: common.ErrorAttachmentNotFound,
		},
		{
			name:        "permission denied",
			mockResp:    nil,
			mockErr:     status.Error(codes.PermissionDenied, "permission denied"),
			expectedErr: common.ErrorPermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockCardServiceClient)
			mc.On("DeleteAttachment", ctx, &pb.DeleteAttachmentRequest{
				UserLink:       cardClientUserLink.String(),
				AttachmentLink: cardClientAttachmentLink.String(),
			}).Return(tt.mockResp, tt.mockErr)

			c := &Card{client: mc}
			err := c.DeleteAttachment(ctx, req)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
