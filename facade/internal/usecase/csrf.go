package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
)


const (
	csrfTokenSeparator = ":"
	csrfDefaultTTL     = 24 * time.Hour
)

type CSRFConfig struct {
	Secret                         string
	TTL                            time.Duration
	ExpireTimeConvertationBase     int
	ExpireTimeConvertationTypeSize int
	PartsCount                     int
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

func (c *CSRF) GetExpireTime(ctx context.Context) time.Time {
	return time.Now().Add(c.cfg.TTL)
}

func (c *CSRF) Generate(ctx context.Context, sessionId string, expireTime int64) (string, error) {
	h := hmac.New(sha256.New, []byte(c.cfg.Secret))
	data := fmt.Sprintf("%s:%d", sessionId, expireTime)
	h.Write([]byte(data))

	token := fmt.Sprintf("%s:%s", hex.EncodeToString(h.Sum(nil)), strconv.FormatInt(expireTime, c.cfg.ExpireTimeConvertationBase))

	return token, nil
}

func (c *CSRF) Check(ctx context.Context, sessionId string, token string) error {
	tokenData := strings.Split(token, csrfTokenSeparator)
	if len(tokenData) != c.cfg.PartsCount {
		return common.ErrInvalidCSRFToken
	}

	expireTime, err := strconv.ParseInt(tokenData[1], c.cfg.ExpireTimeConvertationBase, c.cfg.ExpireTimeConvertationTypeSize)
	if err != nil {
		return common.ErrCannotParseExpireTimeCSRFToken
	}

	if expireTime < time.Now().Unix() {
		return common.ErrCSRFTokenExpired
	}

	h := hmac.New(sha256.New, []byte(c.cfg.Secret))
	data := fmt.Sprintf("%s:%d", sessionId, expireTime)
	h.Write([]byte(data))

	expected := h.Sum(nil)
	received, err := hex.DecodeString(tokenData[0])
	if err != nil {
		return common.ErrCannotDecodeReceivedCSRFToken
	}

	if !hmac.Equal(received, expected) {
		return common.ErrCSRFTokensDoNotEqual
	}

	return nil
}
