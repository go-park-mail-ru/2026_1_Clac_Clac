package registration

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service"
)

var (
	ErrorDecodeRequest = errors.New("decoding request is incorrect")
)

type RequestParameters struct {
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type RegisterHandler struct {
	srv service.RegistrationService
}

func CreatedRegisterHandler(srv service.RegistrationService) *RegisterHandler {
	return &RegisterHandler{
		srv: srv,
	}
}

func (h *RegisterHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var request RequestParameters
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.srv.Register(r.Context(), request.Name, request.Surname, request.Password, request.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "user was successsfully created"})
}
