package api_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/api"
	"github.com/stretchr/testify/assert"
)

func TestNewSessionCookie(t *testing.T) {
	key := "session_id"
	value := "abc123"
	expires := time.Now().Add(24 * time.Hour)

	cookie := api.NewSessionCookie(key, value, expires)

	assert.Equal(t, key, cookie.Name)
	assert.Equal(t, value, cookie.Value)
	assert.True(t, cookie.Expires.Equal(expires))
	assert.True(t, cookie.Secure)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
}

func TestNewCSRFCookie(t *testing.T) {
	name := "csrf_token"
	value := "token-xyz"
	expireTime := time.Now().Add(1 * time.Hour)

	cookie := api.NewCSRFCookie(name, value, expireTime)

	assert.Equal(t, name, cookie.Name)
	assert.Equal(t, value, cookie.Value)
	assert.True(t, cookie.Expires.Equal(expireTime))
	assert.True(t, cookie.Secure)
	assert.False(t, cookie.HttpOnly)
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
}

func TestNewExpiredCookie(t *testing.T) {
	key := "session_id"

	cookie := api.NewExpiredCookie(key)

	assert.Equal(t, key, cookie.Name)
	assert.Empty(t, cookie.Value)
	assert.Equal(t, -1, cookie.MaxAge)
	assert.True(t, cookie.HttpOnly)
	assert.True(t, cookie.Secure)
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
}

func TestNewSessionCookie_DifferentExpiry(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(1 * time.Hour)

	cookiePast := api.NewSessionCookie("key", "val", past)
	assert.True(t, cookiePast.Expires.Before(time.Now()))

	cookieFuture := api.NewSessionCookie("key", "val", future)
	assert.True(t, cookieFuture.Expires.After(time.Now()))
}
