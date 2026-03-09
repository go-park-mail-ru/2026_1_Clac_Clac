package api

import (
	"net/http"
	"time"
)

const (
	defaultHttpOnly = true
	defaultPath     = "/"
	defaultMaxAge   = -1
	zeroValue       = ""
)

func NewCookie(key string, value string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     key,
		Value:    value,
		Expires:  expires,
		HttpOnly: defaultHttpOnly,
		Path:     defaultPath,
	}
}

func NewExpiredCookie(key string) *http.Cookie {
	return &http.Cookie{
		Name:     key,
		Value:    zeroValue,
		MaxAge:   defaultMaxAge,
		HttpOnly: defaultHttpOnly,
		Path:     defaultPath,
	}
}
