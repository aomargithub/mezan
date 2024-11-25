package http

import (
	"database/sql"
	"fmt"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/aomargithub/mezan/internal/db"
	_ "github.com/jackc/pgx/v5/stdlib"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

type Server struct {
	Mux                      http.Handler
	StaticDir                string
	Logger                   *slog.Logger
	DSN                      string
	DB                       *sql.DB
	sessionManager           *scs.SessionManager
	mezaniService            db.MezaniService
	userService              db.UserService
	templates                map[string]*template.Template
	authenticatedUserIdKey   string
	authenticatedUserNameKey string
}

type Authentication struct {
	Name string
	Id   int
}

func (s Server) createAuthentication(r *http.Request) *Authentication {
	if s.sessionManager.Exists(r.Context(), s.authenticatedUserIdKey) {
		id := s.sessionManager.GetInt(r.Context(), s.authenticatedUserIdKey)
		name := s.sessionManager.GetString(r.Context(), s.authenticatedUserNameKey)
		return &Authentication{
			Name: name,
			Id:   id,
		}
	}
	return nil
}

func (a Authentication) isAuthenticated() bool {
	return strings.TrimSpace(a.Name) != "" || a.Id == 0
}

func (s Server) serverError(w http.ResponseWriter, r *http.Request, err error) {
	trace := string(debug.Stack())
	s.Logger.Error(err.Error(), "uri", r.URL.RequestURI(), "method", r.Method, "trace", trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (s Server) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (s *Server) initServices() {
	s.mezaniService = db.MezaniService{
		DB: s.DB,
	}

	s.userService = db.UserService{
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

	mux.Handle("GET /mezanis/{id}", s.sessionManager.LoadAndSave(s.getMezaniHandler()))
	mux.Handle("GET /mezanis/create", s.sessionManager.LoadAndSave(s.getMezaniCreateHandler()))
	mux.Handle("POST /mezanis/create", s.sessionManager.LoadAndSave(s.postMezaniCreateHandler()))

	mux.Handle("GET /users/signup", s.sessionManager.LoadAndSave(s.getUserSignUpHandler()))
	mux.Handle("POST /users/signup", s.sessionManager.LoadAndSave(s.postUserSignUpHandler()))

	mux.Handle("GET /login", s.sessionManager.LoadAndSave(s.getLoginHandler()))
	mux.Handle("POST /login", s.sessionManager.LoadAndSave(s.postLoginHandler()))

	mux.Handle("POST /logout", s.sessionManager.LoadAndSave(s.postLogoutHandler()))

	mux.Handle("GET /{$}", s.sessionManager.LoadAndSave(s.homeHandler()))

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
	db, err := sql.Open("pgx", s.DSN)
	if err != nil {
		s.Logger.Error(err.Error())
		os.Exit(1)
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		s.Logger.Error(err.Error())
		os.Exit(1)
	}
	s.DB = db
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
	s.authenticatedUserIdKey = "authenticatedUserID"
	s.authenticatedUserNameKey = "authenticatedUserName"
}
