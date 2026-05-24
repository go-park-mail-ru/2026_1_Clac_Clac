package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/pubsub"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	// TODO: перенести в конфиг
	LongPollingTimeout   = 30
	keepAliveInterval    = 15
	streamingUnsupported = "streaming unsupported"
	subscribeFailed      = "subscribe failed"
	internalServerError  = "internal server error"
	contentTypeHeader    = "Content-Type"
	contentTypeSSE       = "text/event-stream"
	cacheControlHeader   = "Cache-Control"
	cacheControlSSE      = "no-cache"
	connectionHeader     = "Connection"
	connectionSSE        = "keep-alive"
)

type RealtimeService interface {
	Subscribe(ctx context.Context, boardLink uuid.UUID) (pubsub.Subscription[common.BoardUpdateEvent], error)
}

type RealtimeHandler struct {
	service RealtimeService
}

func NewRealtimeHandler(service RealtimeService) *RealtimeHandler {
	return &RealtimeHandler{
		service: service,
	}
}

func (h *RealtimeHandler) EventsSSE(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := zerolog.Ctx(ctx)

	boardLink, ok := ctx.Value(middleware.BoardContextLink{}).(uuid.UUID)
	if !ok {
		api.RespondError(w, http.StatusBadRequest, common.ErrBoardLinkMissing.Error())
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		api.RespondError(w, http.StatusInternalServerError, streamingUnsupported)
		return
	}

	w.Header().Set(contentTypeHeader, contentTypeSSE)
	w.Header().Set(cacheControlHeader, cacheControlSSE)
	w.Header().Set(connectionHeader, connectionSSE)

	sub, err := h.service.Subscribe(ctx, boardLink)
	if err != nil {
		api.RespondError(w, http.StatusInternalServerError, subscribeFailed)
		return
	}
	defer sub.Close()

	keepAliveTicker := time.NewTicker(keepAliveInterval * time.Second)
	defer keepAliveTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-sub.C():
			if !ok {
				if err := sub.Err(); err != nil {
					logger.Error().Err(err).Msg("read event")
				}
				continue
			}

			data, err := json.Marshal(event)
			if err != nil {
				logger.Error().Err(err).Msg("json.Marshal")
				api.RespondError(w, http.StatusInternalServerError, internalServerError)
				return
			}

			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-keepAliveTicker.C:
			fmt.Fprint(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}
