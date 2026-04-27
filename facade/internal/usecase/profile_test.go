package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	mockProfileClient "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/usecase/mock_profile_client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProfile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := mockProfileClient.NewProfileClient(t)
		expected := domain.FullInfoUser{UserLink: fixedUserLink, Email: "test@mail.ru", DisplayName: "Test"}
		m.On("GetProfile", context.Background(), fixedUserLink).Return(expected, nil)

		result, err := NewProfile(m).GetProfile(context.Background(), fixedUserLink)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		m := mockProfileClient.NewProfileClient(t)
		m.On("GetProfile", context.Background(), fixedUserLink).Return(domain.FullInfoUser{}, common.ErrorNonexistentUser)

		_, err := NewProfile(m).GetProfile(context.Background(), fixedUserLink)
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrorNonexistentUser))
	})
}

func TestUpdateProfile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := mockProfileClient.NewProfileClient(t)
		info := domain.UpdatedInfo{UserLink: fixedUserLink, DisplayName: "NewName", Description: "desc"}
		m.On("UpdateProfile", context.Background(), info).Return(nil)

		err := NewProfile(m).UpdateProfile(context.Background(), info)
		require.NoError(t, err)
	})

	t.Run("MissingField", func(t *testing.T) {
		m := mockProfileClient.NewProfileClient(t)
		info := domain.UpdatedInfo{UserLink: fixedUserLink}
		m.On("UpdateProfile", context.Background(), info).Return(common.ErrorMissingRequiredField)

		err := NewProfile(m).UpdateProfile(context.Background(), info)
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrorMissingRequiredField))
	})
}

func TestUpdateAvatar(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := mockProfileClient.NewProfileClient(t)
		info := domain.AvatarInfo{UserLink: fixedUserLink, FileData: []byte{0xFF, 0xD8}, ContentType: "image/jpeg"}
		m.On("UpdateAvatar", context.Background(), info).Return("https://cdn.example.com/avatar.jpg", nil)

		url, err := NewProfile(m).UpdateAvatar(context.Background(), info)
		require.NoError(t, err)
		assert.Equal(t, "https://cdn.example.com/avatar.jpg", url)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		m := mockProfileClient.NewProfileClient(t)
		info := domain.AvatarInfo{UserLink: fixedUserLink}
		m.On("UpdateAvatar", context.Background(), info).Return("", common.ErrorNonexistentUser)

		_, err := NewProfile(m).UpdateAvatar(context.Background(), info)
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrorNonexistentUser))
	})
}

func TestDeleteAvatar(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := mockProfileClient.NewProfileClient(t)
		m.On("DeleteAvatar", context.Background(), fixedUserLink).Return(nil)

		err := NewProfile(m).DeleteAvatar(context.Background(), fixedUserLink)
		require.NoError(t, err)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		m := mockProfileClient.NewProfileClient(t)
		m.On("DeleteAvatar", context.Background(), fixedUserLink).Return(common.ErrorNonexistentUser)

		err := NewProfile(m).DeleteAvatar(context.Background(), fixedUserLink)
		require.Error(t, err)
		assert.True(t, errors.Is(err, common.ErrorNonexistentUser))
	})
}
