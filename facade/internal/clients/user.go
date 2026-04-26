package clients

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/contracts/user"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type User struct {
	client pb.UserServiceClient
}

func NewUserClient(connection *grpc.ClientConn) *User {
	return &User{
		client: pb.NewUserServiceClient(connection),
	}
}

func convertUserGRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	switch st.Code() {
	case codes.AlreadyExists:
		return ErrUserAlreadyExists
	case codes.NotFound:
		msg := st.Message()
		switch {
		case strings.Contains(msg, "email"):
			return ErrEmailNotFound
		default:
			return ErrUserNotFound
		}
	case codes.InvalidArgument:
		msg := st.Message()
		switch {
		case strings.Contains(msg, "wrong"):
			return ErrWrongCredentials
		case strings.Contains(msg, "null"):
			return ErrNullInNotNullField
		default:
			return ErrInvalidInput
		}
	case codes.Unavailable:
		return ErrVKOAuthUnavailable
	default:
		return err
	}
}

func (u *User) GetProfile(ctx context.Context, userLink uuid.UUID) (domain.User, error) {
	req := &pb.UserLinkRequest{
		UserLink: userLink.String(),
	}

	resp, err := u.client.GetProfile(ctx, req)
	if err != nil {
		return domain.User{}, fmt.Errorf("client.GetProfile: %w", convertUserGRPCError(err))
	}

	convertedUserLink, err := uuid.Parse(resp.UserLink)
	if err != nil {
		return domain.User{}, common.ErrorParseLink
	}

	return domain.User{
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
		return fmt.Errorf("client.UpdateProfile: %w", convertUserGRPCError(err))
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
		return "", fmt.Errorf("client.UpdateAvatar: %w", convertUserGRPCError(err))
	}

	return avatarURL.AvatarUrl, nil
}

func (u *User) DeleteAvatar(ctx context.Context, userLink uuid.UUID) error {
	req := &pb.UserLinkRequest{
		UserLink: userLink.String(),
	}

	_, err := u.client.DeleteAvatar(ctx, req)
	if err != nil {
		return fmt.Errorf("client.DeleteAvatar: %w", convertUserGRPCError(err))
	}

	return nil
}

func (u *User) GetUser(ctx context.Context, entryUser domain.EntryUserInfo) (domain.User, error) {
	req := &pb.GetUserRequest{
		Email:    entryUser.Email,
		Password: entryUser.Password,
	}

	resp, err := u.client.GetUser(ctx, req)
	if err != nil {
		return domain.User{}, fmt.Errorf("client.GetUser: %w", convertUserGRPCError(err))
	}

	convertedUserLink, err := uuid.Parse(resp.UserLink)
	if err != nil {
		return domain.User{}, common.ErrorParseLink
	}

	return domain.User{
		UserLink:    convertedUserLink,
		Email:       resp.Email,
		DisplayName: resp.DisplayName,
		AvatarURL:   resp.Avatar,
	}, nil
}

func (u *User) CreateUser(ctx context.Context, infoUser domain.NewUser) (domain.User, error) {
	req := &pb.CreateRequest{
		DisplayName:      infoUser.DisplayName,
		Email:            infoUser.Email,
		Password:         infoUser.Password,
		RepeatedPassword: infoUser.RepeatedPassword,
	}

	resp, err := u.client.CreateUser(ctx, req)
	if err != nil {
		return domain.User{}, fmt.Errorf("client.CreateUser: %w", convertUserGRPCError(err))
	}

	convertedUserLink, err := uuid.Parse(resp.UserLink)
	if err != nil {
		return domain.User{}, common.ErrorParseLink
	}

	return domain.User{
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
		return uuid.Nil, fmt.Errorf("client.GetUserLink: %w", convertUserGRPCError(err))
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
		return fmt.Errorf("client.ResetPassword: %w", convertUserGRPCError(err))
	}

	return nil
}

func (u *User) ProcessUserWithVK(ctx context.Context, code string) (uuid.UUID, error) {
	req := &pb.ProcessUserVKRequest{
		Code: code,
	}

	resp, err := u.client.ProcessUserWithVK(ctx, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("client.ProcessUserWithVK: %w", convertUserGRPCError(err))
	}

	userLink, err := uuid.Parse(resp.UserLink)
	if err != nil {
		return uuid.Nil, common.ErrorParseLink
	}

	return userLink, nil
}
