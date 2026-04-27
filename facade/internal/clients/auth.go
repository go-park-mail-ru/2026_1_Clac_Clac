package clients

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/contracts/auth"
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

func (a *Auth) CreateSession(ctx context.Context, userLink uuid.UUID) (string, error) {
	req := &pb.CreateSessionRequest{
		UserLink: userLink.String(),
	}

	resp, err := a.client.CreateSession(ctx, req)
	if err != nil {
		return "", fmt.Errorf("client.CreateSession: %w", convertGRPCError(err))
	}

	return resp.SessionId, nil
}

func (a *Auth) CheckSession(ctx context.Context, sessionID string) (uuid.UUID, error) {
	req := &pb.GetUserLinkRequest{
		SessionId: sessionID,
	}

	resp, err := a.client.GetUserLink(ctx, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("client.GetUserLink: %w", convertGRPCError(err))
	}

	userLink, err := uuid.Parse(resp.UserLink)
	if err != nil {
		return uuid.Nil, common.ErrorParseLink
	}

	return userLink, nil
}

func (a *Auth) DeleteSession(ctx context.Context, sessionID string) error {
	req := &pb.DeleteSessionRequest{
		SessionId: sessionID,
	}

	_, err := a.client.DeleteSession(ctx, req)
	if err != nil {
		return fmt.Errorf("client.DeleteSession: %w", convertGRPCError(err))
	}

	return nil
}

func (a *Auth) RefreshSession(ctx context.Context, sessionID string) error {
	req := &pb.ExtendSessionRequest{
		SessionId: sessionID,
	}

	_, err := a.client.ExtendSession(ctx, req)
	if err != nil {
		return fmt.Errorf("client.ExtendSession: %w", convertGRPCError(err))
	}

	return nil
}

func (a *Auth) ExchangeVKCode(ctx context.Context, code string) (accessToken string, email string, err error) {
	req := &pb.ExchangeVKCodeRequest{
		Code: code,
	}

	resp, err := a.client.ExchangeVKCode(ctx, req)
	if err != nil {
		return "", "", fmt.Errorf("client.ExchangeVKCode: %w", convertGRPCError(err))
	}

	return resp.AccessToken, resp.Email, nil
}
