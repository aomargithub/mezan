package http

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/aomargithub/mezan/internal/db"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/justinas/nosurf"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type contextKey string

const authenticatedUserIdSessionKey = "authenticatedUserIdSessionKey"
const authenticatedUserNameSessionKey = "authenticatedUserNameSessionKey"
const isAuthenticatedCtxKey = contextKey("isAuthenticated")

type Server struct {
	Mux                     http.Handler
	StaticDir               string
	Logger                  *slog.Logger
	DSN                     string
	DB                      *sql.DB
	sessionManager          *scs.SessionManager
	mezaniService           db.MezaniService
	mezaniShareService      db.MezaniShareService
	userService             db.UserService
	expenseService          db.ExpenseService
	expenseShareService     db.ExpenseShareService
	expenseItemService      db.ExpenseItemService
	expenseItemShareService db.ExpenseItemShareService
	membershipService       db.MembershipService
	templates               map[string]*template.Template
}

type Authentication struct {
	Name string
	Id   int
}

type CommonView struct {
	CsrfToken string
	*Authentication
	Flash string
}

type CommonCreateView struct {
	CommonView
	Validator
}

func (s Server) commonView(r *http.Request) CommonView {
	return CommonView{
		CsrfToken:      s.csrfToken(r),
		Authentication: s.createAuthentication(r),
		Flash:          s.sessionManager.PopString(r.Context(), "flash"),
	}
}

func (s Server) commonCreateView(r *http.Request) CommonCreateView {
	return CommonCreateView{
		CommonView: s.commonView(r),
	}
}

func (s Server) createAuthentication(r *http.Request) *Authentication {
	if s.isAuthenticated(r) {
		id := s.sessionManager.GetInt(r.Context(), authenticatedUserIdSessionKey)
		name := s.sessionManager.GetString(r.Context(), authenticatedUserNameSessionKey)
		return &Authentication{
			Name: name,
			Id:   id,
		}
	}
	return nil
}

func (s Server) csrfToken(r *http.Request) string {
	return nosurf.Token(r)
}

func (s *Server) initServices() {
	s.userService = db.UserService{
		DB: s.DB,
	}

	s.mezaniService = db.MezaniService{
		DB: s.DB,
	}

	s.mezaniShareService = db.MezaniShareService{
		DB: s.DB,
	}

	s.expenseService = db.ExpenseService{
		DB: s.DB,
	}

	s.expenseShareService = db.ExpenseShareService{
		DB: s.DB,
	}

	s.expenseItemService = db.ExpenseItemService{
		DB: s.DB,
	}

	s.expenseItemShareService = db.ExpenseItemShareService{
		DB: s.DB,
	}

	s.membershipService = db.MembershipService{
		DB: s.DB,
	}
}

func (s *Server) initTemplateCache() {
	cache := map[string]*template.Template{}
	pages, _ := filepath.Glob("./ui/html/pages/*.tmpl")

	for _, page := range pages {

		name := filepath.Base(page)

		files := []string{
			"./ui/html/base.tmpl",
			"./ui/html/partials/nav.tmpl",
			"./ui/html/partials/expenseItem.tmpl",
			page,
		}

		ts, _ := template.ParseFiles(files...)

		cache[name] = ts
	}
	s.templates = cache
}

func (s *Server) registerHandlers() {
	s.initServices()
	fileServer := http.FileServer(http.Dir(s.StaticDir))

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.Handle("GET /mezanis/{id}", s.sessionManager.LoadAndSave(s.noSurf(s.authenticate(s.requireAuthentication(s.getMezaniViewHandler())))))
	mux.Handle("GET /mezanis/shareId/{shareId}", s.sessionManager.LoadAndSave(s.noSurf(s.authenticate(s.requireAuthentication(s.getMezaniViewByShareIdHandler())))))
	mux.Handle("GET /mezanis/create", s.sessionManager.LoadAndSave(s.noSurf(s.authenticate(s.requireAuthentication(s.getMezaniCreateHandler())))))
	mux.Handle("POST /mezanis/create", s.sessionManager.LoadAndSave(s.noSurf(s.authenticate(s.requireAuthentication(s.mezaniCreateHandler())))))
	mux.Handle("GET /mezanis/{mezaniId}/expenses/create", s.sessionManager.LoadAndSave(s.noSurf(s.authenticate(s.requireAuthentication(s.getExpenseCreateHandler())))))
	mux.Handle("POST /mezanis/{mezaniId}/expenses/create", s.sessionManager.LoadAndSave(s.noSurf(s.authenticate(s.requireAuthentication(s.expenseCreateHandler())))))
	mux.Handle("GET /expenses/{expenseId}", s.sessionManager.LoadAndSave(s.noSurf(s.authenticate(s.requireAuthentication(s.getExpenseViewHandler())))))
	mux.Handle("GET /expenses/{expenseId}/items/create", s.sessionManager.LoadAndSave(s.noSurf(s.authenticate(s.requireAuthentication(s.getExpenseItemCreateHandler())))))
	mux.Handle("POST /expenses/{expenseId}/items/create", s.sessionManager.LoadAndSave(s.noSurf(s.authenticate(s.requireAuthentication(s.expenseItemCreateHandler())))))
	mux.Handle("GET /expenses/{expenseId}/participation", s.sessionManager.LoadAndSave(s.noSurf(s.authenticate(s.requireAuthentication(s.getExpenseParticipationCreateHandler())))))
	mux.Handle("POST /expenses/{expenseId}/participation", s.sessionManager.LoadAndSave(s.noSurf(s.authenticate(s.requireAuthentication(s.expenseParticipationCreateHandler())))))
	mux.Handle("GET /expenseItems/{expenseItemId}", s.sessionManager.LoadAndSave(s.noSurf(s.authenticate(s.requireAuthentication(s.getExpenseItemViewHandler())))))
	mux.Handle("PUT /expenseItems/{expenseItemId}", s.sessionManager.LoadAndSave(s.noSurf(s.authenticate(s.requireAuthentication(s.expenseItemUpdateHandler())))))
	mux.Handle("GET /expenseItems/{expenseItemId}/update", s.sessionManager.LoadAndSave(s.noSurf(s.authenticate(s.requireAuthentication(s.getExpenseItemUpdateHandler())))))
	mux.Handle("GET /expenseItems/{expenseItemId}/participation", s.sessionManager.LoadAndSave(s.noSurf(s.authenticate(s.requireAuthentication(s.getExpenseItemParticipationCreateHandler())))))
	mux.Handle("POST /expenseItems/{expenseItemId}/participation", s.sessionManager.LoadAndSave(s.noSurf(s.authenticate(s.requireAuthentication(s.expenseItemParticipationCreateHandler())))))

	mux.Handle("GET /users/signup", s.sessionManager.LoadAndSave(s.noSurf(s.getUserSignUpHandler())))
	mux.Handle("POST /users/signup", s.sessionManager.LoadAndSave(s.noSurf(s.postUserSignUpHandler())))

	mux.Handle("GET /login", s.sessionManager.LoadAndSave(s.noSurf(s.getLoginHandler())))
	mux.Handle("POST /login", s.sessionManager.LoadAndSave(s.noSurf(s.postLoginHandler())))

	mux.Handle("POST /logout", s.sessionManager.LoadAndSave(s.noSurf(s.postLogoutHandler())))

	mux.Handle("GET /{$}", s.sessionManager.LoadAndSave(s.noSurf(s.authenticate(s.requireAuthentication(s.homeHandler())))))

	s.Mux = s.recoverPanic(s.logRequest(commonHeaders(mux)))
}

func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		w.Header().Set("Server", "Go")

		next.ServeHTTP(w, r)
	})
}

func (s Server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ip     = r.RemoteAddr
			proto  = r.Proto
			method = r.Method
			uri    = r.URL.RequestURI()
		)
		s.Logger.Info("received request", "ip", ip, "proto", proto, "method", method, "uri", uri)
		next.ServeHTTP(w, r)
	})
}

func (s Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := s.sessionManager.GetInt(r.Context(), authenticatedUserIdSessionKey)

		if userId == 0 {
			next.ServeHTTP(w, r)
			return
		}

		exists, err := s.userService.Exists(userId)
		if err != nil {
			s.serverError(w, r, err)
			return
		}

		if exists {
			ctx := context.WithValue(r.Context(), isAuthenticatedCtxKey, true)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func (s Server) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.isAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Otherwise set the "Cache-Control: no-store" header so that pages
		// require commonFormData are not stored in the users browser cache (or
		// other intermediary cache).
		w.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func (s Server) noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHandler
}

func (s Server) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedCtxKey).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}

func (s Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				s.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s Server) render(w http.ResponseWriter, r *http.Request, page string, httpStatus int, templateData any) {
	ts, ok := s.templates[page]

	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		s.serverError(w, r, err)
	}
	w.WriteHeader(httpStatus)
	err := ts.ExecuteTemplate(w, "base", templateData)
	if err != nil {
		s.serverError(w, r, err)
	}
}

func (s *Server) initSessionManager() {
	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(s.DB)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true
	s.sessionManager = sessionManager
}

func (s *Server) initDB() {
	database, err := sql.Open("pgx", s.DSN)
	if err != nil {
		s.Logger.Error(err.Error())
		os.Exit(1)
	}

	err = database.Ping()
	if err != nil {
		_ = database.Close()
		s.Logger.Error(err.Error())
		os.Exit(1)
	}
	s.DB = database
}

func (s *Server) initLogger() {
	s.Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))
}

func (s *Server) Init() {
	s.initLogger()
	s.initDB()
	s.initTemplateCache()
	s.initServices()
	s.initSessionManager()
	s.registerHandlers()
}
