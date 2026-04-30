package router

import (
	"context"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type AuthHandler interface {
	MeHandler(w http.ResponseWriter, r *http.Request)
	LogInUser(w http.ResponseWriter, r *http.Request)
	RegisterUser(w http.ResponseWriter, r *http.Request)
	LogOutUser(w http.ResponseWriter, r *http.Request)
	VkOAuthCallback(w http.ResponseWriter, r *http.Request)
}

type ProfileHandler interface {
	GetProfile(w http.ResponseWriter, r *http.Request)
	GetProfileByLink(w http.ResponseWriter, r *http.Request)
	UpdateProfile(w http.ResponseWriter, r *http.Request)
	UpdateAvatar(w http.ResponseWriter, r *http.Request)
	DeleteAvatar(w http.ResponseWriter, r *http.Request)
	ResetUserPassword(w http.ResponseWriter, r *http.Request)
}

type MailSenderHandler interface {
	SendRecoveryEmail(w http.ResponseWriter, r *http.Request)
	CheckRecoveryCode(w http.ResponseWriter, r *http.Request)
}

type CSRFHandler interface {
	SetCSRFCookieHandler(w http.ResponseWriter, r *http.Request)
}

type Tools struct {
	Auth        AuthHandler
	Profile     ProfileHandler
	MailSender  MailSenderHandler
	CSRF        CSRFHandler
	AuthChecker middleware.SessionCheker
	RateLimiter middleware.CheckLimit
	CSRFChecker func(ctx context.Context, sessionID, token string) error
}

func NewRouter(deps Tools, conf *config.Config, logger *zerolog.Logger) *mux.Router {
	r := mux.NewRouter().PathPrefix("/api").Subrouter()

	r.Use(middleware.RecoveryMiddleware(logger))
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.CORSMiddleware(&conf.CORS))
	r.Use(middleware.TimeOutMiddleware(5 * time.Second))

	textLimit := middleware.LimitRequestSizeMiddleware(conf.App.MaxTextRequestSize)
	imageLimit := middleware.LimitRequestSizeMiddleware(conf.App.MaxUploadImageSize)

	r.HandleFunc("/csrf", deps.CSRF.SetCSRFCookieHandler).Methods(http.MethodGet)
	r.HandleFunc("/healthcheck", healthcheck).Methods(http.MethodGet)

	loginRateConf := conf.Services.RateLimiters.GetParameters(config.LogInUser)
	registerRateConf := conf.Services.RateLimiters.GetParameters(config.RegisterUser)

	loginRateMW := middleware.RateLimiterMiddleware(deps.RateLimiter, domain.RateLimitConfig{
		Limit:   loginRateConf.Limit,
		Action:  loginRateConf.Action,
		WindowS: int64(loginRateConf.Window.Seconds()),
	}, logger)

	registerRateMW := middleware.RateLimiterMiddleware(deps.RateLimiter, domain.RateLimitConfig{
		Limit:   registerRateConf.Limit,
		Action:  registerRateConf.Action,
		WindowS: int64(registerRateConf.Window.Seconds()),
	}, logger)

	public := r.PathPrefix("/").Subrouter()
	public.Use(textLimit)

	public.Handle("/login", loginRateMW(http.HandlerFunc(deps.Auth.LogInUser))).Methods(http.MethodPost)
	public.Handle("/register", registerRateMW(http.HandlerFunc(deps.Auth.RegisterUser))).Methods(http.MethodPost)
	public.HandleFunc("/logout", deps.Auth.LogOutUser).Methods(http.MethodPost)
	public.HandleFunc("/oauth/vk", deps.Auth.VkOAuthCallback)
	public.HandleFunc("/forgot-password", deps.MailSender.SendRecoveryEmail).Methods(http.MethodPost)
	public.HandleFunc("/check-code", deps.MailSender.CheckRecoveryCode).Methods(http.MethodPost)
	public.HandleFunc("/reset-password", deps.Profile.ResetUserPassword).Methods(http.MethodPost)

	csrfProtected := r.PathPrefix("/").Subrouter()
	csrfProtected.Use(middleware.CSRFMiddleware(deps.CSRFChecker))

	protected := csrfProtected.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthMiddleware(deps.AuthChecker, logger, conf.Services.Auth.Handler.SessionLifetime))

	withText := protected.PathPrefix("/").Subrouter()
	withText.Use(textLimit)

	withImage := protected.PathPrefix("/").Subrouter()
	withImage.Use(imageLimit)

	withText.HandleFunc("/me", deps.Auth.MeHandler).Methods(http.MethodGet)

	withText.HandleFunc("/profiles", deps.Profile.GetProfile).Methods(http.MethodGet)
	withText.HandleFunc("/profiles/{user_link}", deps.Profile.GetProfileByLink).Methods(http.MethodGet)
	withText.HandleFunc("/profiles/info", deps.Profile.UpdateProfile).Methods(http.MethodPost)
	withImage.HandleFunc("/profiles/avatar", deps.Profile.UpdateAvatar).Methods(http.MethodPut)
	withText.HandleFunc("/profiles/avatar", deps.Profile.DeleteAvatar).Methods(http.MethodDelete)

	withText.HandleFunc("/cards/{card_link}/subtasks", notImplemented).Methods(http.MethodPost)
	withText.HandleFunc("/cards/{card_link}/subtasks", notImplemented).Methods(http.MethodGet)
	withText.HandleFunc("/cards/{card_link}/subtasks/{link}", notImplemented).Methods(http.MethodPut)
	withText.HandleFunc("/cards/{card_link}/subtasks/{link}", notImplemented).Methods(http.MethodDelete)
	return r
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}

func notImplemented(w http.ResponseWriter, r *http.Request) {
	api.HandleError(api.RespondError(w, http.StatusNotImplemented, "not implemented"))
}
