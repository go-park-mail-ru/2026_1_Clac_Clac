package clients

import (
	"context"
	"fmt"

	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/auth/v1"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/common"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type Auth struct {
	client pb.AuthServiceClient
}

func NewAuthClient(connection *grpc.ClientConn) *Auth {
	return &Auth{
		client: pb.NewAuthServiceClient(connection),
	}
}

func (a *Auth) CheckSession(ctx context.Context, sessionID string) (uuid.UUID, error) {
	req := &pb.GetUserLinkRequest{
		SessionId: sessionID,
	}

	resp, err := a.client.GetUserLink(ctx, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("AuthClient.GetUserLink: %w", err)
	}

	userLink, err := uuid.Parse(resp.UserLink)
	if err != nil {
		return uuid.Nil, common.ErrParseLink
	}

	return userLink, nil
}
