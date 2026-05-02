package router

import (
	"context"
	"net/http"

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

type CardHandler interface {
	GetCard(w http.ResponseWriter, r *http.Request)
	DeleteCard(w http.ResponseWriter, r *http.Request)
	UpdateCard(w http.ResponseWriter, r *http.Request)
	ReorderCards(w http.ResponseWriter, r *http.Request)
	CreateCard(w http.ResponseWriter, r *http.Request)
	GetComments(w http.ResponseWriter, r *http.Request)
	CreateComment(w http.ResponseWriter, r *http.Request)
	DeleteComment(w http.ResponseWriter, r *http.Request)
	UpdateComment(w http.ResponseWriter, r *http.Request)
	CreateSubtask(w http.ResponseWriter, r *http.Request)
	UpdateSubtask(w http.ResponseWriter, r *http.Request)
	DeleteSubtask(w http.ResponseWriter, r *http.Request)
}
type BoardHandler interface {
	GetBoards(w http.ResponseWriter, r *http.Request)
	GetBoard(w http.ResponseWriter, r *http.Request)
	CreateBoard(w http.ResponseWriter, r *http.Request)
	DeleteBoard(w http.ResponseWriter, r *http.Request)
	UpdateBoard(w http.ResponseWriter, r *http.Request)
	UploadBackground(w http.ResponseWriter, r *http.Request)
	GetMembers(w http.ResponseWriter, r *http.Request)
}

type SectionHandler interface {
	GetSections(w http.ResponseWriter, r *http.Request)
	GetSection(w http.ResponseWriter, r *http.Request)
	GetCards(w http.ResponseWriter, r *http.Request)
	CreateSection(w http.ResponseWriter, r *http.Request)
	DeleteSection(w http.ResponseWriter, r *http.Request)
	ReorderSections(w http.ResponseWriter, r *http.Request)
	UpdateSection(w http.ResponseWriter, r *http.Request)
}
type AppealHandler interface {
	CreateAppeal(w http.ResponseWriter, r *http.Request)
	GetAppeals(w http.ResponseWriter, r *http.Request)
	UploadAttachment(w http.ResponseWriter, r *http.Request)
	DeleteAppeal(w http.ResponseWriter, r *http.Request)
	GetStats(w http.ResponseWriter, r *http.Request)
	ChangeAppealStatus(w http.ResponseWriter, r *http.Request)
}

type Tools struct {
	Auth        AuthHandler
	Profile     ProfileHandler
	MailSender  MailSenderHandler
	CSRF        CSRFHandler
	Card        CardHandler
	AuthChecker middleware.SessionCheker
	RateLimiter middleware.CheckLimit
	CSRFChecker func(ctx context.Context, sessionID, token string) error
	Board       BoardHandler
	Section     SectionHandler
	Appeal      AppealHandler
}

func NewRouter(deps Tools, conf *config.Config, logger *zerolog.Logger) *mux.Router {
	r := mux.NewRouter().PathPrefix("/api").Subrouter()

	r.Use(middleware.SentryHubMiddleware())
	r.Use(middleware.RecoveryMiddleware(logger))
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.CORSMiddleware(&conf.CORS))
	r.Use(middleware.TimeOutMiddleware(conf.App.RequestTimeout))

	textLimit := middleware.LimitRequestSizeMiddleware(conf.App.MaxTextRequestSize)
	imageLimit := middleware.LimitRequestSizeMiddleware(conf.App.MaxUploadImageSize)

	r.HandleFunc("/healthcheck", healthcheck).Methods(http.MethodGet)

	loginRateConf := conf.Services.RateLimiters.GetParameters(config.LogInUser)
	registerRateConf := conf.Services.RateLimiters.GetParameters(config.RegisterUser)

	loginRateMW := middleware.RateLimiterMiddleware(deps.RateLimiter, domain.RateLimitConfig{
		Limit:   loginRateConf.Limit,
		Action:  loginRateConf.Action,
		WindowS: int64(loginRateConf.Window.Seconds()),
		TTL:     loginRateConf.TTL,
	}, logger)

	registerRateMW := middleware.RateLimiterMiddleware(deps.RateLimiter, domain.RateLimitConfig{
		Limit:   registerRateConf.Limit,
		Action:  registerRateConf.Action,
		WindowS: int64(registerRateConf.Window.Seconds()),
		TTL:     registerRateConf.TTL,
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

	protected := r.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthMiddleware(deps.AuthChecker, logger, conf.Services.Auth.Handler.SessionLifetime))

	protected.HandleFunc("/csrf", deps.CSRF.SetCSRFCookieHandler).Methods(http.MethodGet)

	csrfProtected := protected.PathPrefix("/").Subrouter()
	csrfProtected.Use(middleware.CSRFMiddleware(deps.CSRFChecker))

	withTextLimit := csrfProtected.PathPrefix("/").Subrouter()
	withTextLimit.Use(textLimit)

	withImageLimit := csrfProtected.PathPrefix("/").Subrouter()
	withImageLimit.Use(imageLimit)

	withTextLimit.HandleFunc("/me", deps.Auth.MeHandler).Methods(http.MethodGet)

	withTextLimit.HandleFunc("/profiles", deps.Profile.GetProfile).Methods(http.MethodGet)
	withTextLimit.HandleFunc("/profiles/{user_link}", deps.Profile.GetProfileByLink).Methods(http.MethodGet)
	withTextLimit.HandleFunc("/profiles/info", deps.Profile.UpdateProfile).Methods(http.MethodPost)
	withImageLimit.HandleFunc("/profiles/avatar", deps.Profile.UpdateAvatar).Methods(http.MethodPut)
	withTextLimit.HandleFunc("/profiles/avatar", deps.Profile.DeleteAvatar).Methods(http.MethodDelete)

	withTextLimit.HandleFunc("/cards", deps.Card.CreateCard).Methods(http.MethodPost)
	withTextLimit.HandleFunc("/cards/{link}", deps.Card.GetCard).Methods(http.MethodGet)
	withTextLimit.HandleFunc("/cards/{link}", deps.Card.DeleteCard).Methods(http.MethodDelete)
	withTextLimit.HandleFunc("/cards/{link}", deps.Card.UpdateCard).Methods(http.MethodPut)
	withTextLimit.HandleFunc("/cards/{link}/reorder", deps.Card.ReorderCards).Methods(http.MethodPatch)

	withTextLimit.HandleFunc("/cards/{link}/comments", deps.Card.GetComments).Methods(http.MethodGet)
	withTextLimit.HandleFunc("/cards/{link}/comments", deps.Card.CreateComment).Methods(http.MethodPost)
	withTextLimit.HandleFunc("/comments/{comment_link}", deps.Card.DeleteComment).Methods(http.MethodDelete)
	withTextLimit.HandleFunc("/comments/{comment_link}", deps.Card.UpdateComment).Methods(http.MethodPut)

	withTextLimit.HandleFunc("/cards/{link}/subtasks", deps.Card.CreateSubtask).Methods(http.MethodPost)
	withTextLimit.HandleFunc("/subtasks/{subtask_link}", deps.Card.UpdateSubtask).Methods(http.MethodPut)
	withTextLimit.HandleFunc("/subtasks/{subtask_link}", deps.Card.DeleteSubtask).Methods(http.MethodDelete)

	withTextLimit.HandleFunc("/sections", deps.Section.CreateSection).Methods(http.MethodPost)
	withTextLimit.HandleFunc("/sections/{link}", deps.Section.GetSection).Methods(http.MethodGet)
	withTextLimit.HandleFunc("/sections/{link}", deps.Section.DeleteSection).Methods(http.MethodDelete)
	withTextLimit.HandleFunc("/sections/{link}", deps.Section.UpdateSection).Methods(http.MethodPut)
	withTextLimit.HandleFunc("/sections/{link}/cards", deps.Section.GetCards).Methods(http.MethodGet)

	withTextLimit.HandleFunc("/boards", deps.Board.GetBoards).Methods(http.MethodGet)
	withTextLimit.HandleFunc("/boards", deps.Board.CreateBoard).Methods(http.MethodPost)
	withTextLimit.HandleFunc("/boards/{link}", deps.Board.GetBoard).Methods(http.MethodGet)
	withTextLimit.HandleFunc("/boards/{link}", deps.Board.DeleteBoard).Methods(http.MethodDelete)
	withTextLimit.HandleFunc("/boards/{link}", deps.Board.UpdateBoard).Methods(http.MethodPut)
	withImageLimit.HandleFunc("/boards/{link}/background", deps.Board.UploadBackground).Methods(http.MethodPut)
	withTextLimit.HandleFunc("/boards/{link}/users", deps.Board.GetMembers).Methods(http.MethodGet)
	withTextLimit.HandleFunc("/boards/{board_link}/sections", deps.Section.GetSections).Methods(http.MethodGet)
	withTextLimit.HandleFunc("/boards/{board_link}/sections/reorder", deps.Section.ReorderSections).Methods(http.MethodPatch)

	withTextLimit.HandleFunc("/appeals", deps.Appeal.CreateAppeal).Methods(http.MethodPost)
	withTextLimit.HandleFunc("/appeals", deps.Appeal.GetAppeals).Methods(http.MethodGet)
	withImageLimit.HandleFunc("/appeals/{link}/attachment", deps.Appeal.UploadAttachment).Methods(http.MethodPut)
	withTextLimit.HandleFunc("/appeal/{link}", deps.Appeal.DeleteAppeal).Methods(http.MethodDelete)
	withTextLimit.HandleFunc("/appeals/stats", deps.Appeal.GetStats).Methods(http.MethodGet)
	withTextLimit.HandleFunc("/appeals/{link}", deps.Appeal.ChangeAppealStatus).Methods(http.MethodPatch)

	return r
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	api.HandleError(api.Respond(w, http.StatusOK, api.StatusOK))
}
