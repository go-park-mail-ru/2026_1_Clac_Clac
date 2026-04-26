package clients

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/contracts/user"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type User struct {
	client pb.UserServiceClient
}

func NewUserClient(connection *grpc.ClientConn) *User {
	return &User{
		client: pb.NewUserServiceClient(connection),
	}
}

func (u *User) GetProfile(ctx context.Context, userLink uuid.UUID) (domain.FullInfoUser, error) {
	req := &pb.UserLinkRequest{
		UserLink: userLink.String(),
	}

	resp, err := u.client.GetProfile(ctx, req)
	if err != nil {
		return domain.FullInfoUser{}, fmt.Errorf("client.GetProfile: %w", convertGRPCError(err))
	}

	convertedUserLink, err := uuid.Parse(resp.UserLink)
	if err != nil {
		return domain.FullInfoUser{}, common.ErrorParseLink
	}

	return domain.FullInfoUser{
		UserLink:    convertedUserLink,
		Email:       resp.Email,
		DisplayName: resp.DisplayName,
		Description: resp.Description,
		AvatarURL:   resp.AvatarUrl,
	}, nil
}

func (u *User) UpdateProfile(ctx context.Context, updatedInfo domain.UpdatedInfo) error {
	req := &pb.UpdateProfileRequest{
		UserLink:    updatedInfo.UserLink.String(),
		DisplayName: updatedInfo.DisplayName,
		Description: updatedInfo.Description,
	}

	_, err := u.client.UpdateProfile(ctx, req)
	if err != nil {
		return fmt.Errorf("client.UpdateProfile: %w", convertGRPCError(err))
	}

	return nil
}

func (u *User) UpdateAvatar(ctx context.Context, avatarInfo domain.AvatarInfo) (string, error) {
	req := &pb.UpdateAvatarRequest{
		UserLink:      avatarInfo.UserLink.String(),
		FileData:      avatarInfo.FileData,
		ContentType:   avatarInfo.ContentType,
		FileExtension: avatarInfo.FileExtension,
	}

	avatarURL, err := u.client.UpdateAvatar(ctx, req)
	if err != nil {
		return "", fmt.Errorf("client.UpdateAvatar: %w", convertGRPCError(err))
	}

	return avatarURL.AvatarUrl, nil
}

func (u *User) DeleteAvatar(ctx context.Context, userLink uuid.UUID) error {
	req := &pb.UserLinkRequest{
		UserLink: userLink.String(),
	}

	_, err := u.client.DeleteAvatar(ctx, req)
	if err != nil {
		return fmt.Errorf("client.DeleteAvatar: %w", convertGRPCError(err))
	}

	return nil
}

func (u *User) GetUser(ctx context.Context, entryUser domain.Credentials) (domain.FullInfoUser, error) {
	req := &pb.GetUserRequest{
		Email:    entryUser.Email,
		Password: entryUser.Password,
	}

	resp, err := u.client.GetUser(ctx, req)
	if err != nil {
		return domain.FullInfoUser{}, fmt.Errorf("client.GetUser: %w", convertGRPCError(err))
	}

	convertedUserLink, err := uuid.Parse(resp.UserLink)
	if err != nil {
		return domain.FullInfoUser{}, common.ErrorParseLink
	}

	return domain.FullInfoUser{
		UserLink:    convertedUserLink,
		Email:       resp.Email,
		DisplayName: resp.DisplayName,
		AvatarURL:   resp.Avatar,
	}, nil
}

func (u *User) CreateUser(ctx context.Context, infoUser domain.NewCredentialsUser) (domain.FullInfoUser, error) {
	req := &pb.CreateRequest{
		DisplayName: infoUser.DisplayName,
		Email:       infoUser.Email,
		Password:    infoUser.Password,
	}

	resp, err := u.client.CreateUser(ctx, req)
	if err != nil {
		return domain.FullInfoUser{}, fmt.Errorf("client.CreateUser: %w", convertGRPCError(err))
	}

	convertedUserLink, err := uuid.Parse(resp.UserLink)
	if err != nil {
		return domain.FullInfoUser{}, common.ErrorParseLink
	}

	return domain.FullInfoUser{
		UserLink:    convertedUserLink,
		Email:       resp.Email,
		DisplayName: resp.DisplayName,
		AvatarURL:   resp.Avatar,
	}, nil
}

func (u *User) GetUserLink(ctx context.Context, email string) (uuid.UUID, error) {
	req := &pb.GetUserLinkRequest{
		Email: email,
	}

	resp, err := u.client.GetUserLink(ctx, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("client.GetUserLink: %w", convertGRPCError(err))
	}

	userLink, err := uuid.Parse(resp.UserLink)
	if err != nil {
		return uuid.Nil, common.ErrorParseLink
	}

	return userLink, nil
}

func (u *User) RessetPassword(ctx context.Context, updatedPassword domain.UpdatedPassoword) error {
	req := &pb.ResetPasswordRequest{
		UserLink:         updatedPassword.UserLink.String(),
		Password:         updatedPassword.Password,
		RepeatedPassword: updatedPassword.RepeatedPassword,
	}

	_, err := u.client.ResetPassword(ctx, req)
	if err != nil {
		return fmt.Errorf("client.ResetPassword: %w", convertGRPCError(err))
	}

	return nil
}

func (u *User) ProcessUserWithVK(ctx context.Context, accessToken string, email string) (uuid.UUID, error) {
	req := &pb.ProcessUserVKRequest{
		AccessToken: accessToken,
		Email:       email,
	}

	resp, err := u.client.ProcessUserWithVK(ctx, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("client.ProcessUserWithVK: %w", convertGRPCError(err))
	}

	userLink, err := uuid.Parse(resp.UserLink)
	if err != nil {
		return uuid.Nil, common.ErrorParseLink
	}

	return userLink, nil
}
