package clients

import (
	"context"
	"fmt"

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

func (b *Board) CanView(ctx context.Context, userLink, boardLink uuid.UUID) error {
	req := &pb.CanViewRequest{
		UserLink:  userLink.String(),
		BoardLink: boardLink.String(),
	}

	_, err := b.client.CanView(ctx, req)
	if err != nil {
		return fmt.Errorf("BoardClient.CanView: %w", err)
	}

	return nil
}
