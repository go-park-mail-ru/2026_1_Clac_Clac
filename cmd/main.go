package main

import (
	"net/http"
	"os"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/engine"
	"github.com/rs/zerolog"
)

// Сделано в качетсве теста и примера
type ExampleUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func init() {
	// TODO: Добавить Debug и Production режимы
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func main() {
	// TODO: Прокидывать логгер в Engine через конфиг
	loggerOutput := zerolog.ConsoleWriter{Out: os.Stdout}
	logger := zerolog.New(loggerOutput).With().Timestamp().Logger()

	e := engine.New(&logger)

	// Сделано в качетсве теста и примера
	e.GET("/hello", func(w http.ResponseWriter, r *http.Request) {
		err := engine.RespondWithString(w, http.StatusOK, "Hello, World!")
		if err != nil {
			engine.GetLoggerFromRequest(r).Error().Err(err).Msg("/hello")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	// Сделано в качетсве теста и примера
	e.GET("/user", func(w http.ResponseWriter, r *http.Request) {
		user := ExampleUser{ID: 5, Name: "Vova"}
		logger := engine.GetLoggerFromRequest(r)

		logger.Info().Str("username", user.Name).Msg("Access")

		err := engine.RespondWithJSON(w, http.StatusOK, user)
		if err != nil {
			engine.GetLoggerFromRequest(r).Error().Err(err).Msg("/user")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	e.Start("localhost:8080")
}
