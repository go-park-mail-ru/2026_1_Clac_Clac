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

	registrationService := service.NewRegistrationService(authRepository, service.HashPassword, service.GenerateSessionID)
	registrationHandler := handlers.NewRegisterHandler(registrationService)

	logInService := service.NewLogInService(authRepository, service.CheckPassword, service.GenerateSessionID)
	logInHandler := handlers.NewLogInHandler(logInService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", registrationHandler.RegisterUser)
	mux.HandleFunc("POST /login", logInHandler.LogInUser)

	// authWare := middleware.AuthMiddleware(authRepository)

	err := http.ListenAndServe(":8081", mux)
	if err != nil {
		logger.Fatal().
			Err(err).
			Str("status", "failed").
			Msg("cannot created server")
	}
}
