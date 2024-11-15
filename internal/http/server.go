package http

import (
	"fmt"
	"database/sql"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"runtime/debug"
	"github.com/aomargithub/mezan/internal/db"
)

type Server struct {
	Mux http.Handler
	Addr      string
    StaticDir string
	Logger *slog.Logger
	Db *sql.DB
	mezaniService db.MezaniService
	templates map[string]*template.Template
}


func (s Server) ServerError (w http.ResponseWriter, r *http.Request, err error) {
	trace := string(debug.Stack())
	s.Logger.Error(err.Error(), "uri", r.URL.RequestURI(), "method", r.Method, "trace", trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}


func (s Server) ClientError (w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (s *Server) initRepos() {
	s.mezaniService =  db.MezaniService{
		DB: s.Db,
	}
}

func (s *Server) initTemplateCache() {
	cache := map[string]*template.Template{}
	pages, _ := filepath.Glob("./ui/html/pages/*.tmpl")
	
	for _, page := range pages {

		name := filepath.Base(page)

		files := []string {
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
	s.initRepos()
	fileServer := http.FileServer(http.Dir(s.StaticDir))

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /mezanis/{id}", s.createGetMezaniHandler())
	mux.HandleFunc("GET /mezanis/create", s.createGetMezaniCreateHandler())
	mux.HandleFunc("POST /mezanis/create", s.createMezaniCreateHandler())
	mux.HandleFunc("GET /{$}", s.createHomeHandler())

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
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		var (
			ip = r.RemoteAddr
			proto = r.Proto
			method = r.Method
			uri    = r.URL.RequestURI()
		)
		s.Logger.Info("received request", "ip", ip, "proto", proto, "method", method, "uri", uri)
		next.ServeHTTP(w, r)
	})
}

func (s Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request){
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				s.ServerError(w, r, fmt.Errorf("%s", err))
			} 
		}()
		next.ServeHTTP(w, r)
	} )
}

func (s *Server) Init() {
	s.initTemplateCache()
	s.initRepos()
	s.registerHandlers()
}