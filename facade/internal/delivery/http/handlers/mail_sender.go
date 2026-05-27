package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/dto"
	handlerCommon "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	sentryLogger "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/logger"
	"github.com/google/uuid"
	"github.com/mailru/easyjson"
	"github.com/rs/zerolog"
)

const (
	nameCoolDown = "recovery_email"
)

type MailSenderUsecase interface {
	SendRecoveryCode(ctx context.Context, userLink uuid.UUID, email string) error
	CheckRecoveryCode(ctx context.Context, tokenID string) error
	ExchangeTokenForUser(ctx context.Context, resetToken domain.ResetToken) (uuid.UUID, error)
}

type GeterUserLink interface {
	GetUserLink(ctx context.Context, email string) (uuid.UUID, error)
}

type CoolDownUsecase interface {
	CheckCoolDown(ctx context.Context, cooldown domain.Cooldown) (domain.CooldownResult, error)
}

type MailSenderConfig struct {
	CoolDownExpirationSec int64
}

type MailSender struct {
	mailSender    MailSenderUsecase
	coolDown      CoolDownUsecase
	geterUserLink GeterUserLink
	cfg           MailSenderConfig
}

func NewMailSender(mailSender MailSenderUsecase, coolDown CoolDownUsecase, geterUserLink GeterUserLink, cfg MailSenderConfig) *MailSender {
	return &MailSender{
		mailSender:    mailSender,
		coolDown:      coolDown,
		geterUserLink: geterUserLink,
		cfg:           cfg,
	}
}

// SendRecoveryEmail отправляет код восстановления пароля
//
//	@Summary	Отправить код восстановления
//	@Tags		Auth
//	@Accept		json
//	@Produce	json
//	@Param		input	body		dto.PasswordRecoveryRequest	true	"Email пользователя"
//	@Success	200		{object}	api.Response				"OK"
//	@Failure	400		{object}	api.ErrorResponse			"Invalid email"
//	@Failure	404		{object}	api.ErrorResponse			"User does not exist"
//	@Failure	429		{object}	api.ErrorResponse			"Too many requests"
//	@Failure	500		{object}	api.ErrorResponse			"Cannot send recovery code"
//	@Router		/forgot-password [post]
func (ms *MailSender) SendRecoveryEmail(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request dto.PasswordRecoveryRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &request); err != nil {
		_, _ = api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	ok := ValidateEmail(request.Email)
	if !ok {
		_, _ = api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidEmailOrPassword.Error())
		return
	}

	result, err := ms.coolDown.CheckCoolDown(r.Context(), domain.Cooldown{
		Name:        nameCoolDown,
		Email:       request.Email,
		ExpirationS: ms.cfg.CoolDownExpirationSec,
	})
	if err != nil {
		sentryLogger.CaptureFromContext(r.Context(), err, "SendRecoveryEmail", map[string]interface{}{
			"email":  request.Email,
			"action": "check_cool_down",
		})
		_, _ = api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrInternalServerError.Error())
		return
	}

	if !result.Allowed {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", result.WaitS))
		_, _ = api.RespondError(w, http.StatusTooManyRequests,
			fmt.Sprintf("Too many requests. Wait %d seconds", result.WaitS))
		return
	}

	userLink, err := ms.geterUserLink.GetUserLink(r.Context(), request.Email)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentEmail) || errors.Is(err, common.ErrorNonexistentUser) {
			_, _ = api.Respond(w, http.StatusOK, api.StatusOK)
			return
		}
		errLog := fmt.Errorf("GetUserLink: %w", err)
		logger.Err(errLog).Msg("send recovery code")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "SendRecoveryEmail", map[string]interface{}{
			"email":  request.Email,
			"action": "get_user_link",
		})
		_, _ = api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrInternalServerError.Error())
		return
	}

	if err := ms.mailSender.SendRecoveryCode(r.Context(), userLink, request.Email); err != nil {
		if errors.Is(err, common.ErrorNonexistentEmail) || errors.Is(err, common.ErrorNonexistentUser) {
			_, _ = api.Respond(w, http.StatusOK, api.StatusOK)
			return
		}
		errLog := fmt.Errorf("auth.SendRecoveryCode: %w", err)
		logger.Err(errLog).Msg("send recovery code")
		sentryLogger.CaptureFromContext(r.Context(), errLog, "SendRecoveryEmail", map[string]interface{}{
			"email":  request.Email,
			"action": "send_recovery_code",
		})
		_, _ = api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrCannotSendRecoveryCode.Error())
		return
	}

	_, _ = api.Respond(w, http.StatusOK, api.StatusOK)
}

// CheckRecoveryCode проверяет отправленный на почту код
//
//	@Summary	Проверить код восстановления
//	@Tags		Auth
//	@Accept		json
//	@Produce	json
//	@Param		input	body		dto.RecoveryCodeRequest	true	"Код из письма"
//	@Success	200		{object}	api.Response			"OK"
//	@Failure	400		{object}	api.ErrorResponse		"Invalid request schema"
//	@Failure	500		{object}	api.ErrorResponse		"Internal server error"
//	@Router		/check-code [post]
func (ms *MailSender) CheckRecoveryCode(w http.ResponseWriter, r *http.Request) {
	logger := zerolog.Ctx(r.Context())

	var request dto.RecoveryCodeRequest
	if err := easyjson.UnmarshalFromReader(r.Body, &request); err != nil {
		_, _ = api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	if err := ms.mailSender.CheckRecoveryCode(r.Context(), request.Code); err != nil {
		errLog := fmt.Errorf("mailSender.CheckRecoveryCode: %w", err)
		logger.Error().Err(errLog).Msg("auth.CheckRecoveryCode failed")
		if errors.Is(err, common.ErrorResetTokenNotFound) {
			_, _ = api.RespondError(w, http.StatusBadRequest, common.ErrorResetTokenNotFound.Error())
			return
		}
		sentryLogger.CaptureFromContext(r.Context(), errLog, "CheckRecoveryCode", map[string]interface{}{
			"action": "check_recovery_code",
		})
		_, _ = api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrInternalServerError.Error())
		return
	}

	_, _ = api.Respond(w, http.StatusOK, api.StatusOK)
}
