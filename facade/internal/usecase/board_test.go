package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	mockBoardClient "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/usecase/mock_board_client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	fixedBoardLink = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	errBoardTest   = errors.New("client error")
)

func TestBoardGetBoards(t *testing.T) {
	expected := []domain.BoardInfo{
		{Link: fixedBoardLink, Name: "Board 1"},
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mockBoardClient.BoardClient)
		expected     []domain.BoardInfo
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("GetBoards", context.Background(), fixedUserLink).Return(expected, nil)
			},
			expected:    expected,
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("GetBoards", context.Background(), fixedUserLink).Return([]domain.BoardInfo(nil), errBoardTest)
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardClient.NewBoardClient(t)
			tc.mockBehavior(m)

			result, err := NewBoard(m).GetBoards(context.Background(), fixedUserLink)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestBoardGetBoard(t *testing.T) {
	req := domain.GetBoardRequest{UserLink: fixedUserLink, BoardLink: fixedBoardLink}
	expected := domain.BoardInfo{Link: fixedBoardLink, Name: "Board 1"}

	tests := []struct {
		name         string
		mockBehavior func(m *mockBoardClient.BoardClient)
		expected     domain.BoardInfo
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("GetBoard", context.Background(), req).Return(expected, nil)
			},
			expected:    expected,
			expectError: false,
		},
		{
			name: "NotFound",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("GetBoard", context.Background(), req).Return(domain.BoardInfo{}, common.ErrorNonexistentUser)
			},
			expected:    domain.BoardInfo{},
			expectError: true,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("GetBoard", context.Background(), req).Return(domain.BoardInfo{}, errBoardTest)
			},
			expected:    domain.BoardInfo{},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardClient.NewBoardClient(t)
			tc.mockBehavior(m)

			result, err := NewBoard(m).GetBoard(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestBoardCreateBoard(t *testing.T) {
	req := domain.CreateBoardRequest{UserLink: fixedUserLink, Name: "New Board", Description: "desc"}
	expected := domain.BoardInfo{Link: fixedBoardLink, Name: "New Board"}

	tests := []struct {
		name         string
		mockBehavior func(m *mockBoardClient.BoardClient)
		expected     domain.BoardInfo
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("CreateBoard", context.Background(), req).Return(expected, nil)
			},
			expected:    expected,
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("CreateBoard", context.Background(), req).Return(domain.BoardInfo{}, errBoardTest)
			},
			expected:    domain.BoardInfo{},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardClient.NewBoardClient(t)
			tc.mockBehavior(m)

			result, err := NewBoard(m).CreateBoard(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestBoardDeleteBoard(t *testing.T) {
	req := domain.GetBoardRequest{UserLink: fixedUserLink, BoardLink: fixedBoardLink}

	tests := []struct {
		name         string
		mockBehavior func(m *mockBoardClient.BoardClient)
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("DeleteBoard", context.Background(), req).Return(nil)
			},
			expectError: false,
		},
		{
			name: "NotFound",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("DeleteBoard", context.Background(), req).Return(common.ErrorNonexistentUser)
			},
			expectError: true,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("DeleteBoard", context.Background(), req).Return(errBoardTest)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardClient.NewBoardClient(t)
			tc.mockBehavior(m)

			err := NewBoard(m).DeleteBoard(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBoardUpdateBoard(t *testing.T) {
	req := domain.UpdateBoardRequest{UserLink: fixedUserLink, BoardLink: fixedBoardLink, Name: "Updated"}

	tests := []struct {
		name         string
		mockBehavior func(m *mockBoardClient.BoardClient)
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("UpdateBoard", context.Background(), req).Return(nil)
			},
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("UpdateBoard", context.Background(), req).Return(errBoardTest)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardClient.NewBoardClient(t)
			tc.mockBehavior(m)

			err := NewBoard(m).UpdateBoard(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBoardUploadBackground(t *testing.T) {
	bgReq := domain.UploadBackgroundRequest{
		UserLink:  fixedUserLink,
		BoardLink: fixedBoardLink,
		Filename:  "bg.png",
	}
	expected := domain.UploadBackgroundResponse{BackgroundKey: "key/bg.png"}
	reader := strings.NewReader("image data")

	tests := []struct {
		name         string
		mockBehavior func(m *mockBoardClient.BoardClient)
		expected     domain.UploadBackgroundResponse
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("UploadBackground", context.Background(), bgReq, reader).Return(expected, nil)
			},
			expected:    expected,
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("UploadBackground", context.Background(), bgReq, reader).Return(domain.UploadBackgroundResponse{}, errBoardTest)
			},
			expected:    domain.UploadBackgroundResponse{},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardClient.NewBoardClient(t)
			tc.mockBehavior(m)

			result, err := NewBoard(m).UploadBackground(context.Background(), bgReq, reader)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestBoardGetMembers(t *testing.T) {
	req := domain.GetMembersRequest{UserLink: fixedUserLink, BoardLink: fixedBoardLink}
	expected := domain.GetMembersResponse{Members: []domain.MemberInfo{{Link: fixedUserLink, Role: "editor"}}}

	tests := []struct {
		name         string
		mockBehavior func(m *mockBoardClient.BoardClient)
		expected     domain.GetMembersResponse
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("GetMembers", context.Background(), req).Return(expected, nil)
			},
			expected:    expected,
			expectError: false,
		},
		{
			name: "NotFound",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("GetMembers", context.Background(), req).Return(domain.GetMembersResponse{}, common.ErrorNonexistentUser)
			},
			expected:    domain.GetMembersResponse{},
			expectError: true,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("GetMembers", context.Background(), req).Return(domain.GetMembersResponse{}, errBoardTest)
			},
			expected:    domain.GetMembersResponse{},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardClient.NewBoardClient(t)
			tc.mockBehavior(m)

			result, err := NewBoard(m).GetMembers(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestBoardCreateInviteUsecase(t *testing.T) {
	req := domain.CreateInviteRequest{
		UserLink:    fixedUserLink,
		BoardLink:   fixedBoardLink,
		DefaultRole: "editor",
	}
	expected := domain.CreateInviteResponse{
		InviteLink:  uuid.New().String(),
		BoardLink:   fixedBoardLink.String(),
		DefaultRole: "editor",
		Status:      "active",
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mockBoardClient.BoardClient)
		expected     domain.CreateInviteResponse
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("CreateInvite", context.Background(), req).Return(expected, nil)
			},
			expected:    expected,
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("CreateInvite", context.Background(), req).Return(domain.CreateInviteResponse{}, errBoardTest)
			},
			expected:    domain.CreateInviteResponse{},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardClient.NewBoardClient(t)
			tc.mockBehavior(m)

			result, err := NewBoard(m).CreateInvite(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestBoardAcceptInviteUsecase(t *testing.T) {
	req := domain.AcceptInviteRequest{
		InviteLink: uuid.New().String(),
		UserLink:   fixedUserLink,
	}

	tests := []struct {
		name            string
		mockBehavior    func(m *mockBoardClient.BoardClient)
		expectBoardLink string
		expectRole      string
		expectError     bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("AcceptInvite", context.Background(), req).Return(fixedBoardLink.String(), "editor", nil)
			},
			expectBoardLink: fixedBoardLink.String(),
			expectRole:      "editor",
			expectError:     false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("AcceptInvite", context.Background(), req).Return("", "", errBoardTest)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardClient.NewBoardClient(t)
			tc.mockBehavior(m)

			boardLink, role, err := NewBoard(m).AcceptInvite(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectBoardLink, boardLink)
				assert.Equal(t, tc.expectRole, role)
			}
		})
	}
}

func TestBoardCloseInviteUsecase(t *testing.T) {
	req := domain.CloseInviteRequest{
		UserLink:   fixedUserLink,
		InviteLink: uuid.New().String(),
	}

	tests := []struct {
		name        string
		mockBehavior func(m *mockBoardClient.BoardClient)
		expectError bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("CloseInvite", context.Background(), req).Return(nil)
			},
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("CloseInvite", context.Background(), req).Return(errBoardTest)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardClient.NewBoardClient(t)
			tc.mockBehavior(m)

			err := NewBoard(m).CloseInvite(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBoardUpdateMemberRoleUsecase(t *testing.T) {
	req := domain.UpdateMemberRoleRequest{
		UserLink:       fixedUserLink,
		BoardLink:      fixedBoardLink,
		TargetUserLink: fixedUserLink,
		NewRole:        "editor",
	}

	tests := []struct {
		name        string
		mockBehavior func(m *mockBoardClient.BoardClient)
		expectError bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("UpdateMemberRole", context.Background(), req).Return(nil)
			},
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("UpdateMemberRole", context.Background(), req).Return(errBoardTest)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardClient.NewBoardClient(t)
			tc.mockBehavior(m)

			err := NewBoard(m).UpdateMemberRole(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBoardRemoveMemberFromBoardUsecase(t *testing.T) {
	req := domain.RemoveMemberRequest{
		UserLink:       fixedUserLink,
		BoardLink:      fixedBoardLink,
		TargetUserLink: fixedUserLink,
	}

	tests := []struct {
		name        string
		mockBehavior func(m *mockBoardClient.BoardClient)
		expectError bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("RemoveMemberFromBoard", context.Background(), req).Return(nil)
			},
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("RemoveMemberFromBoard", context.Background(), req).Return(errBoardTest)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardClient.NewBoardClient(t)
			tc.mockBehavior(m)

			err := NewBoard(m).RemoveMemberFromBoard(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBoardGetActiveInvitesUsecase(t *testing.T) {
	expected := []domain.InviteInfo{
		{InviteLink: uuid.New().String(), BoardLink: fixedBoardLink.String(), DefaultRole: "editor", Status: "active"},
	}

	tests := []struct {
		name        string
		mockBehavior func(m *mockBoardClient.BoardClient)
		expected    []domain.InviteInfo
		expectError bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("GetActiveInvites", context.Background(), fixedUserLink, fixedBoardLink).Return(expected, nil)
			},
			expected:    expected,
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockBoardClient.BoardClient) {
				m.On("GetActiveInvites", context.Background(), fixedUserLink, fixedBoardLink).Return(nil, errBoardTest)
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockBoardClient.NewBoardClient(t)
			tc.mockBehavior(m)

			result, err := NewBoard(m).GetActiveInvites(context.Background(), fixedUserLink, fixedBoardLink)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
