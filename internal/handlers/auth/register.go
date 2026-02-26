package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	common "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth"
)

var (
	ErrorLenPassword      = errors.New("password must contain minimum 6")
	ErrorCountAtSignEmail = errors.New("must use only one @ in email")
	ErrorIncorrectSymbol  = errors.New("allowed only a-z, A-Z, 0-9, and /?!@")
)

type Request struct {
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type Response struct {
	Message string      `json:"message"`
	Profile models.User `json:"profile"`
}

type RegisterHandler struct {
	srv service.RegistrationService
}

func CreatedRegisterHandler(srv service.RegistrationService) *RegisterHandler {
	return &RegisterHandler{
		srv: srv,
	}
}

func CheckAsciiSymbol(strings ...string) bool {
	for _, str := range strings {
		for _, symbol := range str {
			if symbol > 127 {
				return false
			}
		}
	}

	return true
}

func (h *RegisterHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var request Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Errorf("%w: %w", common.ErrorDecodeRequest, err).Error(), http.StatusBadRequest)
		return
	}

	if isAsciiSymbol := CheckAsciiSymbol(request.Name, request.Surname, request.Password, request.Email); !isAsciiSymbol {
		http.Error(w, ErrorIncorrectSymbol.Error(), http.StatusBadRequest)
		return
	}

	if len(request.Password) < 6 {
		http.Error(w, ErrorLenPassword.Error(), http.StatusBadRequest)
		return
	}

	countAtSign := strings.Count(request.Email, "@")
	if countAtSign != 1 {
		http.Error(w, ErrorCountAtSignEmail.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.srv.Register(r.Context(), request.Name, request.Surname, request.Password, request.Email)
	if err != nil {
		http.Error(w, fmt.Errorf("cannot register user: %w", err).Error(), http.StatusBadRequest)
		return
	}

	response := Response{
		Message: "user was successsfully created",
		Profile: user,
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("error encoding response: %v\n", err)
	}
}
