package clients

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/user/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockUserServiceClient struct {
	mock.Mock
}

func (m *mockUserServiceClient) GetProfile(ctx context.Context, in *pb.UserLinkRequest, opts ...grpc.CallOption) (*pb.ProfileResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ProfileResponse), args.Error(1)
}

func (m *mockUserServiceClient) GetProfiles(ctx context.Context, in *pb.GetProfilesRequest, opts ...grpc.CallOption) (*pb.GetProfilesResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetProfilesResponse), args.Error(1)
}

func (m *mockUserServiceClient) UpdateProfile(ctx context.Context, in *pb.UpdateProfileRequest, opts ...grpc.CallOption) (*pb.UpdateProfileResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UpdateProfileResponse), args.Error(1)
}

func (m *mockUserServiceClient) UpdateAvatar(ctx context.Context, in *pb.UpdateAvatarRequest, opts ...grpc.CallOption) (*pb.AvatarResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.AvatarResponse), args.Error(1)
}

func (m *mockUserServiceClient) DeleteAvatar(ctx context.Context, in *pb.UserLinkRequest, opts ...grpc.CallOption) (*pb.DeleteAvatarResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.DeleteAvatarResponse), args.Error(1)
}

func (m *mockUserServiceClient) GetUser(ctx context.Context, in *pb.GetUserRequest, opts ...grpc.CallOption) (*pb.UserResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UserResponse), args.Error(1)
}

func (m *mockUserServiceClient) CreateUser(ctx context.Context, in *pb.CreateRequest, opts ...grpc.CallOption) (*pb.UserResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UserResponse), args.Error(1)
}

func (m *mockUserServiceClient) GetUserLink(ctx context.Context, in *pb.GetUserLinkRequest, opts ...grpc.CallOption) (*pb.GetUserLinkResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetUserLinkResponse), args.Error(1)
}

func (m *mockUserServiceClient) ResetPassword(ctx context.Context, in *pb.ResetPasswordRequest, opts ...grpc.CallOption) (*pb.ResetPasswordResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ResetPasswordResponse), args.Error(1)
}

func (m *mockUserServiceClient) ProcessUserWithVK(ctx context.Context, in *pb.ProcessUserVKRequest, opts ...grpc.CallOption) (*pb.ProcessUserVKResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ProcessUserVKResponse), args.Error(1)
}

func TestUserGetProfile(t *testing.T) {
	ctx := context.Background()
	validUUID := uuid.New()

	tests := []struct {
		name         string
		userLink     uuid.UUID
		mockResp     *pb.ProfileResponse
		mockErr      error
		expectedUser domain.FullInfoUser
		expectedErr  error
	}{
		{
			name:     "success",
			userLink: validUUID,
			mockResp: &pb.ProfileResponse{
				UserLink:    validUUID.String(),
				Email:       "test@example.com",
				DisplayName: "Test User",
				Description: "Bio",
				AvatarUrl:   "https://cdn/avatar.png",
			},
			mockErr: nil,
			expectedUser: domain.FullInfoUser{
				UserLink:    validUUID,
				Email:       "test@example.com",
				DisplayName: "Test User",
				Description: "Bio",
				AvatarURL:   "https://cdn/avatar.png",
			},
			expectedErr: nil,
		},
		{
			name:         "grpc error",
			userLink:     validUUID,
			mockResp:     nil,
			mockErr:      status.Error(codes.NotFound, "user not found"),
			expectedUser: domain.FullInfoUser{},
			expectedErr:  common.ErrorNonexistentUser,
		},
		{
			name:         "invalid uuid in response",
			userLink:     validUUID,
			mockResp:     &pb.ProfileResponse{UserLink: "bad-uuid"},
			mockErr:      nil,
			expectedUser: domain.FullInfoUser{},
			expectedErr:  common.ErrorParseLink,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockUserServiceClient)
			mc.On("GetProfile", ctx, &pb.UserLinkRequest{UserLink: tt.userLink.String()}).
				Return(tt.mockResp, tt.mockErr)

			u := &User{client: mc}
			user, err := u.GetProfile(ctx, tt.userLink)

			assert.Equal(t, tt.expectedUser, user)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	ctx := context.Background()
	validUUID := uuid.New()

	tests := []struct {
		name        string
		updatedInfo domain.UpdatedInfo
		mockResp    *pb.UpdateProfileResponse
		mockErr     error
		expectedErr error
	}{
		{
			name: "success",
			updatedInfo: domain.UpdatedInfo{
				UserLink:    validUUID,
				DisplayName: "New Name",
				Description: "New Bio",
			},
			mockResp:    &pb.UpdateProfileResponse{},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name: "grpc error",
			updatedInfo: domain.UpdatedInfo{
				UserLink:    validUUID,
				DisplayName: "New Name",
				Description: "New Bio",
			},
			mockResp:    nil,
			mockErr:     status.Error(codes.NotFound, "user not found"),
			expectedErr: common.ErrorNonexistentUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockUserServiceClient)
			mc.On("UpdateProfile", ctx, &pb.UpdateProfileRequest{
				UserLink:    tt.updatedInfo.UserLink.String(),
				DisplayName: tt.updatedInfo.DisplayName,
				Description: tt.updatedInfo.Description,
			}).Return(tt.mockResp, tt.mockErr)

			u := &User{client: mc}
			err := u.UpdateProfile(ctx, tt.updatedInfo)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateAvatar(t *testing.T) {
	ctx := context.Background()
	validUUID := uuid.New()
	fileData := []byte("image-data")

	tests := []struct {
		name        string
		avatarInfo  domain.AvatarInfo
		mockResp    *pb.AvatarResponse
		mockErr     error
		expectedURL string
		expectedErr error
	}{
		{
			name: "success",
			avatarInfo: domain.AvatarInfo{
				UserLink:      validUUID,
				FileData:      fileData,
				ContentType:   "image/png",
				FileExtension: "png",
			},
			mockResp:    &pb.AvatarResponse{AvatarUrl: "https://cdn/new.png"},
			mockErr:     nil,
			expectedURL: "https://cdn/new.png",
			expectedErr: nil,
		},
		{
			name: "grpc error",
			avatarInfo: domain.AvatarInfo{
				UserLink:      validUUID,
				FileData:      fileData,
				ContentType:   "image/png",
				FileExtension: "png",
			},
			mockResp:    nil,
			mockErr:     status.Error(codes.NotFound, "user not found"),
			expectedURL: "",
			expectedErr: common.ErrorNonexistentUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockUserServiceClient)
			mc.On("UpdateAvatar", ctx, &pb.UpdateAvatarRequest{
				UserLink:      tt.avatarInfo.UserLink.String(),
				FileData:      tt.avatarInfo.FileData,
				ContentType:   tt.avatarInfo.ContentType,
				FileExtension: tt.avatarInfo.FileExtension,
			}).Return(tt.mockResp, tt.mockErr)

			u := &User{client: mc}
			url, err := u.UpdateAvatar(ctx, tt.avatarInfo)

			assert.Equal(t, tt.expectedURL, url)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteAvatar(t *testing.T) {
	ctx := context.Background()
	validUUID := uuid.New()

	tests := []struct {
		name        string
		userLink    uuid.UUID
		mockResp    *pb.DeleteAvatarResponse
		mockErr     error
		expectedErr error
	}{
		{
			name:        "success",
			userLink:    validUUID,
			mockResp:    &pb.DeleteAvatarResponse{},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "grpc error",
			userLink:    validUUID,
			mockResp:    nil,
			mockErr:     status.Error(codes.NotFound, "user not found"),
			expectedErr: common.ErrorNonexistentUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockUserServiceClient)
			mc.On("DeleteAvatar", ctx, &pb.UserLinkRequest{UserLink: tt.userLink.String()}).
				Return(tt.mockResp, tt.mockErr)

			u := &User{client: mc}
			err := u.DeleteAvatar(ctx, tt.userLink)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	ctx := context.Background()
	validUUID := uuid.New()

	tests := []struct {
		name         string
		credentials  domain.Credentials
		mockResp     *pb.UserResponse
		mockErr      error
		expectedUser domain.FullInfoUser
		expectedErr  error
	}{
		{
			name:        "success",
			credentials: domain.Credentials{Email: "user@example.com", Password: "pass123"},
			mockResp: &pb.UserResponse{
				UserLink:    validUUID.String(),
				Email:       "user@example.com",
				DisplayName: "User",
				Avatar:      "https://cdn/avatar.png",
			},
			mockErr: nil,
			expectedUser: domain.FullInfoUser{
				UserLink:    validUUID,
				Email:       "user@example.com",
				DisplayName: "User",
				AvatarURL:   "https://cdn/avatar.png",
			},
			expectedErr: nil,
		},
		{
			name:         "grpc error",
			credentials:  domain.Credentials{Email: "user@example.com", Password: "wrong"},
			mockResp:     nil,
			mockErr:      status.Error(codes.InvalidArgument, "wrong credentials"),
			expectedUser: domain.FullInfoUser{},
			expectedErr:  common.ErrorWrongCredentials,
		},
		{
			name:         "invalid uuid in response",
			credentials:  domain.Credentials{Email: "user@example.com", Password: "pass123"},
			mockResp:     &pb.UserResponse{UserLink: "bad-uuid"},
			mockErr:      nil,
			expectedUser: domain.FullInfoUser{},
			expectedErr:  common.ErrorParseLink,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockUserServiceClient)
			mc.On("GetUser", ctx, &pb.GetUserRequest{
				Email:    tt.credentials.Email,
				Password: tt.credentials.Password,
			}).Return(tt.mockResp, tt.mockErr)

			u := &User{client: mc}
			user, err := u.GetUser(ctx, tt.credentials)

			assert.Equal(t, tt.expectedUser, user)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	validUUID := uuid.New()

	tests := []struct {
		name         string
		newUser      domain.NewCredentialsUser
		mockResp     *pb.UserResponse
		mockErr      error
		expectedUser domain.FullInfoUser
		expectedErr  error
	}{
		{
			name: "success",
			newUser: domain.NewCredentialsUser{
				DisplayName: "New User",
				Email:       "new@example.com",
				Password:    "pass123",
			},
			mockResp: &pb.UserResponse{
				UserLink:    validUUID.String(),
				Email:       "new@example.com",
				DisplayName: "New User",
				Avatar:      "",
			},
			mockErr: nil,
			expectedUser: domain.FullInfoUser{
				UserLink:    validUUID,
				Email:       "new@example.com",
				DisplayName: "New User",
			},
			expectedErr: nil,
		},
		{
			name: "grpc error - user exists",
			newUser: domain.NewCredentialsUser{
				DisplayName: "New User",
				Email:       "existing@example.com",
				Password:    "pass123",
			},
			mockResp:     nil,
			mockErr:      status.Error(codes.AlreadyExists, "user exists"),
			expectedUser: domain.FullInfoUser{},
			expectedErr:  common.ErrorExistingUser,
		},
		{
			name: "invalid uuid in response",
			newUser: domain.NewCredentialsUser{
				DisplayName: "New User",
				Email:       "new@example.com",
				Password:    "pass123",
			},
			mockResp:     &pb.UserResponse{UserLink: "bad-uuid"},
			mockErr:      nil,
			expectedUser: domain.FullInfoUser{},
			expectedErr:  common.ErrorParseLink,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockUserServiceClient)
			mc.On("CreateUser", ctx, &pb.CreateRequest{
				DisplayName: tt.newUser.DisplayName,
				Email:       tt.newUser.Email,
				Password:    tt.newUser.Password,
			}).Return(tt.mockResp, tt.mockErr)

			u := &User{client: mc}
			user, err := u.CreateUser(ctx, tt.newUser)

			assert.Equal(t, tt.expectedUser, user)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetUserLink(t *testing.T) {
	ctx := context.Background()
	validUUID := uuid.New()

	tests := []struct {
		name         string
		email        string
		mockResp     *pb.GetUserLinkResponse
		mockErr      error
		expectedLink uuid.UUID
		expectedErr  error
	}{
		{
			name:         "success",
			email:        "user@example.com",
			mockResp:     &pb.GetUserLinkResponse{UserLink: validUUID.String()},
			mockErr:      nil,
			expectedLink: validUUID,
			expectedErr:  nil,
		},
		{
			name:         "grpc error",
			email:        "unknown@example.com",
			mockResp:     nil,
			mockErr:      status.Error(codes.NotFound, "email not found"),
			expectedLink: uuid.Nil,
			expectedErr:  common.ErrorNonexistentEmail,
		},
		{
			name:         "invalid uuid in response",
			email:        "user@example.com",
			mockResp:     &pb.GetUserLinkResponse{UserLink: "bad-uuid"},
			mockErr:      nil,
			expectedLink: uuid.Nil,
			expectedErr:  common.ErrorParseLink,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockUserServiceClient)
			mc.On("GetUserLink", ctx, &pb.GetUserLinkRequest{Email: tt.email}).
				Return(tt.mockResp, tt.mockErr)

			u := &User{client: mc}
			link, err := u.GetUserLink(ctx, tt.email)

			assert.Equal(t, tt.expectedLink, link)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResetPassword(t *testing.T) {
	ctx := context.Background()
	validUUID := uuid.New()

	tests := []struct {
		name            string
		updatedPassword domain.UpdatedPassword
		mockResp        *pb.ResetPasswordResponse
		mockErr         error
		expectedErr     error
	}{
		{
			name: "success",
			updatedPassword: domain.UpdatedPassword{
				UserLink: validUUID,
				Password: "newpass123",
			},
			mockResp:    &pb.ResetPasswordResponse{},
			mockErr:     nil,
			expectedErr: nil,
		},
		{
			name: "grpc error",
			updatedPassword: domain.UpdatedPassword{
				UserLink: validUUID,
				Password: "newpass123",
			},
			mockResp:    nil,
			mockErr:     status.Error(codes.NotFound, "user not found"),
			expectedErr: common.ErrorNonexistentUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockUserServiceClient)
			mc.On("ResetPassword", ctx, &pb.ResetPasswordRequest{
				UserLink: tt.updatedPassword.UserLink.String(),
				Password: tt.updatedPassword.Password,
			}).Return(tt.mockResp, tt.mockErr)

			u := &User{client: mc}
			err := u.ResetPassword(ctx, tt.updatedPassword)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProcessUserWithVK(t *testing.T) {
	ctx := context.Background()
	validUUID := uuid.New()

	tests := []struct {
		name         string
		accessToken  string
		email        string
		mockResp     *pb.ProcessUserVKResponse
		mockErr      error
		expectedLink uuid.UUID
		expectedErr  error
	}{
		{
			name:         "success",
			accessToken:  "vk-token-abc",
			email:        "vkuser@example.com",
			mockResp:     &pb.ProcessUserVKResponse{UserLink: validUUID.String()},
			mockErr:      nil,
			expectedLink: validUUID,
			expectedErr:  nil,
		},
		{
			name:         "grpc error",
			accessToken:  "invalid-token",
			email:        "vkuser@example.com",
			mockResp:     nil,
			mockErr:      status.Error(codes.Unavailable, "vk unavailable"),
			expectedLink: uuid.Nil,
			expectedErr:  common.ErrorServiceUnavailable,
		},
		{
			name:         "invalid uuid in response",
			accessToken:  "vk-token-abc",
			email:        "vkuser@example.com",
			mockResp:     &pb.ProcessUserVKResponse{UserLink: "bad-uuid"},
			mockErr:      nil,
			expectedLink: uuid.Nil,
			expectedErr:  common.ErrorParseLink,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := new(mockUserServiceClient)
			mc.On("ProcessUserWithVK", ctx, &pb.ProcessUserVKRequest{
				AccessToken: tt.accessToken,
				Email:       tt.email,
			}).Return(tt.mockResp, tt.mockErr)

			u := &User{client: mc}
			link, err := u.ProcessUserWithVK(ctx, tt.accessToken, tt.email)

			assert.Equal(t, tt.expectedLink, link)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
