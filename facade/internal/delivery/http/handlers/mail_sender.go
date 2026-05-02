package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/dto"
	handlerCommon "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/google/uuid"
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
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	ok := ValidateEmail(request.Email)
	if !ok {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidEmailOrPassword.Error())
		return
	}

	result, err := ms.coolDown.CheckCoolDown(r.Context(), domain.Cooldown{
		Name:        nameCoolDown,
		Email:       request.Email,
		ExpirationS: ms.cfg.CoolDownExpirationSec,
	})
	if err != nil {
		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrInternalServerError.Error())
		return
	}

	if !result.Allowed {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", result.WaitS))
		api.RespondError(w, http.StatusTooManyRequests,
			fmt.Sprintf("Too many requests. Wait %d seconds", result.WaitS))
		return
	}

	userLink, err := ms.geterUserLink.GetUserLink(r.Context(), request.Email)
	if err != nil {
		if errors.Is(err, common.ErrorNonexistentEmail) || errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, handlerCommon.ErrUserDoesNotExists.Error())
			return
		}
		logger.Err(fmt.Errorf("GetUserLink: %w", err)).Msg("send recovery code")
		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrInternalServerError.Error())
		return
	}

	if err := ms.mailSender.SendRecoveryCode(r.Context(), userLink, request.Email); err != nil {
		if errors.Is(err, common.ErrorNonexistentEmail) || errors.Is(err, common.ErrorNonexistentUser) {
			api.RespondError(w, http.StatusNotFound, handlerCommon.ErrUserDoesNotExists.Error())
			return
		}
		logger.Err(fmt.Errorf("auth.SendRecoveryCode: %w", err)).Msg("send recovery code")
		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrCannotSendRecoveryCode.Error())
		return
	}

	api.Respond(w, http.StatusOK, api.StatusOK)
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
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.RespondError(w, http.StatusBadRequest, handlerCommon.ErrInvalidRequestSchema.Error())
		return
	}

	if err := ms.mailSender.CheckRecoveryCode(r.Context(), request.Code); err != nil {
		logger.Error().Err(err).Msg("auth.CheckRecoveryCode failed")
		if errors.Is(err, common.ErrorResetTokenNotFound) {
			api.RespondError(w, http.StatusBadRequest, common.ErrorResetTokenNotFound.Error())
			return
		}
		api.RespondError(w, http.StatusInternalServerError, handlerCommon.ErrInternalServerError.Error())
		return
	}

	api.Respond(w, http.StatusOK, api.StatusOK)
}
