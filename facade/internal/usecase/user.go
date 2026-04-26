package usecase

import (
	"context"
	"fmt"

	authv1 "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/contracts/auth"
	userv1 "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/contracts/user"

	grpcclient "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/clients/grpc"
)

type UserUsecase interface {
	GetProfile(ctx context.Context, userLink string) (*userv1.ProfileResponse, error)
	Login(ctx context.Context, email, password string) (sessionID, userLink string, err error)
	Register(ctx context.Context, req *userv1.CreateRequest) (sessionID, userLink string, err error)
	Logout(ctx context.Context, sessionID string) error
}

type userUsecase struct {
	user *grpcclient.UserClient
	auth *grpcclient.AuthClient
}

func NewUserUsecase(user *grpcclient.UserClient, auth *grpcclient.AuthClient) UserUsecase {
	return &userUsecase{user: user, auth: auth}
}

func (u *userUsecase) GetProfile(ctx context.Context, userLink string) (*userv1.ProfileResponse, error) {
	resp, err := u.user.GetProfile(ctx, &userv1.UserLinkRequest{UserLink: userLink})
	if err != nil {
		return nil, fmt.Errorf("user.GetProfile: %w", err)
	}
	return resp, nil
}

func (u *userUsecase) Login(ctx context.Context, email, password string) (string, string, error) {
	userResp, err := u.user.GetUser(ctx, &userv1.GetUserRequest{Email: email, Password: password})
	if err != nil {
		return "", "", fmt.Errorf("user.GetUser: %w", err)
	}

	sessionResp, err := u.auth.CreateSession(ctx, &authv1.CreateSessionRequest{UserLink: userResp.UserLink})
	if err != nil {
		return "", "", fmt.Errorf("auth.CreateSession: %w", err)
	}

	return sessionResp.SessionId, userResp.UserLink, nil
}

func (u *userUsecase) Register(ctx context.Context, req *userv1.CreateRequest) (string, string, error) {
	userResp, err := u.user.CreateUser(ctx, req)
	if err != nil {
		return "", "", fmt.Errorf("user.CreateUser: %w", err)
	}

	sessionResp, err := u.auth.CreateSession(ctx, &authv1.CreateSessionRequest{UserLink: userResp.UserLink})
	if err != nil {
		return "", "", fmt.Errorf("auth.CreateSession: %w", err)
	}

	return sessionResp.SessionId, userResp.UserLink, nil
}

func (u *userUsecase) Logout(ctx context.Context, sessionID string) error {
	_, err := u.auth.DeleteSession(ctx, &authv1.DeleteSessionRequest{SessionId: sessionID})
	if err != nil {
		return fmt.Errorf("auth.DeleteSession: %w", err)
	}
	return nil
}
