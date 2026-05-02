package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCSRF() *CSRF {
	return NewCSRF(CSRFConfig{
		Secret:                         "test-secret",
		TTL:                            time.Hour,
		ExpireTimeConvertationBase:     10,
		ExpireTimeConvertationTypeSize: 64,
		PartsCount:                     2,
	})
}

func TestCSRFGenerate(t *testing.T) {
	svc := newTestCSRF()
	ctx := context.Background()

	expireAt := time.Now().Add(time.Hour).Unix()
	token, err := svc.Generate(ctx, "session-abc", expireAt)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestCSRFCheck(t *testing.T) {
	svc := newTestCSRF()
	ctx := context.Background()

	validExpireAt := time.Now().Add(time.Hour).Unix()
	validToken, err := svc.Generate(ctx, "session-correct", validExpireAt)
	require.NoError(t, err)

	expiredExpireAt := time.Now().Add(-time.Hour).Unix()
	expiredToken, err := svc.Generate(ctx, "session-abc", expiredExpireAt)
	require.NoError(t, err)

	tests := []struct {
		name        string
		sessionID   string
		token       string
		expectedErr error
	}{
		{
			name:        "Success",
			sessionID:   "session-correct",
			token:       validToken,
			expectedErr: nil,
		},
		{
			name:        "WrongSession",
			sessionID:   "session-wrong",
			token:       validToken,
			expectedErr: common.ErrCSRFTokensDoNotEqual,
		},
		{
			name:        "Expired",
			sessionID:   "session-abc",
			token:       expiredToken,
			expectedErr: common.ErrCSRFTokenExpired,
		},
		{
			name:        "InvalidFormat_NoColon",
			sessionID:   "session",
			token:       "no-colon-here",
			expectedErr: common.ErrInvalidCSRFToken,
		},
		{
			name:        "InvalidFormat_BadExpireAt",
			sessionID:   "session",
			token:       "notanumber:abc",
			expectedErr: common.ErrCannotParseExpireTimeCSRFToken,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.Check(ctx, tc.sessionID, tc.token)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCSRFGetExpireTime(t *testing.T) {
	svc := newTestCSRF()
	before := time.Now()
	exp := svc.GetExpireTime(context.Background())
	assert.True(t, exp.After(before))
	assert.True(t, exp.Before(before.Add(2*time.Hour)))
}
