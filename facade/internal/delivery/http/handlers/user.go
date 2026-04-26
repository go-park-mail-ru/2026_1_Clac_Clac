package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/usecase"
	user "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/contracts/user"
)

type UserHandler struct {
	uc     usecase.UserUsecase
	logger *zerolog.Logger
}

func NewUserHandler(uc usecase.UserUsecase, logger *zerolog.Logger) *UserHandler {
	return &UserHandler{uc: uc, logger: logger}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userLink := mux.Vars(r)["user_link"]

	profile, err := h.uc.GetProfile(r.Context(), userLink)
	if err != nil {
		h.logger.Err(err).Msg("GetProfile")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	sessionID, userLink, err := h.uc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Err(err).Msg("Login")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true,
		Path:     "/",
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"user_link": userLink})
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req user.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	sessionID, userLink, err := h.uc.Register(r.Context(), &req)
	if err != nil {
		h.logger.Err(err).Msg("Register")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true,
		Path:     "/",
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"user_link": userLink})
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.uc.Logout(r.Context(), cookie.Value); err != nil {
		h.logger.Err(err).Msg("Logout")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userLink, _ := r.Context().Value(middleware.UserLinkKey).(string)

	var req struct {
		DisplayName string `json:"display_name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	h.logger.Info().Str("user_link", userLink).Msg("UpdateProfile")
	w.WriteHeader(http.StatusNoContent)
}
