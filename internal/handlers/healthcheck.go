package handlers

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
)

// HealthcheckHandler ручка для проверки состояния сервера
//
//	@Summary	Проверка состояния сервера
//	@Tags		Backend
//	@Produce	json
//	@Success	200	{object}	api.Response
//	@Router		/healthcheck [get]
func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}
