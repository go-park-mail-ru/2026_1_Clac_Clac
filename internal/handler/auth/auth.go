package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	common "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	models "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	service "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth"
	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, name, password, email string) (models.User, string, error)
	LogIn(ctx context.Context, email, userID string) (models.User, string, error)
	LogOut(ctx context.Context, sessionID string) error
	GetUserID(ctx context.Context, sessionID string) (uuid.UUID, error)
	DiliveryCodeReseting(ctx context.Context, email string) error
	CheckCode(ctx context.Context, tokenID string) error
	ResetPassword(ctx context.Context, tokenID, newPassword string) error
}

func NewAuthHandler(srv AuthService) *AuthHandler {
	return &AuthHandler{
		srv: srv,
	}
}

type AuthHandler struct {
	srv AuthService
}

type LogInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogInResponse struct {
	Message string      `json:"message"`
	Profile models.User `json:"profile"`
}

func (a *AuthHandler) LogInUser(w http.ResponseWriter, r *http.Request) {
	var request LogInRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.MakeJSONError(w, http.StatusBadRequest, fmt.Errorf("%w: %w", common.ErrorDecodeRequest, err))
		return
	}

	err := ValidatorRequestAuth(request.Email, request.Password)
	if err != nil {
		common.MakeJSONError(w, http.StatusBadRequest, fmt.Errorf("ValidatorRequestAuth: %w", err))
		return
	}

	user, sessionID, err := a.srv.LogIn(r.Context(), request.Email, request.Password)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type RegisterRequest struct {
	DisplayName      string `json:"display_name"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	RepeatedPassword string `json:"repeated_password"`
}

type RegisterResponse struct {
	Message string      `json:"message"`
	Profile models.User `json:"profile"`
}

func (a *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var request RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.MakeJSONError(w, http.StatusBadRequest, fmt.Errorf("%w: %w", common.ErrorDecodeRequest, err))
		return
	}

	err := ValidatorWithCheckPassword(request.Email, request.Password, request.RepeatedPassword)
	if err != nil {
		common.MakeJSONError(w, http.StatusBadRequest, fmt.Errorf("ValidatorRequestAuth: %w", err))
		return
	}

	user, sessionID, err := a.srv.Register(r.Context(), request.DisplayName, request.Password, request.Email)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *AuthHandler) LogOutUser(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		common.MakeJSONError(w, http.StatusUnauthorized, fmt.Errorf("user not authorized"))
		return
	}

	sessionId := cookie.Value

	err = a.srv.LogOut(r.Context(), sessionId)
	if err != nil {
		common.MakeJSONError(w, http.StatusInternalServerError, fmt.Errorf("failed to logout: %w", err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{message: "successfully logged out"}`))
}

type DiliveryRequest struct {
	Email string `json:"email"`
}

func (a *AuthHandler) DiliveryLetter(w http.ResponseWriter, r *http.Request) {
	var request DiliveryRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		common.MakeJSONError(w, http.StatusBadRequest, fmt.Errorf("%w: %w", common.ErrorDecodeRequest, err))
		return
	}

	err = a.srv.DiliveryCodeReseting(r.Context(), request.Email)
	if err != nil {
		common.MakeJSONError(w, http.StatusBadRequest, fmt.Errorf("can not send letter: %w", err))
		return
	}
}

type CodeRequest struct {
	Code string `json:"email"`
}

func (a *AuthHandler) CheckCodeLetter(w http.ResponseWriter, r *http.Request) {
	var request CodeRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		common.MakeJSONError(w, http.StatusBadRequest, fmt.Errorf("%w: %w", common.ErrorDecodeRequest, err))
		return
	}

	err = a.srv.CheckCode(r.Context(), request.Code)
	if err != nil {
		common.MakeJSONError(w, http.StatusBadRequest, fmt.Errorf("can not send letter: %w", err))
		return
	}
}

type NewPasswordRequest struct {
	tokenID          string `jsin:"token_id"`
	Password         string `json:"password"`
	RepeatedPassword string `json:"repeated_password"`
}

func (a *AuthHandler) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	var request NewPasswordRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		common.MakeJSONError(w, http.StatusBadRequest, fmt.Errorf("%w: %w", common.ErrorDecodeRequest, err))
		return
	}

	err = ValidatorRequestNewPassword(request.Password, request.RepeatedPassword)
	if err != nil {
		common.MakeJSONError(w, http.StatusBadRequest, fmt.Errorf("ValidatorRequestNewPassword: %w", err))
		return
	}

	err = a.srv.ResetPassword(r.Context(), request.tokenID, request.Password)
	if err != nil {
		common.MakeJSONError(w, http.StatusBadRequest, fmt.Errorf("can not reset pasdword: %w", err))
		return
	}
}
