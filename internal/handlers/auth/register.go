package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	common "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth"
)

type RegisterRequest struct {
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}

type RegisterResponse struct {
	Message string      `json:"message"`
	Profile models.User `json:"profile"`
}

type RegisterHandler struct {
	srv service.RegistrationService
}

func NewRegisterHandler(srv service.RegistrationService) *RegisterHandler {
	return &RegisterHandler{
		srv: srv,
	}
}

func (h *RegisterHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var request RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.MakeJSONError(w, http.StatusBadRequest, fmt.Errorf("%w: %w", common.ErrorDecodeRequest, err))
		return
	}

	if isAsciiSymbol := CheckAsciiSymbol(request.Password, request.Email); !isAsciiSymbol {
		common.MakeJSONError(w, http.StatusBadRequest, ErrorIncorrectSymbol)
		return
	}

	if len(request.Password) < 6 {
		common.MakeJSONError(w, http.StatusBadRequest, ErrorLenPassword)
		return
	}

	if correctEmail := CheckEmail(request.Email); !correctEmail {
		common.MakeJSONError(w, http.StatusBadRequest, ErrorIncorrectEmail)
		return
	}

	user, sessionID, err := h.srv.Register(r.Context(), request.DisplayName, request.Password, request.Email)
	if err != nil {
		common.MakeJSONError(w, http.StatusBadRequest, fmt.Errorf("cannot register user: %w", err))
		return
	}

	response := RegisterResponse{
		Message: "user was successsfully created",
		Profile: user,
	}

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	}

	http.SetCookie(w, cookie)

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("error encoding response: %v\n", err)
	}
}
