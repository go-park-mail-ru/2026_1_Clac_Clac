package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidCSRFToken = errors.New("csrf token invalid")
	ErrCSRFTokenExpired = errors.New("csrf token expired")
)

const (
	csrfTokenSeparator  = ":"
	csrfDefaultTTL      = 24 * time.Hour
	csrfTokenPartsCount = 2
)

type CSRFConfig struct {
	Secret string
	TTL    time.Duration
}

type CSRF struct {
	cfg CSRFConfig
}

func NewCSRF(cfg CSRFConfig) *CSRF {
	if cfg.TTL == 0 {
		cfg.TTL = csrfDefaultTTL
	}
	return &CSRF{cfg: cfg}
}

func (c *CSRF) GetExpireTime() time.Time {
	return time.Now().Add(c.cfg.TTL)
}

func (c *CSRF) Generate(_ context.Context, sessionID string, expireAt int64) (string, error) {
	mac := hmac.New(sha256.New, []byte(c.cfg.Secret))
	_, err := fmt.Fprintf(mac, "%s|%d", sessionID, expireAt)
	if err != nil {
		return "", fmt.Errorf("hmac write: %w", err)
	}

	sig := hex.EncodeToString(mac.Sum(nil))
	return strconv.FormatInt(expireAt, 10) + csrfTokenSeparator + sig, nil
}

func (a *Service) CheckCSRFToken(ctx context.Context, sessionId string, token string) error {
	tokenData := strings.Split(token, ":")
	if len(tokenData) != csrfTokenPartsCount {
		return ErrInvalidCSRFToken
	}

	expireTime, err := strconv.ParseInt(tokenData[1], csrfTokenExpireTimeConvertationBase, csrfTokenExpireTimeConvertationTypeSize)
	if err != nil {
		return ErrCannotParseExpireTimeCSRFToken
	}

	if expireTime < time.Now().Unix() {
		return ErrCSRFTokenExpired
	}

	h := hmac.New(sha256.New, []byte(a.tools.CsrfSecret))
	data := fmt.Sprintf("%s:%d", sessionId, expireTime)
	h.Write([]byte(data))

	expected := h.Sum(nil)
	recieved, err := hex.DecodeString(tokenData[0])
	if err != nil {
		return ErrCannotDecodeRecievedCSRFToken
	}

	if !hmac.Equal(recieved, expected) {
		return ErrCSRFTokensDoNotEqual
	}

	return nil
}
