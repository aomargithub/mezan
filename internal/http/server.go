package http

import (
	"net/http"
	"log/slog"
	"runtime/debug"
	"database/sql"
	"github.com/aomargithub/mezan/internal/db"
)

type Server struct {
	Mux *http.ServeMux
	MezanaRepo db.MezanaRepo
	Addr      string
    StaticDir string
	Logger *slog.Logger
	Db *sql.DB
	mezanaRepo db.MezanaRepo
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
	s.mezanaRepo =  db.MezanaRepo{
		DB: s.Db,
	}
}

func (s *Server) RegisterHandlers() {
	s.initRepos()
	fileServer := http.FileServer(http.Dir(s.StaticDir))

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /{$}", s.createMezanHandler())

	s.Mux = mux
}