package usecase

import (
	"bytes"
	"context"
	"errors"
	"testing"

	mockAppealClient "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/usecase/mock_appeal_client"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	fixedAppealUserLink = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	fixedAppealLink     = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

func TestAppealUsecase_CreateAppeal(t *testing.T) {
	info := domain.CreateAppealInfo{
		UserLink:    fixedAppealUserLink,
		Email:       "user@example.com",
		DisplayName: "Alice",
		Description: "my issue",
		Category:    "CATEGORY_TECHNICAL",
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mockAppealClient.AppealClient)
		expectErr    bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockAppealClient.AppealClient) {
				m.On("CreateAppeal", context.Background(), info).Return(fixedAppealLink, nil)
			},
			expectErr: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockAppealClient.AppealClient) {
				m.On("CreateAppeal", context.Background(), info).Return(uuid.Nil, errors.New("grpc error"))
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockAppealClient.NewAppealClient(t)
			tc.mockBehavior(m)

			link, err := NewAppeal(m).CreateAppeal(context.Background(), info)

			if tc.expectErr {
				require.Error(t, err)
				assert.Equal(t, uuid.Nil, link)
			} else {
				require.NoError(t, err)
				assert.Equal(t, fixedAppealLink, link)
			}
		})
	}
}

func TestAppealUsecase_GetAppeal(t *testing.T) {
	appeals := []domain.AppealInfo{
		{AppealID: 1, AppealLink: fixedAppealLink, Email: "a@b.com"},
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mockAppealClient.AppealClient)
		expectErr    bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockAppealClient.AppealClient) {
				m.On("GetAppeal", context.Background(), fixedAppealUserLink).Return("user", appeals, nil)
			},
			expectErr: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockAppealClient.AppealClient) {
				m.On("GetAppeal", context.Background(), fixedAppealUserLink).Return("", []domain.AppealInfo{}, errors.New("grpc error"))
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockAppealClient.NewAppealClient(t)
			tc.mockBehavior(m)

			role, got, err := NewAppeal(m).GetAppeal(context.Background(), fixedAppealUserLink)

			if tc.expectErr {
				require.Error(t, err)
				assert.Empty(t, role)
				assert.Empty(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "user", role)
				assert.Equal(t, appeals, got)
			}
		})
	}
}

func TestAppealUsecase_UploadAttachment(t *testing.T) {
	info := domain.UploadAttachmentInfo{
		UserLink:   fixedAppealUserLink,
		AppealLink: fixedAppealLink,
		Filename:   "photo.png",
	}
	attachURL := "https://cdn.example.com/photo.png"
	fileContent := bytes.NewReader([]byte("fake-image"))

	tests := []struct {
		name         string
		mockBehavior func(m *mockAppealClient.AppealClient)
		expectErr    bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockAppealClient.AppealClient) {
				m.On("UploadAttachment", context.Background(), info, fileContent).Return(attachURL, nil)
			},
			expectErr: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockAppealClient.AppealClient) {
				m.On("UploadAttachment", context.Background(), info, fileContent).Return("", errors.New("storage error"))
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockAppealClient.NewAppealClient(t)
			tc.mockBehavior(m)

			url, err := NewAppeal(m).UploadAttachment(context.Background(), info, fileContent)

			if tc.expectErr {
				require.Error(t, err)
				assert.Empty(t, url)
			} else {
				require.NoError(t, err)
				assert.Equal(t, attachURL, url)
			}
		})
	}
}

func TestAppealUsecase_DeleteAppeal(t *testing.T) {
	info := domain.DeleteInfo{
		UserLink:   fixedAppealUserLink,
		AppealLink: fixedAppealLink,
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mockAppealClient.AppealClient)
		expectErr    bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockAppealClient.AppealClient) {
				m.On("DeleteAppeal", context.Background(), info).Return(nil)
			},
			expectErr: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockAppealClient.AppealClient) {
				m.On("DeleteAppeal", context.Background(), info).Return(errors.New("grpc error"))
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockAppealClient.NewAppealClient(t)
			tc.mockBehavior(m)

			err := NewAppeal(m).DeleteAppeal(context.Background(), info)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAppealUsecase_GetStats(t *testing.T) {
	stats := domain.AppealsStats{
		OpenAppeals:   3,
		InWorkAppeals: 1,
		CloseAppeals:  5,
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mockAppealClient.AppealClient)
		expectErr    bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockAppealClient.AppealClient) {
				m.On("GetStats", context.Background(), fixedAppealUserLink).Return(stats, nil)
			},
			expectErr: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockAppealClient.AppealClient) {
				m.On("GetStats", context.Background(), fixedAppealUserLink).Return(domain.AppealsStats{}, errors.New("grpc error"))
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockAppealClient.NewAppealClient(t)
			tc.mockBehavior(m)

			got, err := NewAppeal(m).GetStats(context.Background(), fixedAppealUserLink)

			if tc.expectErr {
				require.Error(t, err)
				assert.Equal(t, domain.AppealsStats{}, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, stats, got)
			}
		})
	}
}

func TestAppealUsecase_ChangeAppealStatus(t *testing.T) {
	info := domain.ChangeAppealStatusInfo{
		UserLink:   fixedAppealUserLink,
		AppealLink: fixedAppealLink,
		NewStatus:  "STATUS_CLOSED",
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mockAppealClient.AppealClient)
		expectErr    bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockAppealClient.AppealClient) {
				m.On("ChangeAppealStatus", context.Background(), info).Return(nil)
			},
			expectErr: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockAppealClient.AppealClient) {
				m.On("ChangeAppealStatus", context.Background(), info).Return(errors.New("grpc error"))
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockAppealClient.NewAppealClient(t)
			tc.mockBehavior(m)

			err := NewAppeal(m).ChangeAppealStatus(context.Background(), info)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
