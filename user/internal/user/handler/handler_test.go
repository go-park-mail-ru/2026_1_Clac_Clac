package handler

import (
	"context"
	"errors"
	"testing"

	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/contracts/user"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/common"
	mockAuthSrv "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/handler/mock_auth_srv"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/service"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/user/service/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	fixedUUID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	jpegMagic = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}
	pngMagic  = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}
)

func defaultConfig() Config {
	return Config{
		MaxLenPassword:        128,
		MinLenPassword:        8,
		MaxLenNameUser:        128,
		MaxLenDescriptionUser: 500,
		MaxReadBytes:          5 * 1024 * 1024,
		SignatureTypeBytes:    512,
		ValidExtensions: map[string]struct{}{
			"image/jpeg": {},
			"image/png":  {},
			"image/webp": {},
		},
		APIMethod: "https://api.vk.com/users.get?token=%s",
	}
}

func newHandler(srv *mockAuthSrv.AuthService) *Handler {
	return NewHandler(srv, defaultConfig(), nil)
}

func TestLogInUser(t *testing.T) {
	t.Run("Success login", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("LogIn", mock.Anything, serviceDto.GetUserInfo{
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

	t.Run("Error invalid email format", func(t *testing.T) {
		_, err := newHandler(mockAuthSrv.NewAuthService(t)).GetUser(context.Background(), &pb.GetUserRequest{
			Email:    "not-email",
			Password: "password1",
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("Error wrong password", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("LogIn", mock.Anything, mock.Anything).Return(serviceDto.UserInfo{}, service.ErrorWrongPassword)

		_, err := newHandler(m).GetUser(context.Background(), &pb.GetUserRequest{
			Email:    "user@mail.ru",
			Password: "password1",
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("Error internal service failure", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("LogIn", mock.Anything, mock.Anything).Return(serviceDto.UserInfo{}, errors.New("db error"))

		_, err := newHandler(m).GetUser(context.Background(), &pb.GetUserRequest{
			Email:    "user@mail.ru",
			Password: "password1",
		})

		assertGRPCCode(t, err, codes.Internal)
	})
}

func TestRegisterUser(t *testing.T) {
	t.Run("Success registration", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("Register", mock.Anything, mock.AnythingOfType("dto.RegistrationUser")).Return(serviceDto.UserInfo{
			Link:        fixedUUID,
			DisplayName: "Artem",
			Email:       "user@mail.ru",
		}, nil)

		resp, err := newHandler(m).CreateUser(context.Background(), &pb.CreateRequest{
			Email:            "user@mail.ru",
			Password:         "password1",
			RepeatedPassword: "password1",
			DisplayName:      "Artem",
		})

		assert.NoError(t, err)
		assert.Equal(t, "user@mail.ru", resp.Email)
	})

	t.Run("Error passwords do not match", func(t *testing.T) {
		_, err := newHandler(mockAuthSrv.NewAuthService(t)).CreateUser(context.Background(), &pb.CreateRequest{
			Email:            "user@mail.ru",
			Password:         "password1",
			RepeatedPassword: "different1",
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("Error user already exists", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("Register", mock.Anything, mock.AnythingOfType("dto.RegistrationUser")).Return(serviceDto.UserInfo{}, common.ErrorExistingUser)

		_, err := newHandler(m).CreateUser(context.Background(), &pb.CreateRequest{
			Email:            "user@mail.ru",
			Password:         "password1",
			RepeatedPassword: "password1",
		})

		assertGRPCCode(t, err, codes.AlreadyExists)
	})

	t.Run("Error internal service failure", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("Register", mock.Anything, mock.AnythingOfType("dto.RegistrationUser")).Return(serviceDto.UserInfo{}, errors.New("db error"))

		_, err := newHandler(m).CreateUser(context.Background(), &pb.CreateRequest{
			Email:            "user@mail.ru",
			Password:         "password1",
			RepeatedPassword: "password1",
		})

		assertGRPCCode(t, err, codes.Internal)
	})
}

func TestGetUserLink(t *testing.T) {
	t.Run("Success get user link", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("GetUserLink", mock.Anything, "user@mail.ru").Return(fixedUUID.String(), nil)

		resp, err := newHandler(m).GetUserLink(context.Background(), &pb.GetUserLinkRequest{
			Email: "user@mail.ru",
		})

		assert.NoError(t, err)
		assert.Equal(t, fixedUUID.String(), resp.UserLink)
	})

	t.Run("Error email not found", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("GetUserLink", mock.Anything, "unknown@mail.ru").Return("", common.ErrorNonexistentEmail)

		_, err := newHandler(m).GetUserLink(context.Background(), &pb.GetUserLinkRequest{
			Email: "unknown@mail.ru",
		})

		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("Error internal failure", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("GetUserLink", mock.Anything, "user@mail.ru").Return("", errors.New("db error"))

		_, err := newHandler(m).GetUserLink(context.Background(), &pb.GetUserLinkRequest{
			Email: "user@mail.ru",
		})

		assertGRPCCode(t, err, codes.Internal)
	})
}

func TestResetPassword(t *testing.T) {
	t.Run("Success reset password", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("ResetPassword", mock.Anything, serviceDto.ResetPasswordInfo{
			UserLink:    fixedUUID.String(),
			NewPassword: "newpassword1",
		}).Return(nil)

		resp, err := newHandler(m).ResetPassword(context.Background(), &pb.ResetPasswordRequest{
			UserLink:         fixedUUID.String(),
			Password:         "newpassword1",
			RepeatedPassword: "newpassword1",
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("Error passwords do not match", func(t *testing.T) {
		_, err := newHandler(mockAuthSrv.NewAuthService(t)).ResetPassword(context.Background(), &pb.ResetPasswordRequest{
			UserLink:         fixedUUID.String(),
			Password:         "newpassword1",
			RepeatedPassword: "different1",
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("Error user not found", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("ResetPassword", mock.Anything, mock.Anything).Return(common.ErrorNonexistentUser)

		_, err := newHandler(m).ResetPassword(context.Background(), &pb.ResetPasswordRequest{
			UserLink:         fixedUUID.String(),
			Password:         "newpassword1",
			RepeatedPassword: "newpassword1",
		})

		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("Error internal failure", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("ResetPassword", mock.Anything, mock.Anything).Return(errors.New("db error"))

		_, err := newHandler(m).ResetPassword(context.Background(), &pb.ResetPasswordRequest{
			UserLink:         fixedUUID.String(),
			Password:         "newpassword1",
			RepeatedPassword: "newpassword1",
		})

		assertGRPCCode(t, err, codes.Internal)
	})
}

func TestGetProfile(t *testing.T) {
	t.Run("Success get profile", func(t *testing.T) {
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

	t.Run("Error invalid UUID", func(t *testing.T) {
		_, err := newHandler(mockAuthSrv.NewAuthService(t)).GetProfile(context.Background(), &pb.UserLinkRequest{
			UserLink: "not-a-uuid",
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("Error user not found", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("GetProfile", mock.Anything, fixedUUID).Return(serviceDto.UserInfo{}, common.ErrorNonexistentUser)

		_, err := newHandler(m).GetProfile(context.Background(), &pb.UserLinkRequest{
			UserLink: fixedUUID.String(),
		})

		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("Error internal failure", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("GetProfile", mock.Anything, fixedUUID).Return(serviceDto.UserInfo{}, errors.New("db error"))

		_, err := newHandler(m).GetProfile(context.Background(), &pb.UserLinkRequest{
			UserLink: fixedUUID.String(),
		})

		assertGRPCCode(t, err, codes.Internal)
	})
}

func TestUpdateProfile(t *testing.T) {
	t.Run("Success update profile", func(t *testing.T) {
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

	t.Run("Error invalid UUID", func(t *testing.T) {
		_, err := newHandler(mockAuthSrv.NewAuthService(t)).UpdateProfile(context.Background(), &pb.UpdateProfileRequest{
			UserLink:    "bad-uuid",
			DisplayName: "Name",
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("Error display name too long", func(t *testing.T) {
		longName := string(make([]byte, 200))
		_, err := newHandler(mockAuthSrv.NewAuthService(t)).UpdateProfile(context.Background(), &pb.UpdateProfileRequest{
			UserLink:    fixedUUID.String(),
			DisplayName: longName,
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("Error internal failure", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("UpdateProfile", mock.Anything, mock.AnythingOfType("dto.UpdatedUserInfo")).Return(errors.New("db error"))

		_, err := newHandler(m).UpdateProfile(context.Background(), &pb.UpdateProfileRequest{
			UserLink:    fixedUUID.String(),
			DisplayName: "Name",
		})

		assertGRPCCode(t, err, codes.Internal)
	})
}

func TestUpdateAvatar(t *testing.T) {
	t.Run("Success update avatar with JPEG", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("UpdateAvatar", mock.Anything, mock.AnythingOfType("dto.UpdatedAvatar")).Return("https://cdn.example.com/avatar.jpg", nil)

		resp, err := newHandler(m).UpdateAvatar(context.Background(), &pb.UpdateAvatarRequest{
			UserLink: fixedUUID.String(),
			FileData: jpegMagic,
		})

		assert.NoError(t, err)
		assert.Contains(t, resp.AvatarUrl, "avatar.jpg")
	})

	t.Run("Error invalid UUID", func(t *testing.T) {
		_, err := newHandler(mockAuthSrv.NewAuthService(t)).UpdateAvatar(context.Background(), &pb.UpdateAvatarRequest{
			UserLink: "bad-uuid",
			FileData: jpegMagic,
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("Error empty file", func(t *testing.T) {
		_, err := newHandler(mockAuthSrv.NewAuthService(t)).UpdateAvatar(context.Background(), &pb.UpdateAvatarRequest{
			UserLink: fixedUUID.String(),
			FileData: []byte{},
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("Error file too large", func(t *testing.T) {
		bigFile := make([]byte, 6*1024*1024)
		copy(bigFile, jpegMagic)

		_, err := newHandler(mockAuthSrv.NewAuthService(t)).UpdateAvatar(context.Background(), &pb.UpdateAvatarRequest{
			UserLink: fixedUUID.String(),
			FileData: bigFile,
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("Error unsupported MIME type", func(t *testing.T) {
		textData := []byte("this is plain text, not an image")

		_, err := newHandler(mockAuthSrv.NewAuthService(t)).UpdateAvatar(context.Background(), &pb.UpdateAvatarRequest{
			UserLink: fixedUUID.String(),
			FileData: textData,
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("Error service failure", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("UpdateAvatar", mock.Anything, mock.AnythingOfType("dto.UpdatedAvatar")).Return("", errors.New("s3 error"))

		_, err := newHandler(m).UpdateAvatar(context.Background(), &pb.UpdateAvatarRequest{
			UserLink: fixedUUID.String(),
			FileData: jpegMagic,
		})

		assertGRPCCode(t, err, codes.Internal)
	})
}

func TestDeleteAvatar(t *testing.T) {
	t.Run("Success delete avatar", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("DeleteAvatar", mock.Anything, fixedUUID).Return(nil)

		resp, err := newHandler(m).DeleteAvatar(context.Background(), &pb.UserLinkRequest{
			UserLink: fixedUUID.String(),
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("Error invalid UUID", func(t *testing.T) {
		_, err := newHandler(mockAuthSrv.NewAuthService(t)).DeleteAvatar(context.Background(), &pb.UserLinkRequest{
			UserLink: "bad-uuid",
		})

		assertGRPCCode(t, err, codes.InvalidArgument)
	})

	t.Run("Error user not found", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("DeleteAvatar", mock.Anything, fixedUUID).Return(common.ErrorNonexistentUser)

		_, err := newHandler(m).DeleteAvatar(context.Background(), &pb.UserLinkRequest{
			UserLink: fixedUUID.String(),
		})

		assertGRPCCode(t, err, codes.NotFound)
	})

	t.Run("Error internal failure", func(t *testing.T) {
		m := mockAuthSrv.NewAuthService(t)
		m.On("DeleteAvatar", mock.Anything, fixedUUID).Return(errors.New("s3 error"))

		_, err := newHandler(m).DeleteAvatar(context.Background(), &pb.UserLinkRequest{
			UserLink: fixedUUID.String(),
		})

		assertGRPCCode(t, err, codes.Internal)
	})
}

func assertGRPCCode(t *testing.T, err error, expected codes.Code) {
	t.Helper()
	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok, "error must be a gRPC status error")
	assert.Equal(t, expected, st.Code())
}
