package clients

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/board/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type Board struct {
	client pb.BoardServiceClient
}

func NewBoardClient(connection *grpc.ClientConn) *Board {
	return &Board{
		client: pb.NewBoardServiceClient(connection),
	}
}

func (b *Board) GetBoards(ctx context.Context, userLink uuid.UUID) ([]domain.BoardInfo, error)

func (b *Board) GetBoard(ctx context.Context, boardInfo domain.GetBoardRequest) (domain.BoardInfo, error)

func (b *Board) CreateBoard(ctx context.Context, boardInfo domain.CreateBoardRequest) (domain.BoardInfo, error)

func (b *Board) DeleteBoard(ctx context.Context, boardInfo domain.GetBoardRequest) error

func (b *Board) UpdateBoard(ctx context.Context, boardInfo domain.UpdateBoardRequest) error

func (b *Board) UploadBackground(ctx context.Context, backgroundInfo domain.UploadBackgroundRequest) (domain.UploadBackgroundResponse, error)

func (b *Board) GetMembers(ctx context.Context, membersInfo domain.GetMembersRequest) (domain.GetMembersResponse, error)
