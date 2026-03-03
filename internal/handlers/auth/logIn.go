package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	common "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth"
)

func NewLogInHandler(srv service.LogInService) *LogInHandler {
	return &LogInHandler{
		srv: srv,
	}
}

type LogInHandler struct {
	srv service.LogInService
}

type LogInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogInResponse struct {
	Message string      `json:"message"`
	Profile models.User `json:"profile"`
}

func (l *LogInHandler) LogInUser(w http.ResponseWriter, r *http.Request) {
	var request LogInRequest
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

	user, sessionID, err := l.srv.Login(r.Context(), request.Email, request.Password)
	if err != nil {
		if errors.Is(err, service.ErrorWrongPassword) {
			common.MakeJSONError(w, http.StatusUnauthorized, errors.New("wrong email or password"))
			return
		}

		common.MakeJSONError(w, http.StatusBadRequest, fmt.Errorf("cannot log in user: %w", err))
		return
	}

	response := LogInResponse{
		Message: "user was successfully logged in",
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

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("error encoding response: %v\n", err)
	}
}
