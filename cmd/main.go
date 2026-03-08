package main

import (
	"net/http"
	"os"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	dbConnection "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/db_connection"

	authRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/auth"
	boardRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/board"
	profileRep "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository/profile"

	authServ "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth"
	boardServ "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/board"
	profileServ "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/profile"

	authHand "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handler/auth"
	boardHand "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handler/board"
	profileHand "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handler/profile"

	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth"

	"github.com/rs/zerolog"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	conectionDb := dbConnection.NewMapDatabse()

	authRepository := authRep.NewAuthRepository(conectionDb)
	boardRepository := boardRep.NewBoardRepository(conectionDb)
	profileRepository := profileRep.NewProfileRepository(conectionDb)

	authService := authServ.NewAuthService(authRepository, service.HashPassword, service.CheckPassword, service.GenerateSessionID)
	boardService := boardServ.NewBoardService(boardRepository)
	profileService := profileServ.NewProfileService(profileRepository)

	authHandler := authHand.NewAuthHandler(authService)
	boardHandler := boardHand.NewBoardHandler(boardService)
	profileHandler := profileHand.NewProfileHandler(profileService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", authHandler.RegisterUser)
	mux.HandleFunc("POST /login", authHandler.LogInUser)
	mux.HandleFunc("POST /logout", authHandler.LogOutUser)

	authWare := middleware.AuthMiddleware(authService)

	mux.Handle("GET /home", authWare(http.HandlerFunc(boardHandler.GetUserBoards)))
	mux.Handle("GET /profile", authWare(http.HandlerFunc(profileHandler.GetProfile)))

	err := http.ListenAndServe(":8081", mux)
	if err != nil {
		logger.Fatal().
			Err(err).
			Str("status", "failed").
			Msg("cannot created server")
	}
}
