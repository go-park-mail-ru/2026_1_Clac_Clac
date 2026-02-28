package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"time"

	common "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth"
)

var (
	ErrorLenPassword     = errors.New("password must contain minimum 6")
	ErrorIncorrectEmail  = errors.New("invalid email format")
	ErrorIncorrectSymbol = errors.New("allowed only a-z, A-Z, 0-9, and /?!@")
)

type ErrorResponce struct {
	Error string `json:"error"`
}

func MakeJSONError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)

	errorResponce := ErrorResponce{
		Error: err.Error(),
	}

	if err = json.NewEncoder(w).Encode(errorResponce); err != nil {
		fmt.Printf("error encoding response: %v\n", err)
	}
}

type Request struct {
	Name     string `json:"name"`
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

func CheckEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (h *RegisterHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var request Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		MakeJSONError(w, http.StatusBadRequest, fmt.Errorf("%w: %w", common.ErrorDecodeRequest, err))
		return
	}

	if isAsciiSymbol := CheckAsciiSymbol(request.Password, request.Email); !isAsciiSymbol {
		MakeJSONError(w, http.StatusBadRequest, ErrorIncorrectSymbol)
		return
	}

	if len(request.Password) < 6 {
		MakeJSONError(w, http.StatusBadRequest, ErrorLenPassword)
		return
	}

	if correctEmail := CheckEmail(request.Email); !correctEmail {
		MakeJSONError(w, http.StatusBadRequest, ErrorIncorrectEmail)
		return
	}

	user, sessionID, err := h.srv.Register(r.Context(), request.Name, request.Password, request.Email)
	if err != nil {
		MakeJSONError(w, http.StatusBadRequest, fmt.Errorf("cannot register user: %w", err))
		return
	}

	response := Response{
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
