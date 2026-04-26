package router

import (
	"context"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/middleware"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type RouterDeps struct {
	Auth        *handlers.Auth
	Profile     *handlers.Profile
	AuthChecker middleware.SessionCheker
	RateLimiter middleware.CheckLimit
	CSRFChecker func(ctx context.Context, sessionID, token string) error
}

func NewRouter(deps RouterDeps, conf *config.Config, logger *zerolog.Logger) *mux.Router {
	r := mux.NewRouter().PathPrefix("/api").Subrouter()

	r.Use(middleware.RecoveryMiddleware(logger))
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.CORSMiddleware(&conf.CORS))
	r.Use(middleware.TimeOutMiddleware(5 * time.Second))

	textLimit := middleware.LimitRequestSizeMiddleware(conf.App.MaxTextRequestSize)
	imageLimit := middleware.LimitRequestSizeMiddleware(conf.App.MaxUploadImageSize)

	r.HandleFunc("/csrf", deps.Auth.SetCSRFCookieHandler).Methods(http.MethodGet)
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
	public.HandleFunc("/forgot-password", deps.Auth.SendRecoveryEmail).Methods(http.MethodPost)
	public.HandleFunc("/check-code", deps.Auth.CheckRecoveryCode).Methods(http.MethodPost)
	public.HandleFunc("/reset-password", deps.Auth.ResetUserPassword).Methods(http.MethodPost)

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

	return r
}

func healthcheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
