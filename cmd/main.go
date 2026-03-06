package main

import (
	"net/http"
	"os"

	handlers "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handlers/auth"
	repository "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/auth"
	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth"
	"github.com/rs/zerolog"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	authRepository := repository.NewMapDB()
	authService := service.NewAuthService(authRepository, service.HashPassword, service.CheckPassword, service.GenerateSessionID)
	authHandler := handlers.NewAuthHandler(authService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", authHandler.RegisterUser)
	mux.HandleFunc("POST /login", authHandler.LogInUser)
	mux.HandleFunc("POST /logout", authHandler.LogOutUser)

	// authWare := middleware.AuthMiddleware(authRepository)

	err := http.ListenAndServe(":8081", mux)
	if err != nil {
		logger.Fatal().
			Err(err).
			Str("status", "failed").
			Msg("cannot created server")
	}
}
