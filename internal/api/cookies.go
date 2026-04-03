package api

import (
	"net/http"
	"time"
)

func NewSessionCookie(key string, value string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     key,
		Value:    value,
		Expires:  expires,
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
}

func NewCSRFCookie(name string, value string, expireTime time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Secure:   true,
		HttpOnly: false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  expireTime,
	}
}

func NewExpiredCookie(key string) *http.Cookie {
	return &http.Cookie{
		Name:     key,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
}
