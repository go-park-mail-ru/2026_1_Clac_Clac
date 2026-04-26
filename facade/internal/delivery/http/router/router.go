package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"

	grpcclient "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/clients/grpc"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/usecase"
)

func NewRouter(clients *grpcclient.Clients, _ *config.Config, logger *zerolog.Logger) http.Handler {
	r := mux.NewRouter()

	userUC := usecase.NewUserUsecase(clients.User, clients.Auth)
	boardUC := usecase.NewBoardUsecase(clients.Board)

	userH := handlers.NewUserHandler(userUC, logger)
	boardH := handlers.NewBoardHandler(boardUC, logger)

	authMW := middleware.NewAuthMiddleware(clients.Auth, logger)

	// Public routes
	public := r.PathPrefix("/api/v1").Subrouter()
	public.HandleFunc("/login", userH.Login).Methods(http.MethodPost)
	public.HandleFunc("/register", userH.Register).Methods(http.MethodPost)

	// Protected routes
	protected := r.PathPrefix("/api/v1").Subrouter()
	protected.Use(authMW.Handle)
	protected.HandleFunc("/logout", userH.Logout).Methods(http.MethodPost)
	protected.HandleFunc("/profile/{user_link}", userH.GetProfile).Methods(http.MethodGet)
	protected.HandleFunc("/profile", userH.UpdateProfile).Methods(http.MethodPut)
	protected.HandleFunc("/boards", boardH.GetBoards).Methods(http.MethodGet)
	protected.HandleFunc("/boards/{board_id}", boardH.GetBoard).Methods(http.MethodGet)

	return r
}
