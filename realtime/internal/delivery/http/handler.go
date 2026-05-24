package delivery

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/delivery/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/middleware"
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/usecase/dto"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	// TODO: перенести в конфиг
	LongPollingTimeout = 30
)

type RealtimeService interface {
	Listen(ctx context.Context, boardLink uuid.UUID) (serviceDto.BoardUpdateInfo, error)
}

type RealtimeHandler struct {
	service RealtimeService
}

func NewRealtimeHandler(service RealtimeService) *RealtimeHandler {
	return &RealtimeHandler{
		service: service,
	}
}

func (h *RealtimeHandler) EventsLongPolling(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger := zerolog.Ctx(ctx)

	boardLink, ok := ctx.Value(middleware.BoardContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusBadRequest, common.ErrBoardLinkMissing.Error())
		return
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, LongPollingTimeout*time.Second)
	defer cancel()

	boardUpdateInfo, err := h.service.Listen(ctxWithTimeout, boardLink)
	if err != nil {
		if errors.Is(err, common.ErrTimeout) {
			api.Respond(w, http.StatusNoContent, api.StatusTimeout)
			return
		}

		logger.Error().Err(err).Msg("RealtimeService.Listen")

		api.RespondError(w, http.StatusBadRequest, common.ErrCannotGetEvents.Error())
		return
	}

	api.HandleError(api.RespondOk(w, dto.BoardUpdateInfo{
		Type:    boardUpdateInfo.Type,
		Payload: boardUpdateInfo.Payload,
	}))
}
