package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	mockAuthClient "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/usecase/mock_auth_client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	fixedSessionID = "12345667"
)

func TestCreateSession(t *testing.T) {
	tests := []struct {
		name              string
		mockBehavior      func(m *mockAuthClient.AuthClient)
		expectedSessionID string
		expectError       bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockAuthClient.AuthClient) {
				m.On("CreateSession", context.Background(), fixedUserLink).Return(fixedSessionID, nil)
			},
			expectedSessionID: fixedSessionID,
			expectError:       false,
		},
		{
			name: "Error",
			mockBehavior: func(m *mockAuthClient.AuthClient) {
				m.On("CreateSession", context.Background(), fixedUserLink).Return("", errors.New("redis down"))
			},
			expectedSessionID: "",
			expectError:       true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockAuthClient.NewAuthClient(t)
			tc.mockBehavior(m)

			sid, err := NewAuth(m).CreateSession(context.Background(), fixedUserLink)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedSessionID, sid)
			}
		})
	}
}

func TestDeleteSession(t *testing.T) {
	tests := []struct {
		name         string
		mockBehavior func(m *mockAuthClient.AuthClient)
		expectedErr  error
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockAuthClient.AuthClient) {
				m.On("DeleteSession", context.Background(), fixedSessionID).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "SessionNotFound",
			mockBehavior: func(m *mockAuthClient.AuthClient) {
				m.On("DeleteSession", context.Background(), fixedSessionID).Return(common.ErrorSessionNotFound)
			},
			expectedErr: common.ErrorSessionNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockAuthClient.NewAuthClient(t)
			tc.mockBehavior(m)

			err := NewAuth(m).DeleteSession(context.Background(), fixedSessionID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.expectedErr))
			} else {
				require.NoError(t, err)
			}
		})
	}
}
