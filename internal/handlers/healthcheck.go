package handlers

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
)

// Обработчик healthcheck приложения
func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}
