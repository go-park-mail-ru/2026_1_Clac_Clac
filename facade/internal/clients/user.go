package clients

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/user/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

const (
	userChunkSize = 1024 * 1024
)

var avatarBufferPool = sync.Pool{
	New: func() any {
		buffer := make([]byte, userChunkSize)
		return &buffer
	},
}

type ConfigUser struct {
	MaxUserAvatarBytesSize int
}

type User struct {
	client pb.UserServiceClient
	cfg    ConfigUser
}

func NewUserClient(connection *grpc.ClientConn, cfg ConfigUser) *User {
	return &User{
		client: pb.NewUserServiceClient(connection),
		cfg:    cfg,
	}
}

func (u *User) GetProfile(ctx context.Context, userLink uuid.UUID) (domain.FullInfoUser, error) {
	req := &pb.UserLinkRequest{
		UserLink: userLink.String(),
	}

	resp, err := u.client.GetProfile(ctx, req)
	if err != nil {
		return domain.FullInfoUser{}, fmt.Errorf("UserClient.GetProfile: %w", convertGRPCError(err))
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

func (u *User) GetProfiles(ctx context.Context, links []uuid.UUID) ([]domain.FullInfoUser, error) {
	rawLinks := make([]string, 0, len(links))
	for _, link := range links {
		rawLinks = append(rawLinks, link.String())
	}

	req := &pb.GetProfilesRequest{
		UserLinks: rawLinks,
	}

	resp, err := u.client.GetProfiles(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("UserClient.GetProfiles: %w", convertGRPCError(err))
	}

	profiles := make([]domain.FullInfoUser, 0, len(resp.Profiles))
	for _, p := range resp.Profiles {
		convertedUserLink, err := uuid.Parse(p.UserLink)
		if err != nil {
			return nil, common.ErrorParseLink
		}

		profiles = append(profiles, domain.FullInfoUser{
			UserLink:    convertedUserLink,
			Email:       p.Email,
			DisplayName: p.DisplayName,
			Description: p.Description,
			AvatarURL:   p.AvatarUrl,
		})
	}

	return profiles, nil
}

func (u *User) UpdateProfile(ctx context.Context, updatedInfo domain.UpdatedInfo) error {
	req := &pb.UpdateProfileRequest{
		UserLink:    updatedInfo.UserLink.String(),
		DisplayName: updatedInfo.DisplayName,
		Description: updatedInfo.Description,
	}

	_, err := u.client.UpdateProfile(ctx, req)
	if err != nil {
		return fmt.Errorf("UserClient.UpdateProfile: %w", convertGRPCError(err))
	}

	return nil
}

func (u *User) UpdateAvatar(ctx context.Context, avatarInfo domain.AvatarInfo) (string, error) {
	stream, err := u.client.UpdateAvatar(ctx)
	if err != nil {
		return "", fmt.Errorf("can not open stream client.UpdateAvatar: %w", err)
	}

	req := &pb.UpdateAvatarRequest{
		Request: &pb.UpdateAvatarRequest_Metadata{
			Metadata: &pb.MetadataUpdateAvatar{
				UserLink:      avatarInfo.UserLink.String(),
				ContentType:   avatarInfo.ContentType,
				FileExtension: avatarInfo.FileExtension,
			},
		},
	}

	if err := stream.Send(req); err != nil {
		return "", fmt.Errorf("metadata avatar stream.Send: %w", err)
	}

	bufPtr := avatarBufferPool.Get().(*[]byte)
	buffer := *bufPtr

	defer avatarBufferPool.Put(bufPtr)

	for {
		n, err := avatarInfo.FileData.Read(buffer)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("avatar.Read: %w", err)
		}

		if n > 0 {
			chunkReq := &pb.UpdateAvatarRequest{
				Request: &pb.UpdateAvatarRequest_FileData{
					FileData: buffer[:n],
				},
			}

			if err := stream.Send(chunkReq); err != nil {
				return "", fmt.Errorf("stream.Send: %w", err)
			}
		}

		if err == io.EOF {
			break
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return "", fmt.Errorf("UserClient.UpdateAvatar: %w", convertGRPCError(err))
	}

	return res.AvatarUrl, nil
}

func (u *User) DeleteAvatar(ctx context.Context, userLink uuid.UUID) error {
	req := &pb.UserLinkRequest{
		UserLink: userLink.String(),
	}

	_, err := u.client.DeleteAvatar(ctx, req)
	if err != nil {
		return fmt.Errorf("UserClient.DeleteAvatar: %w", convertGRPCError(err))
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
		return domain.FullInfoUser{}, fmt.Errorf("UserClient.GetUser: %w", convertGRPCError(err))
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
		return domain.FullInfoUser{}, fmt.Errorf("UserClient.CreateUser: %w", convertGRPCError(err))
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
		return uuid.Nil, fmt.Errorf("UserClient.GetUserLink: %w", convertGRPCError(err))
	}

	userLink, err := uuid.Parse(resp.UserLink)
	if err != nil {
		return uuid.Nil, common.ErrorParseLink
	}

	return userLink, nil
}

func (u *User) ResetPassword(ctx context.Context, updatedPassword domain.UpdatedPassword) error {
	req := &pb.ResetPasswordRequest{
		UserLink: updatedPassword.UserLink.String(),
		Password: updatedPassword.Password,
	}

	_, err := u.client.ResetPassword(ctx, req)
	if err != nil {
		return fmt.Errorf("UserClient.ResetPassword: %w", convertGRPCError(err))
	}

	return nil
}

func (u *User) ProcessUserWithVK(ctx context.Context, code, codeVerifier, state, deviceID string) (uuid.UUID, error) {
	req := &pb.ProcessUserVKRequest{
		Code:         code,
		CodeVerifier: codeVerifier,
		State:        state,
		DeviceId:     deviceID,
	}

	resp, err := u.client.ProcessUserWithVK(ctx, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("UserClient.ProcessUserWithVK: %w", convertGRPCError(err))
	}

	userLink, err := uuid.Parse(resp.UserLink)
	if err != nil {
		return uuid.Nil, common.ErrorParseLink
	}

	return userLink, nil
}
