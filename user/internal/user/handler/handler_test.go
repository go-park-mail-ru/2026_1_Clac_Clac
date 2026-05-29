package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/user/v1"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/common"
	mockAuthSrv "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/handler/mock_auth_srv"
	mockHTTPClient "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/handler/mock_http_client"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/service"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/service/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	fixedUUID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	jpegMagic = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}
)

func defaultConfig() Config {
	return Config{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		RedirectURI:  "https://example.com/oauth/vk",
	}
}

func newHandler(srv *mockAuthSrv.AuthService) *Handler {
	return NewHandler(srv, defaultConfig(), nil)
}

func newHandlerWithHTTP(srv *mockAuthSrv.AuthService, httpClient *mockHTTPClient.HTTPClient) *Handler {
	return NewHandler(srv, defaultConfig(), httpClient)
}

func vkAuthResponse(accessToken string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(`{"access_token":"` + accessToken + `"}`))
}

func vkUserInfoResponse(firstName, email string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(`{"user":{"user_id":"123","first_name":"` + firstName + `","email":"` + email + `","avatar":""}}`))
}

func vkEmptyUserInfoResponse() io.ReadCloser {
	return io.NopCloser(strings.NewReader(`{"user":{"user_id":"0","first_name":"","email":"","avatar":""}}`))
}

func assertGRPCCode(t *testing.T, err error, expected codes.Code) {
	t.Helper()
	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok, "error must be a gRPC status error")
	assert.Equal(t, expected, st.Code())
}

func TestLogInUser(t *testing.T) {
	t.Run("SuccessLogin", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("GetUser", mock.Anything, serviceDto.GetUserInfo{
			Email:    "user@mail.ru",
			Password: "password1",
		}).Return(serviceDto.UserInfo{
			Link:        fixedUUID,
			DisplayName: "Artem",
			Email:       "user@mail.ru",
		}, nil)

		resp, err := newHandler(m).GetUser(context.Background(), &pb.GetUserRequest{
			Email:    "user@mail.ru",
			Password: "password1",
		})

		assert.NoError(t, err)
		assert.Equal(t, "user@mail.ru", resp.Email)
		assert.Equal(t, fixedUUID.String(), resp.UserLink)
	})

	t.Run("WrongPassword", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("GetUser", mock.Anything, mock.Anything).Return(serviceDto.UserInfo{}, service.ErrorWrongPassword)

		_, err := newHandler(m).GetUser(context.Background(), &pb.GetUserRequest{
			Email:    "user@mail.ru",
			Password: "wrong",
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("InternalError", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("GetUser", mock.Anything, mock.Anything).Return(serviceDto.UserInfo{}, errors.New("db error"))

		_, err := newHandler(m).GetUser(context.Background(), &pb.GetUserRequest{
			Email:    "user@mail.ru",
			Password: "password1",
		})

		assertGRPCCode(t, err, codes.Internal)
	})
}

func TestRegisterUser(t *testing.T) {
	t.Run("SuccessRegistration", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("CreateUser", mock.Anything, mock.AnythingOfType("dto.EntityUser")).Return(serviceDto.UserInfo{
			Link:        fixedUUID,
			DisplayName: "Artem",
			Email:       "user@mail.ru",
		}, nil)

		resp, err := newHandler(m).CreateUser(context.Background(), &pb.CreateRequest{
			Email:       "user@mail.ru",
			Password:    "password1",
			DisplayName: "Artem",
		})

		assert.NoError(t, err)
		assert.Equal(t, "user@mail.ru", resp.Email)
	})

	t.Run("UserAlreadyExists", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("CreateUser", mock.Anything, mock.AnythingOfType("dto.EntityUser")).Return(serviceDto.UserInfo{}, common.ErrorExistingUser)

		_, err := newHandler(m).CreateUser(context.Background(), &pb.CreateRequest{
			Email:    "user@mail.ru",
			Password: "password1",
		})

		assertGRPCCode(t, err, codes.AlreadyExists)
	})

	t.Run("NullFieldError", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("CreateUser", mock.Anything, mock.AnythingOfType("dto.EntityUser")).Return(serviceDto.UserInfo{}, common.ErrorNotNullValue)

		_, err := newHandler(m).CreateUser(context.Background(), &pb.CreateRequest{
			Email:    "user@mail.ru",
			Password: "password1",
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("InternalError", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("CreateUser", mock.Anything, mock.AnythingOfType("dto.EntityUser")).Return(serviceDto.UserInfo{}, errors.New("db error"))

		_, err := newHandler(m).CreateUser(context.Background(), &pb.CreateRequest{
			Email:    "user@mail.ru",
			Password: "password1",
		})

		assertGRPCCode(t, err, codes.Internal)
	})
}

func TestGetUserLink(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("GetUserLink", mock.Anything, "user@mail.ru").Return(fixedUUID.String(), nil)

		resp, err := newHandler(m).GetUserLink(context.Background(), &pb.GetUserLinkRequest{
			Email: "user@mail.ru",
		})

		assert.NoError(t, err)
		assert.Equal(t, fixedUUID.String(), resp.UserLink)
	})

	t.Run("EmailNotFound", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("GetUserLink", mock.Anything, "notfound@mail.ru").Return("", common.ErrorNonexistentEmail)

		_, err := newHandler(m).GetUserLink(context.Background(), &pb.GetUserLinkRequest{
			Email: "notfound@mail.ru",
		})

		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("InternalError", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("GetUserLink", mock.Anything, "user@mail.ru").Return("", errors.New("db error"))

		_, err := newHandler(m).GetUserLink(context.Background(), &pb.GetUserLinkRequest{
			Email: "user@mail.ru",
		})

		assertGRPCCode(t, err, codes.Internal)
	})
}

func TestResetPassword(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("ResetPassword", mock.Anything, serviceDto.ResetPasswordInfo{
			UserLink:    fixedUUID.String(),
			NewPassword: "newpassword1",
		}).Return(nil)

		resp, err := newHandler(m).ResetPassword(context.Background(), &pb.ResetPasswordRequest{
			UserLink: fixedUUID.String(),
			Password: "newpassword1",
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("ResetPassword", mock.Anything, mock.Anything).Return(common.ErrorNonexistentUser)

		_, err := newHandler(m).ResetPassword(context.Background(), &pb.ResetPasswordRequest{
			UserLink: fixedUUID.String(),
			Password: "newpassword1",
		})

		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("NullFieldError", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("ResetPassword", mock.Anything, mock.Anything).Return(common.ErrorNotNullValue)

		_, err := newHandler(m).ResetPassword(context.Background(), &pb.ResetPasswordRequest{
			UserLink: fixedUUID.String(),
			Password: "newpassword1",
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("InternalError", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("ResetPassword", mock.Anything, mock.Anything).Return(errors.New("db error"))

		_, err := newHandler(m).ResetPassword(context.Background(), &pb.ResetPasswordRequest{
			UserLink: fixedUUID.String(),
			Password: "newpassword1",
		})

		assertGRPCCode(t, err, codes.Internal)
	})
}

func TestProcessUserWithVK(t *testing.T) {
	const (
		vkCode        = "vk_auth_code"
		vkVerifier    = "vk_verifier"
		vkState       = "vk_state"
		vkAccessToken = "vk_access_token_abc"
		firstName     = "Artem"
		vkEmail       = "user@vk.com"
		expectedLink  = "00000000-0000-0000-0000-000000000001"
	)

	t.Run("SuccessExistingUser", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		httpMock := mockHTTPClient.NewHTTPClient(t)

		httpMock.On("PostForm", "https://id.vk.ru/oauth2/auth", mock.AnythingOfType("url.Values")).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       vkAuthResponse(vkAccessToken),
		}, nil)

		httpMock.On("PostForm", "https://id.vk.ru/oauth2/user_info", mock.AnythingOfType("url.Values")).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       vkUserInfoResponse(firstName, vkEmail),
		}, nil)

		m.On("EnsureUserByEmail", mock.Anything, serviceDto.EntityUser{
			DisplayName: firstName,
			Email:       vkEmail,
		}).Return(expectedLink, nil)

		resp, err := newHandlerWithHTTP(m, httpMock).ProcessUserWithVK(context.Background(), &pb.ProcessUserVKRequest{
			Code:         vkCode,
			CodeVerifier: vkVerifier,
			State:        vkState,
		})

		assert.NoError(t, err)
		assert.Equal(t, expectedLink, resp.UserLink)
	})

	t.Run("MissingCode", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		_, err := newHandlerWithHTTP(m, mockHTTPClient.NewHTTPClient(t)).ProcessUserWithVK(context.Background(), &pb.ProcessUserVKRequest{
			CodeVerifier: vkVerifier,
			State:        vkState,
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("MissingCodeVerifier", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		_, err := newHandlerWithHTTP(m, mockHTTPClient.NewHTTPClient(t)).ProcessUserWithVK(context.Background(), &pb.ProcessUserVKRequest{
			Code:  vkCode,
			State: vkState,
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("MissingState", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		_, err := newHandlerWithHTTP(m, mockHTTPClient.NewHTTPClient(t)).ProcessUserWithVK(context.Background(), &pb.ProcessUserVKRequest{
			Code:         vkCode,
			CodeVerifier: vkVerifier,
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("VKAuthRequestFails", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		httpMock := mockHTTPClient.NewHTTPClient(t)

		httpMock.On("PostForm", "https://id.vk.ru/oauth2/auth", mock.AnythingOfType("url.Values")).Return((*http.Response)(nil), errors.New("network error"))

		_, err := newHandlerWithHTTP(m, httpMock).ProcessUserWithVK(context.Background(), &pb.ProcessUserVKRequest{
			Code:         vkCode,
			CodeVerifier: vkVerifier,
			State:        vkState,
		})

		assertGRPCCode(t, err, codes.Unavailable)
	})

	t.Run("VKAuthEmptyToken", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		httpMock := mockHTTPClient.NewHTTPClient(t)

		httpMock.On("PostForm", "https://id.vk.ru/oauth2/auth", mock.AnythingOfType("url.Values")).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       vkAuthResponse(""),
		}, nil)

		_, err := newHandlerWithHTTP(m, httpMock).ProcessUserWithVK(context.Background(), &pb.ProcessUserVKRequest{
			Code:         vkCode,
			CodeVerifier: vkVerifier,
			State:        vkState,
		})

		assertGRPCCode(t, err, codes.Unavailable)
	})

	t.Run("VKAuthErrorResponse", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		httpMock := mockHTTPClient.NewHTTPClient(t)

		httpMock.On("PostForm", "https://id.vk.ru/oauth2/auth", mock.AnythingOfType("url.Values")).Return(&http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader(`{"error":"invalid_grant","error_description":"Code has expired or already been used"}`)),
		}, nil)

		_, err := newHandlerWithHTTP(m, httpMock).ProcessUserWithVK(context.Background(), &pb.ProcessUserVKRequest{
			Code:         vkCode,
			CodeVerifier: vkVerifier,
			State:        vkState,
		})

		assertGRPCCode(t, err, codes.Unavailable)
	})

	t.Run("VKUserInfoRequestFails", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		httpMock := mockHTTPClient.NewHTTPClient(t)

		httpMock.On("PostForm", "https://id.vk.ru/oauth2/auth", mock.AnythingOfType("url.Values")).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       vkAuthResponse(vkAccessToken),
		}, nil)

		httpMock.On("PostForm", "https://id.vk.ru/oauth2/user_info", mock.AnythingOfType("url.Values")).Return((*http.Response)(nil), errors.New("network error"))

		_, err := newHandlerWithHTTP(m, httpMock).ProcessUserWithVK(context.Background(), &pb.ProcessUserVKRequest{
			Code:         vkCode,
			CodeVerifier: vkVerifier,
			State:        vkState,
		})

		assertGRPCCode(t, err, codes.Unavailable)
	})

	t.Run("VKUserInfoEmptyUser", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		httpMock := mockHTTPClient.NewHTTPClient(t)

		httpMock.On("PostForm", "https://id.vk.ru/oauth2/auth", mock.AnythingOfType("url.Values")).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       vkAuthResponse(vkAccessToken),
		}, nil)

		httpMock.On("PostForm", "https://id.vk.ru/oauth2/user_info", mock.AnythingOfType("url.Values")).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       vkEmptyUserInfoResponse(),
		}, nil)

		_, err := newHandlerWithHTTP(m, httpMock).ProcessUserWithVK(context.Background(), &pb.ProcessUserVKRequest{
			Code:         vkCode,
			CodeVerifier: vkVerifier,
			State:        vkState,
		})

		assertGRPCCode(t, err, codes.Internal)
	})

	t.Run("EnsureUserByEmailFails", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		httpMock := mockHTTPClient.NewHTTPClient(t)

		httpMock.On("PostForm", "https://id.vk.ru/oauth2/auth", mock.AnythingOfType("url.Values")).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       vkAuthResponse(vkAccessToken),
		}, nil)

		httpMock.On("PostForm", "https://id.vk.ru/oauth2/user_info", mock.AnythingOfType("url.Values")).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       vkUserInfoResponse(firstName, vkEmail),
		}, nil)

		m.On("EnsureUserByEmail", mock.Anything, mock.Anything).Return("", errors.New("db error"))

		_, err := newHandlerWithHTTP(m, httpMock).ProcessUserWithVK(context.Background(), &pb.ProcessUserVKRequest{
			Code:         vkCode,
			CodeVerifier: vkVerifier,
			State:        vkState,
		})

		assertGRPCCode(t, err, codes.Internal)
	})
}

func TestGetProfile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("GetProfile", mock.Anything, fixedUUID).Return(serviceDto.UserInfo{
			Link:        fixedUUID,
			DisplayName: "Artem",
			Email:       "user@mail.ru",
			AvatarURL:   "https://cdn.example.com/avatar.jpg",
		}, nil)

		resp, err := newHandler(m).GetProfile(context.Background(), &pb.UserLinkRequest{
			UserLink: fixedUUID.String(),
		})

		assert.NoError(t, err)
		assert.Equal(t, "Artem", resp.DisplayName)
	})

	t.Run("InvalidUUID", func(t *testing.T) {
		_, err := newHandler(mockAuthSrv.NewAuthService(t)).GetProfile(context.Background(), &pb.UserLinkRequest{
			UserLink: "not-a-uuid",
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("GetProfile", mock.Anything, fixedUUID).Return(serviceDto.UserInfo{}, common.ErrorNonexistentUser)

		_, err := newHandler(m).GetProfile(context.Background(), &pb.UserLinkRequest{
			UserLink: fixedUUID.String(),
		})

		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("InternalError", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("GetProfile", mock.Anything, fixedUUID).Return(serviceDto.UserInfo{}, errors.New("db error"))

		_, err := newHandler(m).GetProfile(context.Background(), &pb.UserLinkRequest{
			UserLink: fixedUUID.String(),
		})

		assertGRPCCode(t, err, codes.Internal)
	})
}

func TestUpdateProfile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("UpdateProfile", mock.Anything, mock.AnythingOfType("dto.UpdatedUserInfo")).Return(nil)

		resp, err := newHandler(m).UpdateProfile(context.Background(), &pb.UpdateProfileRequest{
			UserLink:    fixedUUID.String(),
			DisplayName: "New Name",
			Description: "New description",
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("InvalidUUID", func(t *testing.T) {
		_, err := newHandler(mockAuthSrv.NewAuthService(t)).UpdateProfile(context.Background(), &pb.UpdateProfileRequest{
			UserLink:    "bad-uuid",
			DisplayName: "Name",
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("MissingRequiredField", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("UpdateProfile", mock.Anything, mock.AnythingOfType("dto.UpdatedUserInfo")).Return(common.ErrorMissingRequiredField)

		_, err := newHandler(m).UpdateProfile(context.Background(), &pb.UpdateProfileRequest{
			UserLink:    fixedUUID.String(),
			DisplayName: "Name",
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("InternalError", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("UpdateProfile", mock.Anything, mock.AnythingOfType("dto.UpdatedUserInfo")).Return(errors.New("db error"))

		_, err := newHandler(m).UpdateProfile(context.Background(), &pb.UpdateProfileRequest{
			UserLink:    fixedUUID.String(),
			DisplayName: "Name",
		})

		assertGRPCCode(t, err, codes.Internal)
	})
}

type mockUpdateAvatarStream struct {
	grpc.ServerStream
	ctx      context.Context
	requests []*pb.UpdateAvatarRequest
	callIdx  int
	resp     *pb.AvatarResponse
	sendErr  error
}

func (m *mockUpdateAvatarStream) Recv() (*pb.UpdateAvatarRequest, error) {
	if m.callIdx >= len(m.requests) {
		return nil, io.EOF
	}
	req := m.requests[m.callIdx]
	m.callIdx++
	return req, nil
}

func (m *mockUpdateAvatarStream) SendAndClose(res *pb.AvatarResponse) error {
	m.resp = res
	return m.sendErr
}

func (m *mockUpdateAvatarStream) Context() context.Context {
	return m.ctx
}

func (m *mockUpdateAvatarStream) SetHeader(md metadata.MD) error  { return nil }
func (m *mockUpdateAvatarStream) SendHeader(md metadata.MD) error { return nil }
func (m *mockUpdateAvatarStream) SetTrailer(md metadata.MD)       {}
func (m *mockUpdateAvatarStream) SendMsg(msg any) error           { return nil }
func (m *mockUpdateAvatarStream) RecvMsg(msg any) error           { return nil }

func TestUpdateAvatar(t *testing.T) {
	t.Run("SuccessJPEG", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("UpdateAvatar", mock.Anything, mock.AnythingOfType("dto.UpdatedAvatar")).Return("https://cdn.example.com/avatar.jpg", nil)

		stream := &mockUpdateAvatarStream{
			ctx: context.Background(),
			requests: []*pb.UpdateAvatarRequest{
				{
					Request: &pb.UpdateAvatarRequest_Metadata{
						Metadata: &pb.MetadataUpdateAvatar{
							UserLink:    fixedUUID.String(),
							ContentType: "image/jpeg",
						},
					},
				},
				{
					Request: &pb.UpdateAvatarRequest_FileData{
						FileData: jpegMagic,
					},
				},
			},
		}

		err := newHandler(m).UpdateAvatar(stream)

		assert.NoError(t, err)
		assert.Contains(t, stream.resp.AvatarUrl, "avatar.jpg")
	})

	t.Run("InvalidUUID", func(t *testing.T) {
		stream := &mockUpdateAvatarStream{
			ctx: context.Background(),
			requests: []*pb.UpdateAvatarRequest{
				{
					Request: &pb.UpdateAvatarRequest_Metadata{
						Metadata: &pb.MetadataUpdateAvatar{
							UserLink:    "bad-uuid",
							ContentType: "image/jpeg",
						},
					},
				},
				{
					Request: &pb.UpdateAvatarRequest_FileData{
						FileData: jpegMagic,
					},
				},
			},
		}

		err := newHandler(mockAuthSrv.NewAuthService(t)).UpdateAvatar(stream)

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("UpdateAvatar", mock.Anything, mock.AnythingOfType("dto.UpdatedAvatar")).Return("", common.ErrorNonexistentUser)

		stream := &mockUpdateAvatarStream{
			ctx: context.Background(),
			requests: []*pb.UpdateAvatarRequest{
				{
					Request: &pb.UpdateAvatarRequest_Metadata{
						Metadata: &pb.MetadataUpdateAvatar{
							UserLink:    fixedUUID.String(),
							ContentType: "image/jpeg",
						},
					},
				},
				{
					Request: &pb.UpdateAvatarRequest_FileData{
						FileData: jpegMagic,
					},
				},
			},
		}

		err := newHandler(m).UpdateAvatar(stream)

		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("ServiceFailure", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("UpdateAvatar", mock.Anything, mock.AnythingOfType("dto.UpdatedAvatar")).Return("", errors.New("s3 error"))

		stream := &mockUpdateAvatarStream{
			ctx: context.Background(),
			requests: []*pb.UpdateAvatarRequest{
				{
					Request: &pb.UpdateAvatarRequest_Metadata{
						Metadata: &pb.MetadataUpdateAvatar{
							UserLink:    fixedUUID.String(),
							ContentType: "image/jpeg",
						},
					},
				},
				{
					Request: &pb.UpdateAvatarRequest_FileData{
						FileData: jpegMagic,
					},
				},
			},
		}

		err := newHandler(m).UpdateAvatar(stream)

		assertGRPCCode(t, err, codes.Internal)
	})
}

func TestDeleteAvatar(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("DeleteAvatar", mock.Anything, fixedUUID).Return(nil)

		resp, err := newHandler(m).DeleteAvatar(context.Background(), &pb.UserLinkRequest{
			UserLink: fixedUUID.String(),
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("InvalidUUID", func(t *testing.T) {
		_, err := newHandler(mockAuthSrv.NewAuthService(t)).DeleteAvatar(context.Background(), &pb.UserLinkRequest{
			UserLink: "bad-uuid",
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("DeleteAvatar", mock.Anything, fixedUUID).Return(common.ErrorNonexistentUser)

		_, err := newHandler(m).DeleteAvatar(context.Background(), &pb.UserLinkRequest{
			UserLink: fixedUUID.String(),
		})

		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("InternalError", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("DeleteAvatar", mock.Anything, fixedUUID).Return(errors.New("s3 error"))

		_, err := newHandler(m).DeleteAvatar(context.Background(), &pb.UserLinkRequest{
			UserLink: fixedUUID.String(),
		})

		assertGRPCCode(t, err, codes.Internal)
	})
}
