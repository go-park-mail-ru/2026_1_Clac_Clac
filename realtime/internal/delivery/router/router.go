package delivery

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type RealtimeHandler interface {
	EventsLongPolling(w http.ResponseWriter, r *http.Request)
}

type Tools struct {
	Realtime     RealtimeHandler
	AuthChecker  middleware.SessionChecker
	BoardChecker middleware.BoardChecker
}

func NewRouter(deps Tools, conf *config.Config, logger *zerolog.Logger) *mux.Router {
	mainRouter := mux.NewRouter()
	subRouter := mainRouter.PathPrefix("/api/events").Subrouter()

	subRouter.Use(middleware.RecoveryMiddleware(logger))
	subRouter.Use(middleware.LoggerMiddleware(logger))
	subRouter.Use(middleware.AuthMiddleware(deps.AuthChecker, logger, conf.Services.Auth.SessionLifetime))
	subRouter.Use(middleware.BoardAccessMiddleware(deps.BoardChecker))

	subRouter.HandleFunc("/{board_link}", deps.Realtime.EventsLongPolling).Methods(http.MethodGet)

	return mainRouter
}
