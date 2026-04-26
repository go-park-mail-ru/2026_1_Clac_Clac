package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCSRF() *CSRF {
	return NewCSRF(CSRFConfig{Secret: "test-secret", TTL: time.Hour})
}

func TestCSRFGenerateAndCheck(t *testing.T) {
	svc := newTestCSRF()
	ctx := context.Background()
	sessionID := "session-abc"

	expireAt := time.Now().Add(time.Hour).Unix()
	token, err := svc.Generate(ctx, sessionID, expireAt)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	err = svc.Check(ctx, sessionID, token)
	assert.NoError(t, err)
}

func TestCSRFCheckWrongSession(t *testing.T) {
	svc := newTestCSRF()
	ctx := context.Background()

	expireAt := time.Now().Add(time.Hour).Unix()
	token, err := svc.Generate(ctx, "session-correct", expireAt)
	require.NoError(t, err)

	err = svc.Check(ctx, "session-wrong", token)
	assert.ErrorIs(t, err, ErrInvalidCSRFToken)
}

func TestCSRFCheckExpired(t *testing.T) {
	svc := newTestCSRF()
	ctx := context.Background()

	expireAt := time.Now().Add(-time.Hour).Unix()
	token, err := svc.Generate(ctx, "session-abc", expireAt)
	require.NoError(t, err)

	err = svc.Check(ctx, "session-abc", token)
	assert.ErrorIs(t, err, ErrCSRFTokenExpired)
}

func TestCSRFCheckInvalidFormat(t *testing.T) {
	svc := newTestCSRF()
	ctx := context.Background()

	err := svc.Check(ctx, "session", "no-colon-here")
	assert.ErrorIs(t, err, ErrInvalidCSRFToken)
}

func TestCSRFCheckBadExpireAt(t *testing.T) {
	svc := newTestCSRF()
	ctx := context.Background()

	err := svc.Check(ctx, "session", "notanumber:abc")
	assert.ErrorIs(t, err, ErrInvalidCSRFToken)
}

func TestCSRFGetExpireTime(t *testing.T) {
	svc := newTestCSRF()
	before := time.Now()
	exp := svc.GetExpireTime()
	assert.True(t, exp.After(before))
	assert.True(t, exp.Before(before.Add(2*time.Hour)))
}
